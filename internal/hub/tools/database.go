package tools

import (
	"context"

	"github.com/avozda/global-latency-tracker/internal/probe"
)

type DatabaseInterface interface {
	Close() error
	Ping(ctx context.Context) error
	InsertProbeResult(result probe.Result) (probe.Record, error)
	GetProbeResult(id int64) (probe.Record, error)
	GetProbeResults(limit int, offset int, targetURL string) ([]probe.Record, error)
}

func OpenDatabase(dsn string) (DatabaseInterface, error) {
	return OpenPostgreSQL(dsn)
}
