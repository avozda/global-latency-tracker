package tools

import (
	"context"
	"database/sql"
	"time"

	"github.com/avozda/global-latency-tracker/internal/probe"

	_ "github.com/lib/pq"
)

type PostgreSQL struct {
	db *sql.DB
}

func OpenPostgreSQL(dsn string) (*PostgreSQL, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &PostgreSQL{db: db}, nil
}

func (p *PostgreSQL) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

func (p *PostgreSQL) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgreSQL) InsertProbeResult(result probe.Result) (probe.Record, error) {
	var record probe.Record
	err := p.db.QueryRow(`
INSERT INTO probe_results (
	target_url, region, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
	id, target_url, region, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error, created_at`,
		result.TargetURL, result.Region, result.StatusCode, result.DNSLookupMS, result.TCPConnectionMS,
		result.TLSHandshakeMS, result.ServerProcessingMS, result.TTFBMS, result.TotalRoundTripMS,
		result.MeasuredAt, result.Error,
	).Scan(
		&record.ID, &record.TargetURL, &record.Region, &record.StatusCode, &record.DNSLookupMS, &record.TCPConnectionMS,
		&record.TLSHandshakeMS, &record.ServerProcessingMS, &record.TTFBMS, &record.TotalRoundTripMS,
		&record.MeasuredAt, &record.Error, &record.CreatedAt,
	)
	return record, err
}

func (p *PostgreSQL) GetProbeResult(id int64) (probe.Record, error) {
	var result probe.Record
	err := p.db.QueryRow(`
SELECT
	id, target_url, region, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error, created_at
FROM probe_results
WHERE id = $1`, id).Scan(
		&result.ID, &result.TargetURL, &result.Region, &result.StatusCode, &result.DNSLookupMS, &result.TCPConnectionMS,
		&result.TLSHandshakeMS, &result.ServerProcessingMS, &result.TTFBMS, &result.TotalRoundTripMS,
		&result.MeasuredAt, &result.Error, &result.CreatedAt,
	)
	return result, err
}

func (p *PostgreSQL) GetProbeResults(limit int, offset int, targetURL string) ([]probe.Record, error) {
	var rows *sql.Rows
	var err error

	if targetURL == "" {
		rows, err = p.db.Query(`
SELECT
	id, target_url, region, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error, created_at
FROM probe_results
ORDER BY measured_at DESC
LIMIT $1 OFFSET $2`, limit, offset)
	} else {
		rows, err = p.db.Query(`
SELECT
	id, target_url, region, status_code, dns_lookup_ms, tcp_connection_ms,
	tls_handshake_ms, server_processing_ms, ttfb_ms, total_roundtrip_ms,
	measured_at, error, created_at
FROM probe_results
WHERE target_url = $1
ORDER BY measured_at DESC
LIMIT $2 OFFSET $3`, targetURL, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]probe.Record, 0)
	for rows.Next() {
		var result probe.Record
		if err := rows.Scan(
			&result.ID, &result.TargetURL, &result.Region, &result.StatusCode, &result.DNSLookupMS, &result.TCPConnectionMS,
			&result.TLSHandshakeMS, &result.ServerProcessingMS, &result.TTFBMS, &result.TotalRoundTripMS,
			&result.MeasuredAt, &result.Error, &result.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (p *PostgreSQL) DeleteProbeResultsBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := p.db.ExecContext(ctx, `
DELETE FROM probe_results
WHERE measured_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
