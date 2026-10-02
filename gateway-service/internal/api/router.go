package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/manage/health", h.Health)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/cars", h.GetCars)

		r.Get("/rental", h.GetRentals)
		r.Post("/rental", h.CreateRental)
		r.Get("/rental/{rentalUid}", h.GetRentalByUID)
		r.Post("/rental/{rentalUid}/finish", h.FinishRental)
		r.Delete("/rental/{rentalUid}", h.CancelRental)
	})
	return r
}
