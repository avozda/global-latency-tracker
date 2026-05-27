package main

import (
	"fmt"
	"net/http"

	"github.com/avozda/global-latency-tracker/internal/hub/handlers"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	var r chi.Router = chi.NewRouter()

	handlers.RegisterRoutes(r)

	fmt.Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Error(err)
	}
}
