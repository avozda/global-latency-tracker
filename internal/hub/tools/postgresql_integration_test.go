//go:build integration

package tools

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/avozda/global-latency-tracker/internal/probe"
)

func setupTestDB(t *testing.T) *PostgreSQL {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping Postgres integration tests")
	}

	if err := RunMigrations(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	db, err := OpenPostgreSQL(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.db.Exec("TRUNCATE probe_results RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	return db
}

func sampleResult(targetURL, region string, measuredAt time.Time) probe.Result {
	return probe.Result{
		TargetURL:          targetURL,
		Region:             region,
		StatusCode:         200,
		DNSLookupMS:        1.1,
		TCPConnectionMS:    2.2,
		TLSHandshakeMS:     3.3,
		ServerProcessingMS: 4.4,
		TTFBMS:             5.5,
		TotalRoundTripMS:   6.6,
		MeasuredAt:         measuredAt,
	}
}

func TestPostgreSQLPing(t *testing.T) {
	db := setupTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestInsertAndGetProbeResult(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	record, err := db.InsertProbeResult(sampleResult("https://example.com", "US-West", now))
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if record.ID == 0 {
		t.Fatal("expected non-zero id")
	}
	if record.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}

	got, err := db.GetProbeResult(record.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.TargetURL != "https://example.com" || got.Region != "US-West" {
		t.Fatalf("unexpected record: %+v", got)
	}
	if got.TotalRoundTripMS != 6.6 {
		t.Fatalf("expected round trip 6.6, got %v", got.TotalRoundTripMS)
	}
}

func TestGetProbeResultsFilteringAndOrdering(t *testing.T) {
	db := setupTestDB(t)

	base := time.Now().UTC().Truncate(time.Millisecond)
	// Insert in non-chronological order to verify ORDER BY measured_at DESC.
	if _, err := db.InsertProbeResult(sampleResult("https://a.com", "US-West", base.Add(-2*time.Hour))); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := db.InsertProbeResult(sampleResult("https://a.com", "US-West", base)); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := db.InsertProbeResult(sampleResult("https://b.com", "Europe", base.Add(-1*time.Hour))); err != nil {
		t.Fatalf("insert: %v", err)
	}

	all, err := db.GetProbeResults(50, 0, "")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 results, got %d", len(all))
	}
	if !all[0].MeasuredAt.After(all[1].MeasuredAt) {
		t.Fatalf("expected results ordered by measured_at DESC")
	}

	filtered, err := db.GetProbeResults(50, 0, "https://a.com")
	if err != nil {
		t.Fatalf("list filtered: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered results, got %d", len(filtered))
	}
	for _, r := range filtered {
		if r.TargetURL != "https://a.com" {
			t.Fatalf("unexpected target in filter: %q", r.TargetURL)
		}
	}

	limited, err := db.GetProbeResults(1, 1, "")
	if err != nil {
		t.Fatalf("list limited: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("expected 1 result with limit/offset, got %d", len(limited))
	}
}

func TestDeleteProbeResultsBefore(t *testing.T) {
	db := setupTestDB(t)

	base := time.Now().UTC().Truncate(time.Millisecond)
	if _, err := db.InsertProbeResult(sampleResult("https://a.com", "US-West", base.Add(-48*time.Hour))); err != nil {
		t.Fatalf("insert old: %v", err)
	}
	if _, err := db.InsertProbeResult(sampleResult("https://a.com", "US-West", base)); err != nil {
		t.Fatalf("insert new: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleted, err := db.DeleteProbeResultsBefore(ctx, base.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}

	remaining, err := db.GetProbeResults(50, 0, "")
	if err != nil {
		t.Fatalf("list remaining: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining, got %d", len(remaining))
	}
}
