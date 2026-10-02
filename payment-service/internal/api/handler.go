package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LeUrok/DS-lab2/payment-service/internal/model"
	"github.com/LeUrok/DS-lab2/payment-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

type paymentResponse struct {
	PaymentUID string `json:"paymentUid"`
	Status     string `json:"status"`
	Price      int    `json:"price"`
}

type createPaymentRequest struct {
	Price int `json:"price"`
}

func (h *PaymentHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payment, err := h.svc.Create(r.Context(), req.Price)
	if errors.Is(err, model.ErrInvalidPrice) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(payment))
}

func (h *PaymentHandler) GetByUID(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "paymentUid")
	payment, err := h.svc.GetByUID(r.Context(), uid)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "payment not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get payment")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(payment))
}

func (h *PaymentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "paymentUid")
	err := h.svc.Cancel(r.Context(), uid)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "payment not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to cancel payment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func toResponse(p *model.Payment) paymentResponse {
	return paymentResponse{
		PaymentUID: p.PaymentUID,
		Status:     p.Status,
		Price:      p.Price,
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
