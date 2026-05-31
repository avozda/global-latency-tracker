package handlers

import (
	"github.com/avozda/global-latency-tracker/internal/hub/middleware"
	"github.com/avozda/global-latency-tracker/internal/hub/tools"
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

type API struct {
	DB tools.DatabaseInterface
}

func RegisterRoutes(r chi.Router, db tools.DatabaseInterface) {
	api := &API{DB: db}

	r.Use(chimiddle.StripSlashes)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Get("/metrics", api.GetMetrics)
		r.Get("/metrics/{id}", api.GetMetric)
		r.Post("/metrics", api.PostMetrics)
	})
}
