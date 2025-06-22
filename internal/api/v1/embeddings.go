package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type openAIEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}
	if err := validateEnvVars(); err != nil {
		fmt.Printf("Environment validation failed: %v\n", err)
	}
}

// getEmbedding retrieves embeddings for the given text using the OpenAI API
func getEmbedding(text string, model string) ([]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	if model == "" {
		model = "text-embedding-3-small" // Default model
	}

	body, err := json.Marshal(openAIEmbeddingRequest{
		Input: text,
		Model: model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var rawBody bytes.Buffer
		rawBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", rawBody.String())
	}

	var result openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return result.Data[0].Embedding, nil
}

// updateEmbeddings updates product embeddings in the database for records where embedding is NULL
func updateEmbeddings() error {
	ctx := context.Background()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// First, check if there are any products in the database
	var productCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&productCount)
	if err != nil {
		return fmt.Errorf("failed to count products: %w", err)
	}
	
	if productCount == 0 {
		fmt.Println("No products found in database. Please upload products first.")
		return fmt.Errorf("no products found in database")
	}

	// Check how many products need embeddings
	var nullEmbeddingCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE embedding IS NULL`).Scan(&nullEmbeddingCount)
	if err != nil {
		return fmt.Errorf("failed to count products without embeddings: %w", err)
	}

	fmt.Printf("Found %d products in database, %d need embeddings\n", productCount, nullEmbeddingCount)

	if nullEmbeddingCount == 0 {
		fmt.Println("All products already have embeddings.")
		return nil
	}

	rows, err := pool.Query(ctx, `SELECT id, name, category, description FROM products WHERE embedding IS NULL`)
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	successCount := 0
	errorCount := 0

	for rows.Next() {
		var id int
		var name, category, desc string
		if err := rows.Scan(&id, &name, &category, &desc); err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			errorCount++
			continue
		}

		// Combine name, category, and description for better embedding
		var combinedText string
		if category != "" {
			combinedText = fmt.Sprintf("%s. Category: %s. %s", name, category, desc)
		} else {
			combinedText = fmt.Sprintf("%s. %s", name, desc)
		}

		fmt.Printf("Processing product %d: %s\n", id, name)

		embedding, err := getEmbedding(combinedText, "")
		if err != nil {
			fmt.Printf("Embedding error for product %d (%s): %v\n", id, name, err)
			errorCount++
			continue
		}

		vectorStr := fmt.Sprintf("[%s]", formatFloatSlice(embedding))
		
		_, err = pool.Exec(ctx, `UPDATE products SET embedding = $1::vector WHERE id = $2`, vectorStr, id)
		if err != nil {
			fmt.Printf("DB update error for product %d (%s): %v\n", id, name, err)
			errorCount++
			continue
		}

		fmt.Printf("Successfully updated embedding for product %d: %s\n", id, name)
		successCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over rows: %w", err)
	}

	fmt.Printf("Embedding update completed. Success: %d, Errors: %d\n", successCount, errorCount)
	return nil
}

// formatFloatSlice formats a slice of float32 values into a comma-separated string
func formatFloatSlice(floats []float32) string {
	strs := make([]string, len(floats))
	for i, f := range floats {
		strs[i] = fmt.Sprintf("%f", f)
	}
	return strings.Join(strs, ",")
}

// validateEnvVars checks if all required environment variables are set
func validateEnvVars() error {
	required := []string{"OPENAI_API_KEY", "DATABASE_URL"}
	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}
	return nil
}