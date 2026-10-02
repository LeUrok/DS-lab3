package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/LeUrok/DS-lab2/cars-service/internal/model"
	"github.com/LeUrok/DS-lab2/cars-service/internal/service"
	"github.com/go-chi/chi/v5"
)

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

type CarHandler struct {
	svc *service.CarService
}

func NewCarHandler(svc *service.CarService) *CarHandler {
	return &CarHandler{svc: svc}
}

func (h *CarHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *CarHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 0)
	size := parseIntQuery(r, "size", 10)
	showall := r.URL.Query().Get("showAll") == "true"

	cars, total, err := h.svc.GetAll(r.Context(), page, size, showall)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cars")
		return
	}

	items := make([]*carResponse, 0, len(cars))
	for _, c := range cars {
		items = append(items, toResponse(c))
	}
	writeJSON(w, http.StatusOK, paginationResponse{
		Page:          page,
		PageSize:      size,
		TotalElements: total,
		Items:         items,
	})
}

func (h *CarHandler) GetByUID(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "carUid")
	car, err := h.svc.GetByUID(r.Context(), uid)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get car")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(car))
}

func (h *CarHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "carUid")
	err := h.svc.Reserve(r.Context(), uid)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if errors.Is(err, model.ErrAlreadyReserved) {
		writeError(w, http.StatusConflict, "car already reserved")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reserve car")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CarHandler) Unreserve(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "carUid")
	err := h.svc.Unreserve(r.Context(), uid)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if errors.Is(err, model.ErrNotReserved) {
		writeError(w, http.StatusConflict, "car not reserved")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unreserve car")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func toResponse(c *model.Car) *carResponse {
	return &carResponse{
		CarUID:             c.CarUID,
		Brand:              c.Brand,
		Model:              c.Model,
		RegistrationNumber: c.RegistrationNumber,
		Power:              c.Power,
		Type:               c.Type,
		Price:              c.Price,
		Available:          c.Available,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
