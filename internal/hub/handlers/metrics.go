package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/avozda/global-latency-tracker/api"
	"github.com/avozda/global-latency-tracker/internal/probe"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func (a *API) GetMetrics(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			api.RequestErrorHandler(w, errors.New("limit must be an integer between 1 and 100"))
			return
		}
		limit = parsed
	}

	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			api.RequestErrorHandler(w, errors.New("offset must be a non-negative integer"))
			return
		}
		offset = parsed
	}

	targetURL := strings.TrimSpace(r.URL.Query().Get("target_url"))
	if targetURL != "" {
		if err := probe.ValidateTargetURL(targetURL); err != nil {
			api.RequestErrorHandler(w, err)
			return
		}
	}

	var results []probe.Record
	if a.DB != nil {
		var err error
		results, err = a.DB.GetProbeResults(limit, offset, targetURL)
		if err != nil {
			log.WithError(err).Error("Failed to get probe results")
			results = []probe.Record{} // return empty on error
		}
	} else {
		log.Error("Database is not initialized")
		results = []probe.Record{}
	}

	if err := writeJSON(w, http.StatusOK, results); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}

func (a *API) GetMetric(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		api.RequestErrorHandler(w, errors.New("id must be a positive integer"))
		return
	}

	if a.DB == nil {
		log.Error("Database is not initialized")
		api.NotFoundErrorHandler(w)
		return
	}

	record, err := a.DB.GetProbeResult(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.NotFoundErrorHandler(w)
			return
		}
		log.WithError(err).Error("Failed to get probe result")
		api.NotFoundErrorHandler(w) // pretend not found on error to not crash/error out 500? Or maybe empty record?
		return
	}

	if err := writeJSON(w, http.StatusOK, record); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}

func (a *API) PostMetrics(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var result probe.Result

	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		log.Error(err)
		api.RequestErrorHandler(w, err)
		return
	}

	if err := result.Validate(); err != nil {
		log.Error(err)
		api.RequestErrorHandler(w, err)
		return
	}

	log.Info("["+time.Now().Format(time.RFC3339)+"]"+" Probe result received for target URL: ", result.TargetURL, " measured at ", result.MeasuredAt.Format(time.RFC3339))

	var record probe.Record
	if a.DB != nil {
		var err error
		record, err = a.DB.InsertProbeResult(result)
		if err != nil {
			log.WithError(err).Error("Failed to insert probe result")
			// Return a dummy record or just the input mapped to record
			record = probe.Record{
				TargetURL:          result.TargetURL,
				Region:             result.Region,
				StatusCode:         result.StatusCode,
				DNSLookupMS:        result.DNSLookupMS,
				TCPConnectionMS:    result.TCPConnectionMS,
				TLSHandshakeMS:     result.TLSHandshakeMS,
				ServerProcessingMS: result.ServerProcessingMS,
				TTFBMS:             result.TTFBMS,
				TotalRoundTripMS:   result.TotalRoundTripMS,
				MeasuredAt:         result.MeasuredAt,
				Error:              result.Error,
				CreatedAt:          time.Now().UTC(),
			}
		} else {
			log.Info("["+time.Now().Format(time.RFC3339)+"]"+" Probe result inserted successfully for target URL: ", result.TargetURL, " measured at ", result.MeasuredAt.Format(time.RFC3339))
		}
	} else {
		log.Error("Database is not initialized, skipping insert")
		record = probe.Record{
			TargetURL:          result.TargetURL,
			Region:             result.Region,
			StatusCode:         result.StatusCode,
			DNSLookupMS:        result.DNSLookupMS,
			TCPConnectionMS:    result.TCPConnectionMS,
			TLSHandshakeMS:     result.TLSHandshakeMS,
			ServerProcessingMS: result.ServerProcessingMS,
			TTFBMS:             result.TTFBMS,
			TotalRoundTripMS:   result.TotalRoundTripMS,
			MeasuredAt:         result.MeasuredAt,
			Error:              result.Error,
			CreatedAt:          time.Now().UTC(),
		}
	}

	if err := writeJSON(w, http.StatusCreated, record); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Error(err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := io.Copy(w, &buf)
	return err
}
