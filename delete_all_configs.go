package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
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

	// First, let's see how many configurations exist
	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM user_configurations").Scan(&count)
	if err != nil {
		fmt.Printf("Failed to count configurations: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d configurations in the database\n", count)

	if count == 0 {
		fmt.Println("No configurations to delete.")
		return
	}

	// Ask for confirmation
	fmt.Print("Are you sure you want to delete ALL configurations? This action cannot be undone. (yes/no): ")
	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		fmt.Println("Operation cancelled.")
		return
	}

	// Delete all configurations
	// Note: This will also delete related records due to CASCADE constraints
	result, err := conn.Exec(context.Background(), "DELETE FROM user_configurations")
	if err != nil {
		fmt.Printf("Failed to delete configurations: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted %d configurations and all related data\n", result.RowsAffected())
} 