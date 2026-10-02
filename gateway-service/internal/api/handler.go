package api

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/LeUrok/DS-lab2/gateway-service/internal/client"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	cars    *client.CarsClient
	rental  *client.RentalClient
	payment *client.PaymentClient
}

func NewHandler(cars *client.CarsClient, rental *client.RentalClient, payment *client.PaymentClient) *Handler {
	return &Handler{cars: cars, rental: rental, payment: payment}
}

const dateLayout = "2006-01-02"

type carResponse struct {
	CarUID             string `json:"carUid"`
	Brand              string `json:"brand"`
	Model              string `json:"model"`
	RegistrationNumber string `json:"registrationNumber"`
	Power              int    `json:"power"`
	Type               string `json:"type"`
	Price              int    `json:"price"`
	Available          bool   `json:"available"`
}

type paginationResponse struct {
	Page          int            `json:"page"`
	PageSize      int            `json:"pageSize"`
	TotalElements int            `json:"totalElements"`
	Items         []*carResponse `json:"items"`
}

type carInfo struct {
	CarUID             string `json:"carUid"`
	Brand              string `json:"brand"`
	Model              string `json:"model"`
	RegistrationNumber string `json:"registrationNumber"`
}

type paymentInfo struct {
	PaymentUID string `json:"paymentUid"`
	Status     string `json:"status"`
	Price      int    `json:"price"`
}

type rentalResponse struct {
	RentalUID string      `json:"rentalUid"`
	Status    string      `json:"status"`
	DateFrom  string      `json:"dateFrom"`
	DateTo    string      `json:"dateTo"`
	Car       carInfo     `json:"car"`
	Payment   paymentInfo `json:"payment"`
}

type createRentalRequest struct {
	CarUID   string `json:"carUid"`
	DateFrom string `json:"dateFrom"`
	DateTo   string `json:"dateTo"`
}

