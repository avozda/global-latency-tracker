package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/avozda/global-latency-tracker/internal/hub/handlers"
	"github.com/avozda/global-latency-tracker/internal/hub/tools"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)

	db, err := tools.OpenDatabase(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var r chi.Router = chi.NewRouter()
	handlers.RegisterRoutes(r, db)

	fmt.Println("Starting server on port 8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Error(err)
	}
}
