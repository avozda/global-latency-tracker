package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/avozda/global-latency-tracker/internal/hub/tools"

	log "github.com/sirupsen/logrus"
)

func retentionConfig() (time.Duration, time.Duration, error) {
	retention, err := durationFromEnv("PROBE_RESULTS_RETENTION", 7*24*time.Hour)
	if err != nil {
		return 0, 0, err
	}

	cleanupInterval, err := durationFromEnv("PROBE_RESULTS_CLEANUP_INTERVAL", 24*time.Hour)
	if err != nil {
		return 0, 0, err
	}

	return retention, cleanupInterval, nil
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}

	return duration, nil
}

func startProbeResultsRetention(ctx context.Context, db tools.DatabaseInterface, retention time.Duration, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		log.Infof("Probe results retention enabled: keeping %s, cleanup interval %s", retention, interval)
		runProbeResultsCleanup(ctx, db, retention)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runProbeResultsCleanup(ctx, db, retention)
			}
		}
	}()

	return done
}

func runProbeResultsCleanup(ctx context.Context, db tools.DatabaseInterface, retention time.Duration) {
	cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cutoff := time.Now().UTC().Add(-retention)
	deleted, err := db.DeleteProbeResultsBefore(cleanupCtx, cutoff)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.WithError(err).Error("Failed to clean up old probe results")
		return
	}

	log.Infof("Cleaned up %d probe results older than %s", deleted, cutoff.Format(time.RFC3339))
}
