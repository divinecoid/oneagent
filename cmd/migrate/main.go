package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Read migration file
	migrationPath := filepath.Join("internal", "db", "migrations_update_gpt4.sql")
	migrations, err := os.ReadFile(migrationPath)
	if err != nil {
		fmt.Printf("Failed to read migrations file: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@localhost:5432/oneagent?sslmode=disable"
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Execute migration
	_, err = conn.Exec(context.Background(), string(migrations))
	if err != nil {
		fmt.Printf("Failed to execute migrations: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated configurations to use GPT-4-turbo-preview")
} 