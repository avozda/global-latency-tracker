package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/avozda/global-latency-tracker/api"
	"github.com/avozda/global-latency-tracker/internal/probe"
	log "github.com/sirupsen/logrus"
)

func (a *API) PostMetrics(w http.ResponseWriter, r *http.Request) {
	var result probe.Result
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		log.Error(err)
		api.RequestErrorHandler(w, err)
		return
	}

	log.Info("["+time.Now().Format(time.RFC3339)+"]"+" Probe result received for target URL: ", result.TargetURL, " measured at ", result.MeasuredAt.Format(time.RFC3339))

	record, err := a.DB.InsertProbeResult(result)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	log.Info("["+time.Now().Format(time.RFC3339)+"]"+" Probe result inserted successfully for target URL: ", result.TargetURL, " measured at ", result.MeasuredAt.Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}
