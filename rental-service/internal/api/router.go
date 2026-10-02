package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *RentalHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/manage/health", h.Health)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/rental", h.GetByUsername)
		r.Post("/rental", h.Create)
		r.Get("/rental/{rentalUid}", h.GetByUID)
		r.Post("/rental/{rentalUid}/finish", h.Finish)
		r.Delete("/rental/{rentalUid}", h.Cancel)
	})
	return r
}
