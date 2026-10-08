package postgres

import (
	"banking-app/internal/config"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// go:embed migrations/*.sql
var migrationFS embed.FS

func runMigrations(cfg config.Config) error {
	src, err := iofs.New(migrationFS, ".")
	if err != nil {
		return err
	}

	url := cfg.PG.MigrateDSN()

	m, err := migrate.NewWithSourceInstance("iofs", src, url)
	if err != nil {
		return err
	}

	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
