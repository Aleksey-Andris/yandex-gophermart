package app

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/Aleksey-Andris/yandex-gophermart/config"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

func startMigrations(l *logger.Logger, cfg *config.Config) {
	db, err := sql.Open("pgx", cfg.DBURI)
	if err != nil {
		l.Fatalf("Migrate: failed opening db: %s", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		l.Fatalf("Migrate: failed creating driver: %s", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		l.Fatalf("Migrate: failed creating migrate: %s", err)
	}
	defer m.Close()
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			l.Info("Migrate: no change")
		} else {
			l.Fatalf("Migrate: migrations failed: %s", err)
		}
	}
}
