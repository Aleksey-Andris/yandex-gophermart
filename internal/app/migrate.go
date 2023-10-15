package app

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/Aleksey-Andris/yandex-gophermart/config"
)

func startMigration(cfg *config.Config) error {
	db, err := sql.Open("postgres", cfg.PG.URL)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"./migrations/",
		"postgres", driver)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Up()
}
