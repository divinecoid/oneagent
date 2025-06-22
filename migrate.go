package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Read migration file for adding seller_id to products
	migrationPath := filepath.Join("internal", "db", "migrations_products_add_seller.sql")
	migrationsBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		fmt.Printf("Failed to read migrations file: %v\n", err)
		os.Exit(1)
	}

	migrations := string(migrationsBytes)

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

	statements := splitSQLStatementsDollar(migrations)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		fmt.Printf("Executing statement:\n%s\n---\n", stmt)
		_, err := conn.Exec(context.Background(), stmt)
		if err != nil {
			fmt.Printf("Failed to execute statement: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Successfully ran add seller_id to products migration")
}

// splitSQLStatementsDollar splits SQL by semicolons, but keeps everything between $$ and $$; as a single statement
func splitSQLStatementsDollar(sql string) []string {
	var stmts []string
	var sb strings.Builder
	inDollar := false
	for i := 0; i < len(sql); i++ {
		if !inDollar && i+1 < len(sql) && sql[i] == '$' && sql[i+1] == '$' {
			inDollar = true
			sb.WriteByte(sql[i])
			i++
			sb.WriteByte(sql[i])
			continue
		}
		if inDollar && i+2 < len(sql) && sql[i] == '$' && sql[i+1] == '$' && sql[i+2] == ';' {
			// End of dollar-quoted block
			sb.WriteByte(sql[i])
			sb.WriteByte(sql[i+1])
			sb.WriteByte(sql[i+2])
			stmts = append(stmts, sb.String())
			sb.Reset()
			inDollar = false
			i += 2
			continue
		}
		if !inDollar && sql[i] == ';' {
			stmts = append(stmts, sb.String()+";")
			sb.Reset()
			continue
		}
		sb.WriteByte(sql[i])
	}
	if sb.Len() > 0 {
		stmts = append(stmts, sb.String())
	}
	return stmts
} 