package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/LeUrok/DS-lab2/rental-service/internal/model"
	"github.com/LeUrok/DS-lab2/rental-service/internal/service"
	"github.com/go-chi/chi/v5"
)

const dateLayout = "2006-01-02"

type rentalResponse struct {
	RentalUID  string `json:"rentalUid"`
	Username   string `json:"username"`
	PaymentUID string `json:"paymentUid"`
	CarUID     string `json:"carUid"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
	Status     string `json:"status"`
}

type RentalHandler struct {
	svc *service.RentalService
}

func NewRentalHandler(svc *service.RentalService) *RentalHandler {
	return &RentalHandler{svc: svc}
}

func (h *RentalHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *RentalHandler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	rentals, err := h.svc.GetByUsername(r.Context(), username)
	if errors.Is(err, model.ErrForbidden) {
		writeError(w, http.StatusForbidden, "username is required")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get rentals")
		return
	}

	items := make([]rentalResponse, 0, len(rentals))
	for _, rental := range rentals {
		items = append(items, toResponse(rental))
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *RentalHandler) GetByUID(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "rentalUid")
	username := r.URL.Query().Get("username")

	rental, err := h.svc.GetByUID(r.Context(), uid, username)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "rental not found")
		return
	}
	if errors.Is(err, model.ErrForbidden) {
		writeError(w, http.StatusForbidden, "rental does not belong to user")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get rental")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(rental))
}

type createRentalRequest struct {
	Username   string `json:"username"`
	CarUID     string `json:"carUid"`
	PaymentUID string `json:"paymentUid"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
}

func (h *RentalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dateFrom, err := parseDate(req.DateFrom)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dateFrom")
		return
	}
	dateTo, err := parseDate(req.DateTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dateTo")
		return
	}

	rental, err := h.svc.Create(r.Context(), &model.Rental{
		Username:   req.Username,
		CarUID:     req.CarUID,
		PaymentUID: req.PaymentUID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	})
	if errors.Is(err, model.ErrInvalidDates) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create rental")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(rental))
}

func (h *RentalHandler) Finish(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "rentalUid")
	username := r.URL.Query().Get("username")

	err := h.svc.Finish(r.Context(), uid, username)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "rental not found")
		return
	}
	if errors.Is(err, model.ErrForbidden) {
		writeError(w, http.StatusForbidden, "rental does not belong to user")
		return
	}
	if errors.Is(err, model.ErrAlreadyFinished) {
		writeError(w, http.StatusConflict, "rental already finished")
		return
	}
	if errors.Is(err, model.ErrAlreadyCanceled) {
		writeError(w, http.StatusConflict, "rental already canceled")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to finish rental")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RentalHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "rentalUid")
	username := r.URL.Query().Get("username")

	err := h.svc.Cancel(r.Context(), uid, username)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "rental not found")
		return
	}
	if errors.Is(err, model.ErrForbidden) {
		writeError(w, http.StatusForbidden, "rental does not belong to user")
		return
	}
	if errors.Is(err, model.ErrAlreadyFinished) {
		writeError(w, http.StatusConflict, "rental already finished")
		return
	}
	if errors.Is(err, model.ErrAlreadyCanceled) {
		writeError(w, http.StatusConflict, "rental already canceled")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to cancel rental")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toResponse(r *model.Rental) rentalResponse {
	return rentalResponse{
		RentalUID:  r.RentalUID,
		Username:   r.Username,
		PaymentUID: r.PaymentUID,
		CarUID:     r.CarUID,
		DateFrom:   r.DateFrom.Format("2006-01-02"),
		DateTo:     r.DateTo.Format("2006-01-02"),
		Status:     r.Status,
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

func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}
