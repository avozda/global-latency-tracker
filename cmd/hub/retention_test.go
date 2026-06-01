package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avozda/global-latency-tracker/internal/probe"
)

type fakeRetentionDB struct {
	cutoff    time.Time
	called    bool
	deleted   int64
	returnErr error
}

func (f *fakeRetentionDB) Close() error                          { return nil }
func (f *fakeRetentionDB) Ping(context.Context) error            { return nil }
func (f *fakeRetentionDB) InsertProbeResult(probe.Result) (probe.Record, error) {
	return probe.Record{}, nil
}
func (f *fakeRetentionDB) GetProbeResult(int64) (probe.Record, error) { return probe.Record{}, nil }
func (f *fakeRetentionDB) GetProbeResults(int, int, string) ([]probe.Record, error) {
	return nil, nil
}
func (f *fakeRetentionDB) DeleteProbeResultsBefore(_ context.Context, cutoff time.Time) (int64, error) {
	f.called = true
	f.cutoff = cutoff
	return f.deleted, f.returnErr
}

func TestDurationFromEnv(t *testing.T) {
	t.Run("default when unset", func(t *testing.T) {
		t.Setenv("SOME_DURATION", "")
		got, err := durationFromEnv("SOME_DURATION", 5*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 5*time.Minute {
			t.Fatalf("expected fallback 5m, got %v", got)
		}
	})

	t.Run("valid value", func(t *testing.T) {
		t.Setenv("SOME_DURATION", "2h")
		got, err := durationFromEnv("SOME_DURATION", 5*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 2*time.Hour {
			t.Fatalf("expected 2h, got %v", got)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		t.Setenv("SOME_DURATION", "nonsense")
		if _, err := durationFromEnv("SOME_DURATION", 5*time.Minute); err == nil {
			t.Fatal("expected error for invalid duration")
		}
	})

	t.Run("non-positive value", func(t *testing.T) {
		t.Setenv("SOME_DURATION", "0s")
		if _, err := durationFromEnv("SOME_DURATION", 5*time.Minute); err == nil {
			t.Fatal("expected error for non-positive duration")
		}
	})
}

func TestRetentionConfig(t *testing.T) {
	t.Setenv("PROBE_RESULTS_RETENTION", "48h")
	t.Setenv("PROBE_RESULTS_CLEANUP_INTERVAL", "1h")

	retention, interval, err := retentionConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retention != 48*time.Hour {
		t.Fatalf("expected 48h retention, got %v", retention)
	}
	if interval != time.Hour {
		t.Fatalf("expected 1h interval, got %v", interval)
	}
}

func TestRunProbeResultsCleanup(t *testing.T) {
	t.Run("computes cutoff", func(t *testing.T) {
		db := &fakeRetentionDB{deleted: 3}
		retention := 24 * time.Hour
		before := time.Now().UTC().Add(-retention)

		runProbeResultsCleanup(context.Background(), db, retention)

		if !db.called {
			t.Fatal("expected DeleteProbeResultsBefore to be called")
		}
		after := time.Now().UTC().Add(-retention)
		if db.cutoff.Before(before.Add(-time.Second)) || db.cutoff.After(after.Add(time.Second)) {
			t.Fatalf("cutoff %v not within expected window", db.cutoff)
		}
	})

	t.Run("tolerates errors", func(t *testing.T) {
		db := &fakeRetentionDB{returnErr: errors.New("boom")}
		runProbeResultsCleanup(context.Background(), db, time.Hour)
		if !db.called {
			t.Fatal("expected DeleteProbeResultsBefore to be called")
		}
	})
}
