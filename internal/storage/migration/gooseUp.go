package migration

import (
	"database/sql"
	"github.com/pressly/goose/v3"
	"log"
	"path/filepath"
	"runtime"
)

func GooseUp(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose.SetDialect: %v", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)
	migrationsPath := filepath.Join(basePath, "../../../migrations")

	err := goose.Up(db, migrationsPath)
	if err != nil {
		log.Fatalf("goose.Up: %v", err)
	}
	return err
}
