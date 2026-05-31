package tools

import (
	"github.com/avozda/global-latency-tracker/internal/probe"
)

type DatabaseInterface interface {
	Close() error
	InsertProbeResult(result probe.Result) error
	GetProbeResult(id int64) (probe.Result, error)
	GetProbeResults(limit int, offset int) ([]probe.Result, error)
}

func OpenDatabase(dsn string) (DatabaseInterface, error) {
	return OpenPostgreSQL(dsn)
}
