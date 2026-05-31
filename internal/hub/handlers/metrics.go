package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/avozda/global-latency-tracker/api"
	"github.com/avozda/global-latency-tracker/internal/hub/tools"
	"github.com/avozda/global-latency-tracker/internal/probe"
	log "github.com/sirupsen/logrus"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {
	var result probe.Result
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		log.Error(err)
		api.RequestErrorHandler(w, err)
		return
	}

	var database tools.DatabaseInterface
	database, err = tools.GetDatabase()
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	err = database.InsertProbeResult(result)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
