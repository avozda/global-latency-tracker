package tools

import (
	"github.com/avozda/global-latency-tracker/internal/probe"
)

type DatabaseInterface interface {
	Close() error
	InsertProbeResult(result probe.Result) (probe.Record, error)
	GetProbeResult(id int64) (probe.Record, error)
	GetProbeResults(limit int, offset int) ([]probe.Record, error)
}

func OpenDatabase(dsn string) (DatabaseInterface, error) {
	return OpenPostgreSQL(dsn)
}
