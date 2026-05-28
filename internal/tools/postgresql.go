package tools

import (
	"database/sql"
	"os"
)

type PostgreSQL struct {
	db *sql.DB
}

func (p *PostgreSQL) ConnectDatabase() error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	p.db = db
	return nil
}