type createRentalResponse struct {
	RentalUID string      `json:"rentalUid"`
	Status    string      `json:"status"`
	CarUID    string      `json:"carUid"`
	DateFrom  string      `json:"dateFrom"`
	DateTo    string      `json:"dateTo"`
	Payment   paymentInfo `json:"payment"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetCars(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 0)
	size := parseIntQuery(r, "size", 10)
	showAll := r.URL.Query().Get("showAll") == "true"

	resp, err := h.cars.GetAll(r.Context(), page, size, showAll)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cars")
		return
	}

	items := make([]*carResponse, 0, len(resp.Items))
	for _, c := range resp.Items {
		items = append(items, &carResponse{
			CarUID:             c.CarUID,
			Brand:              c.Brand,
			Model:              c.Model,
			RegistrationNumber: c.RegistrationNumber,
			Power:              c.Power,
			Type:               c.Type,
			Price:              c.Price,
			Available:          c.Available,
		})
	}

	writeJSON(w, http.StatusOK, paginationResponse{
		Page:          resp.Page,
		PageSize:      resp.PageSize,
		TotalElements: resp.TotalElements,
		Items:         items,
	})
}

func (h *Handler) GetRentals(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")
	if username == "" {
		writeError(w, http.StatusBadRequest, "X-User-Name header is required")
		return
	}

	rentals, err := h.rental.GetByUsername(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get rentals")
		return
	}

	result := make([]rentalResponse, 0, len(rentals))
	for _, rental := range rentals {
		rr, err := h.enrichRental(r.Context(), rental)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to enrich rental")
			return
		}
		result = append(result, *rr)
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetRentalByUID(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "rentalUid")
	username := r.Header.Get("X-User-Name")
	if username == "" {
		writeError(w, http.StatusBadRequest, "X-User-Name header is required")
		return
	}

	rental, err := h.rental.GetByUID(r.Context(), uid, username)
	if errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusNotFound, "rental not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get rental")
		return
	}

	rr, err := h.enrichRental(r.Context(), rental)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enrich rental")
		return
	}
	writeJSON(w, http.StatusOK, rr)
}

func (h *Handler) CreateRental(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")
	if username == "" {
		writeError(w, http.StatusBadRequest, "X-User-Name header is required")
		return
	}

	var req createRentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dateFrom, err := time.Parse(dateLayout, req.DateFrom)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dateFrom")
		return
	}
	dateTo, err := time.Parse(dateLayout, req.DateTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dateTo")
		return
	}
	if !dateFrom.Before(dateTo) {
		writeError(w, http.StatusBadRequest, "dateFrom must be before dateTo")
		return
	}

	car, err := h.cars.GetByUID(r.Context(), req.CarUID)
	if errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get car")
		return
	}
	if !car.Available {
		writeError(w, http.StatusConflict, "car is not available")
		return
	}

	days := int(math.Ceil(dateTo.Sub(dateFrom).Hours() / 24))
	price := days * car.Price

	payment, err := h.payment.Create(r.Context(), price)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment")
		return
	}

	if err := h.cars.Reserve(r.Context(), req.CarUID); err != nil {
		_ = h.payment.Cancel(r.Context(), payment.PaymentUID)
		if errors.Is(err, client.ErrConflict) {
			writeError(w, http.StatusConflict, "car already reserved")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to reserve car")
		return
	}

	rental, err := h.rental.Create(r.Context(), client.CreateRentalRequest{
		Username:   username,
		CarUID:     req.CarUID,
		PaymentUID: payment.PaymentUID,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	})
	if err != nil {
		_ = h.cars.Unreserve(r.Context(), req.CarUID)
		_ = h.payment.Cancel(r.Context(), payment.PaymentUID)
		writeError(w, http.StatusInternalServerError, "failed to create rental")
		return
	}

	writeJSON(w, http.StatusOK, createRentalResponse{
		RentalUID: rental.RentalUID,
		Status:    rental.Status,
		CarUID:    rental.CarUID,
		DateFrom:  rental.DateFrom,
		DateTo:    rental.DateTo,
		Payment: paymentInfo{
			PaymentUID: payment.PaymentUID,
			Status:     payment.Status,
			Price:      payment.Price,
		},
	})
}

func (h *Handler) FinishRental(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "rentalUid")
	username := r.Header.Get("X-User-Name")
	if username == "" {
		writeError(w, http.StatusBadRequest, "X-User-Name header is required")
		return
	}

	rental, err := h.rental.GetByUID(r.Context(), uid, username)
	if errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusNotFound, "rental not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get rental")
		return
	}

	if err := h.rental.Finish(r.Context(), uid, username); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rental not found")
			return
		}
		if errors.Is(err, client.ErrConflict) {
			writeError(w, http.StatusConflict, "rental cannot be finished")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to finish rental")
		return
	}

	if err := h.cars.Unreserve(r.Context(), rental.CarUID); err != nil && !errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "failed to unreserve car")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CancelRental(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "rentalUid")
	username := r.Header.Get("X-User-Name")
	if username == "" {
		writeError(w, http.StatusBadRequest, "X-User-Name header is required")
		return
	}

	rental, err := h.rental.GetByUID(r.Context(), uid, username)
	if errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusNotFound, "rental not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get rental")
		return
	}

	if err := h.rental.Cancel(r.Context(), uid, username); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rental not found")
			return
		}
		if errors.Is(err, client.ErrConflict) {
			writeError(w, http.StatusConflict, "rental cannot be canceled")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to cancel rental")
		return
	}

	if err := h.payment.Cancel(r.Context(), rental.PaymentUID); err != nil && !errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "failed to cancel payment")
		return
	}

	if err := h.cars.Unreserve(r.Context(), rental.CarUID); err != nil && !errors.Is(err, client.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "failed to unreserve car")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) enrichRental(ctx context.Context, rental *client.RentalInfo) (*rentalResponse, error) {
	car, err := h.cars.GetByUID(ctx, rental.CarUID)
	if err != nil {
		return nil, err
	}
	payment, err := h.payment.GetByUID(ctx, rental.PaymentUID)
	if err != nil {
		return nil, err
	}
	return &rentalResponse{
		RentalUID: rental.RentalUID,
		Status:    rental.Status,
		DateFrom:  rental.DateFrom,
		DateTo:    rental.DateTo,
		Car: carInfo{
			CarUID:             car.CarUID,
			Brand:              car.Brand,
			Model:              car.Model,
			RegistrationNumber: car.RegistrationNumber,
		},
		Payment: paymentInfo{
			PaymentUID: payment.PaymentUID,
			Status:     payment.Status,
			Price:      payment.Price,
		},
	}, nil
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
