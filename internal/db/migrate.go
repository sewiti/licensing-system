package db

import (
	"embed"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsDir embed.FS

func MigrateUp(dataSource string) (migrated bool, err error) {
	src, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return false, fmt.Errorf("migrations source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("source", src, dataSource)
	if err != nil {
		_ = src.Close()
		return false, fmt.Errorf("migrations instance: %w", err)
	}

	err = m.Up()
	switch err {
	case nil:
		migrated = true
	case migrate.ErrNoChange:
	case os.ErrNotExist:
		// Schema is in unknown state, usually happens after application
		// roll-back when schema is newer than application expected
	default:
		_, _ = m.Close()
		return migrated, fmt.Errorf("migrations up: %w", err)
	}

	errSrc, errDrv := m.Close()
	if errSrc != nil {
		return migrated, fmt.Errorf("migrations close: source: %w", errSrc)
	}
	if errDrv != nil {
		return migrated, fmt.Errorf("migrations close: driver: %w", errDrv)
	}
	return migrated, nil
}
