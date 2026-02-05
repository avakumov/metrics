package database

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(DSN string) error {

	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	defer db.Close()
	logger.Log.Info("✅ Connected to PostgreSQL for migrate")

	//install embedded file system
	goose.SetBaseFS(embedMigrations)

	//set dialect DB
	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	logger.Log.Info("Migration success")
	return nil
}

func RollbackMigration(DSN string) error {

	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("✅ Connected to PostgreSQL for migrate")

	goose.SetBaseFS(embedMigrations)
	return goose.Down(db, "migrations")
}
