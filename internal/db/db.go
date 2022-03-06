package db

import (
	"database/sql"
	"strings"

	"github.com/Masterminds/squirrel"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

type Handler struct {
	db *sql.DB
	sq squirrel.StatementBuilderType
}

func Open(dataSource string) (*Handler, error) {
	var driver string
	i := strings.IndexRune(dataSource, ':')
	if i >= 0 {
		driver = dataSource[:i]
	}
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		return nil, err
	}
	var sq squirrel.StatementBuilderType
	switch driver {
	case "postgres":
		sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	default:
		sq = squirrel.StatementBuilder
	}
	return &Handler{
		db: db,
		sq: sq.RunWith(db),
	}, nil
}

func (h *Handler) Close() error {
	return h.db.Close()
}
