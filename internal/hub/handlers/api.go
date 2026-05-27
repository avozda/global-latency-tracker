package handlers

import (
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func RegisterRoutes(r chi.Router) {
	r.Use(chimiddle.StripSlashes)
}
