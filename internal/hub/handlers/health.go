package handlers

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (a *API) Healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if a.DB != nil {
		if err := a.DB.Ping(ctx); err != nil {
			log.WithError(err).Error("Database ping failed")
		}
	} else {
		log.Error("Database is not initialized")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
