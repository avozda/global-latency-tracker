package tools

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/avozda/global-latency-tracker/internal/probe"

	_ "github.com/lib/pq"
)

type PostgreSQL struct {
	db *sql.DB
}

func (p *PostgreSQL) GetDatabase() error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return err
	}

	p.db = db
	return nil
}

func (p *PostgreSQL) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

func (p *PostgreSQL) InsertProbeResult(result probe.Result) error {
	_, err := p.db.Exec(`
INSERT INTO probe_results (
	target_url, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		result.TargetURL, result.StatusCode, result.DNSLookupMS, result.TCPConnectionMS,
		result.TLSHandshakeMS, result.ServerProcessingMS, result.TTFBMS, result.TotalRoundTripMS,
		result.MeasuredAt, result.Error,
	)
	return err
}

func (p *PostgreSQL) GetProbeResult(id int64) (probe.Result, error) {
	var result probe.Result
	err := p.db.QueryRow(`
SELECT
	target_url, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error
FROM probe_results
WHERE id = $1`, id).Scan(
		&result.TargetURL, &result.StatusCode, &result.DNSLookupMS, &result.TCPConnectionMS,
		&result.TLSHandshakeMS, &result.ServerProcessingMS, &result.TTFBMS, &result.TotalRoundTripMS,
		&result.MeasuredAt, &result.Error,
	)
	return result, err
}

func (p *PostgreSQL) GetProbeResults(limit int, offset int) ([]probe.Result, error) {
	var results []probe.Result
	err := p.db.QueryRow(`
SELECT
	target_url, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error
FROM probe_results
LIMIT $1 OFFSET $2`, limit, offset).Scan(&results)
	return results, err
}
