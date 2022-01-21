package web

import (
	"database/sql"
)

type Runtime struct {
	db *sql.DB
}

func NewRuntime(db *sql.DB) *Runtime {
	return &Runtime{
		db: db,
	}
}
