package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func Connect() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@localhost:5432/oneagent?sslmode=disable"
	}

	var err error
	DB, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

func runMigrations() error {
	// Read migrations file
	migrationPath := filepath.Join("internal", "db", "migrations.sql")
	migrations, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations file: %v", err)
	}

	// Execute migrations
	_, err = DB.Exec(context.Background(), string(migrations))
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %v", err)
	}

	return nil
}
