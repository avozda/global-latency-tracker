package handlers

import (
	"github.com/avozda/global-latency-tracker/internal/hub/middleware"
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func RegisterRoutes(r chi.Router) {
	r.Use(chimiddle.StripSlashes)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Authenticate)
	})
}
