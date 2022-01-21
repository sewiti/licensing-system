package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

//go:embed migrations
var migrationsDir embed.FS

func Open(dataSource string) (*sql.DB, error) {
	var driver string
	i := strings.IndexRune(dataSource, ':')
	if i >= 0 {
		driver = dataSource[:i]
	}
	return sql.Open(driver, dataSource)
}

func MigrateUp(dataSource string) error {
	src, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("source", src, dataSource)
	if err != nil {
		_ = src.Close()
		return fmt.Errorf("migrations instance: %w", err)
	}

	err = m.Up()
	switch err {
	case nil:
	case migrate.ErrNoChange:
	case os.ErrNotExist:
		// Schema is in unknown state, usually happens after application
		// roll-back when schema is newer than application expected
	default:
		_, _ = m.Close()
		return fmt.Errorf("migrations up: %w", err)
	}

	errSrc, errDrv := m.Close()
	if errSrc != nil {
		return fmt.Errorf("migrations close: source: %w", errSrc)
	}
	if errDrv != nil {
		return fmt.Errorf("migrations close: driver: %w", errDrv)
	}
	return nil
}
