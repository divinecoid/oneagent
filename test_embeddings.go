package main

import (
	"context"
	"fmt"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/oneagent?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Check if products table exists
	var tableExists bool
	err = pool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'products'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		fmt.Printf("Error checking if products table exists: %v\n", err)
		os.Exit(1)
	}

	if !tableExists {
		fmt.Println("❌ Products table does not exist!")
		fmt.Println("Please run: go run migrate.go")
		os.Exit(1)
	}

	fmt.Println("✅ Products table exists")

	// Check if vector extension is enabled
	var extensionExists bool
	err = pool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT FROM pg_extension 
			WHERE extname = 'vector'
		)
	`).Scan(&extensionExists)
	
	if err != nil {
		fmt.Printf("Error checking vector extension: %v\n", err)
		os.Exit(1)
	}

	if !extensionExists {
		fmt.Println("❌ Vector extension is not enabled!")
		fmt.Println("Please run: go run migrate.go")
		os.Exit(1)
	}

	fmt.Println("✅ Vector extension is enabled")

	// Check product count
	var productCount int
	err = pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM products`).Scan(&productCount)
	if err != nil {
		fmt.Printf("Error counting products: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 Found %d products in database\n", productCount)

	// Check products without embeddings
	var nullEmbeddingCount int
	err = pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM products WHERE embedding IS NULL`).Scan(&nullEmbeddingCount)
	if err != nil {
		fmt.Printf("Error counting products without embeddings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 Found %d products without embeddings\n", nullEmbeddingCount)

	// Show sample products
	if productCount > 0 {
		fmt.Println("\n📋 Sample products:")
		rows, err := pool.Query(context.Background(), `
			SELECT id, name, category, price, description, 
			       CASE WHEN embedding IS NULL THEN 'NULL' ELSE 'SET' END as embedding_status
			FROM products 
			LIMIT 5
		`)
		if err != nil {
			fmt.Printf("Error querying sample products: %v\n", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var id int
				var name, category, description, embeddingStatus string
				var price float64
				err := rows.Scan(&id, &name, &category, &price, &description, &embeddingStatus)
				if err == nil {
					fmt.Printf("  ID: %d, Name: %s, Category: %s, Price: %.2f, Embedding: %s\n", 
						id, name, category, price, embeddingStatus)
				}
			}
		}
	}

	// Check environment variables
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		fmt.Println("\n❌ OPENAI_API_KEY environment variable is not set!")
		fmt.Println("Please set it before running embeddings update")
	} else {
		fmt.Printf("\n✅ OPENAI_API_KEY is set (length: %d)\n", len(openaiKey))
	}

	fmt.Println("\n🎯 To update embeddings, run:")
	fmt.Println("POST http://localhost:8080/api/v1/products/update-embeddings")
} 