package migrations

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Up(dir, dbConn string) error {
	return up(dir, dbConn)
}

func up(dir, dbConn string) error {
	m, err := migrate.New(dir, dbConn)
	if err != nil {
		return fmt.Errorf("setting up migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
