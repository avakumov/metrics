package database

import (
	"database/sql"
	"embed"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(db *sql.DB) error {
	//install embedded file system
	goose.SetBaseFS(embedMigrations)

	//set dialect DB
	goose.SetDialect("postgres")

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	logger.Log.Info("Migration success")
	return nil
}

func RollbackMigration(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	return goose.Down(db, "migrations")
}
