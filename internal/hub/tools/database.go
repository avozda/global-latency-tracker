package tools

import (
	"github.com/avozda/global-latency-tracker/internal/probe"
)

type DatabaseInterface interface {
	GetDatabase() error
	Close() error
	InsertProbeResult(result probe.Result) error
	GetProbeResult(id int64) (probe.Result, error)
	GetProbeResults(limit int, offset int) ([]probe.Result, error)
}

func GetDatabase() (DatabaseInterface, error) {
	var database DatabaseInterface = &PostgreSQL{}

	var err error = database.GetDatabase()
	if err != nil {
		return nil, err
	}

	return database, nil
}
