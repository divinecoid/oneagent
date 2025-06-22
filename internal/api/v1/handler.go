package v1

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "encoding/json"
    "bytes"
    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
    "github.com/divinecoid/oneagent/internal/model"
    "github.com/divinecoid/oneagent/internal/db"
    "github.com/jackc/pgx/v5/pgxpool"
    "time"
    "github.com/divinecoid/oneagent/internal/service"
)

func GetUsers(c *gin.Context) {
    data := gin.H{
        "users": []string{"Alice", "Bob"},
    }
    
    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Users retrieved successfully",
        Data:    data,
        Errors:  nil,
        Meta: MetaData{
            RequestID: c.GetHeader("X-Request-ID"),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        },
    })
}

func UploadProductExcel(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Message: "File upload failed",
            Data:    nil,
            Errors: gin.H{
                "upload_error": "Failed to process uploaded file",
                "details":     err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    f, err := file.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to open file",
            Data:    nil,
            Errors: gin.H{
                "file_error": "Failed to open uploaded file",
                "details":    err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }
    defer f.Close()

    xlsx, err := excelize.OpenReader(f)
    if err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Message: "Invalid Excel file",
            Data:    nil,
            Errors: gin.H{
                "file_error": "Failed to parse Excel file",
                "details":   err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    rows, err := xlsx.GetRows("Products")
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to read sheet",
            Data:    nil,
            Errors: gin.H{
                "sheet_error": "Failed to read Excel sheet",
                "details":     err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    fmt.Println("Excel Data:")
    fmt.Println("Row 0 (Headers):", strings.Join(rows[0], ", "))
    
    for i, row := range rows {
        if i == 0 {
            continue // skip header
        }

        if len(row) < 4 {
            fmt.Printf("Row %d: Skipped (insufficient columns, got %d columns)\n", i, len(row))
            continue
        }

        price, err := strconv.ParseFloat(row[2], 64)
        if err != nil {
            fmt.Printf("Row %d: Skipped (invalid price format: %s)\n", i, row[2])
            continue
        }

        product := model.Product{
            Name:        row[0],
            Category:    row[1],
            Price:       price,
            Description: row[3],
        }

        fmt.Printf("Row %d: Name=%s, Category=%s, Price=%.2f, Description=%s\n",
            i, product.Name, product.Category, product.Price, product.Description)

        _, err = db.DB.Exec(context.Background(), `
            INSERT INTO products (name, category, price, description)
            VALUES ($1, $2, $3, $4)
        `, product.Name, product.Category, product.Price, product.Description)

        if err != nil {
            fmt.Printf("DB error on row %d: %v\n", i, err)
        } else {
            fmt.Printf("Row %d: Successfully inserted into database\n", i)
        }
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Upload successful",
        Data:    nil,
        Errors:  nil,
        Meta: MetaData{
            RequestID: c.GetHeader("X-Request-ID"),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        },
    })
}

// UpdateProductEmbeddings updates embeddings for all products that don't have embeddings yet
func UpdateProductEmbeddings(c *gin.Context) {
    if err := updateEmbeddings(); err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to update embeddings",
            Data:    nil,
            Errors: gin.H{
                "update_error": "Failed to update product embeddings",
                "details":     err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Successfully updated product embeddings",
        Data:    nil,
        Errors:  nil,
        Meta: MetaData{
            RequestID: c.GetHeader("X-Request-ID"),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        },
    })
}


type SearchRequest struct {
    Query string `json:"query" binding:"required"`
    Limit int    `json:"limit,omitempty"`
}

type SearchResult struct {
    ID          int64   `json:"id"`
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
    Similarity  float64 `json:"similarity"`
}

func SearchProducts(c *gin.Context) {
    var req SearchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Message: "Invalid request parameters",
            Data:    nil,
            Errors: gin.H{
                "validation_error": err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    if req.Limit <= 0 {
        req.Limit = 5 // default limit
    }

    // Get embedding for the search query
    queryEmbedding, err := getEmbedding(req.Query, "text-embedding-3-small")
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to process search query",
            Data:    nil,
            Errors: gin.H{
                "embedding_error": fmt.Sprintf("Failed to generate embedding: %v", err),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    // Convert the embedding to PostgreSQL vector format
    vectorStr := fmt.Sprintf("[%s]", formatFloatSlice(queryEmbedding))

    // Query the database using vector similarity
    ctx := context.Background()
    dbURL := os.Getenv("DATABASE_URL")
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Database connection error",
            Data:    nil,
            Errors: gin.H{
                "database_error": "Failed to connect to database",
                "details":       err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }
    defer pool.Close()

    rows, err := pool.Query(ctx, `
        SELECT id, name, category, price, description, 1 - (embedding <=> $1::vector) as similarity
        FROM products
        WHERE embedding IS NOT NULL
        ORDER BY embedding <=> $1::vector
        LIMIT $2
    `, vectorStr, req.Limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search query failed: %v", err)})
        return
    }
    defer rows.Close()

    var results []SearchResult
    for rows.Next() {
        var result SearchResult
        err := rows.Scan(&result.ID, &result.Name, &result.Price, &result.Description, &result.Similarity)
        if err != nil {
            c.JSON(http.StatusInternalServerError, APIResponse{
                Success: false,
                Message: "Error scanning search results",
                Data:    nil,
                Errors:  err.Error(),
                Meta: MetaData{
                    RequestID: c.GetHeader("X-Request-ID"),
                    Timestamp: time.Now().UTC().Format(time.RFC3339),
                },
            })
            return
        }
        results = append(results, result)
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Products retrieved successfully",
        Data:    gin.H{"products": results},
        Errors:  nil,
        Meta: MetaData{
            RequestID: c.GetHeader("X-Request-ID"),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        },
    })
}

type ChatRequest struct {
    Question string `json:"question" binding:"required"`
}

type ChatResponse struct {
    Answer string `json:"answer"`
}

type OpenAIRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    MaxTokens int      `json:"max_tokens,omitempty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OpenAIResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

func ChatWithProducts(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, APIResponse{
            Success: false,
            Message: "Invalid request parameters",
            Errors: gin.H{
                "validation_error": err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    // Get user configuration
    userID := c.MustGet("user_id").(int64)
    
    config, err := getConfigurationForUser(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to get user configuration",
            Errors: gin.H{
                "config_error": err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    // Get embedding for the question
    queryEmbedding, err := getEmbedding(req.Question, config.OpenAIEmbeddingModel)
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to process question",
            Errors: gin.H{
                "embedding_error": fmt.Sprintf("Failed to generate embedding: %v", err),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    // Convert the embedding to PostgreSQL vector format
    vectorStr := fmt.Sprintf("[%s]", formatFloatSlice(queryEmbedding))

    // Query the database using vector similarity
    ctx := context.Background()
    dbURL := os.Getenv("DATABASE_URL")
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Database connection error",
            Errors: gin.H{
                "database_error": "Failed to connect to database",
                "details":       err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }
    defer pool.Close()

    // Get top 3 most similar products
    rows, err := pool.Query(ctx, `
        SELECT id, name, price, description, 1 - (embedding <=> $1::vector) as similarity
        FROM products
        WHERE embedding IS NOT NULL
        ORDER BY embedding <=> $1::vector
        LIMIT 3
    `, vectorStr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search query failed: %v", err)})
        return
    }
    defer rows.Close()

    var results []SearchResult
    for rows.Next() {
        var result SearchResult
        err := rows.Scan(&result.ID, &result.Name, &result.Price, &result.Description, &result.Similarity)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error scanning results: %v", err)})
            return
        }
        results = append(results, result)
    }

    // Generate context for OpenAI
    context := fmt.Sprintf("Question: %s\n\nAvailable products:\n%s", 
        req.Question, 
        formatProductList(results),
    )

    // Call OpenAI API for response generation
    answer, err := generateOpenAIResponse(context, config)
    if err != nil {
        c.JSON(http.StatusInternalServerError, APIResponse{
            Success: false,
            Message: "Failed to generate chat response",
            Errors: gin.H{
                "ai_error": "Failed to generate AI response",
                "details": err.Error(),
            },
            Meta: MetaData{
                RequestID: c.GetHeader("X-Request-ID"),
                Timestamp: time.Now().UTC().Format(time.RFC3339),
            },
        })
        return
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Message: "Chat response generated successfully",
        Data:    gin.H{"answer": answer},
        Errors:  nil,
        Meta: MetaData{
            RequestID: c.GetHeader("X-Request-ID"),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        },
    })
}

func generateOpenAIResponse(context string, config *model.UserConfiguration) (string, error) {
    if config.OpenAIAPIKey == "" {
        return "", fmt.Errorf("OpenAI API key not set")
    }

    systemPrompt := config.BasicPrompt
    if systemPrompt == "" {
        systemPrompt = `You are a helpful shopping assistant that provides recommendations based on product information. 
Please respond in Indonesian language. Keep responses concise and natural.
Focus on the relevant product features and benefits that match the user's query.`
    }

    payload := OpenAIRequest{
        Model: config.OpenAIModel,
        Messages: []Message{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: context},
        },
        MaxTokens: config.MaxChatReplyChars,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+config.OpenAIAPIKey)
    fmt.Println("OpenAI API Key:", config.OpenAIAPIKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        var rawBody bytes.Buffer
        rawBody.ReadFrom(resp.Body)
        return "", fmt.Errorf("OpenAI API error: %s", rawBody.String())
    }

    var result OpenAIResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    if len(result.Choices) == 0 {
        return "", fmt.Errorf("no response generated")
    }

    return result.Choices[0].Message.Content, nil
}

func formatProductList(products []SearchResult) string {
    var result strings.Builder
    for i, p := range products {
        result.WriteString(fmt.Sprintf(
            "%d. %s (%s) - Rp %.2f\n   %s\n",
            i+1,
            p.Name,
            p.Category,
            p.Price,
            p.Description,
        ))
    }
    return result.String()
}


// APIResponse represents the standard API response format
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Errors  interface{} `json:"errors"`
    Meta    MetaData    `json:"meta"`
}

// MetaData represents metadata for API responses
type MetaData struct {
    RequestID string    `json:"request_id"`
    Timestamp string    `json:"timestamp"`
}


// CORSMiddleware adds CORS headers to allow all origins
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

// getConfigurationForUser retrieves the configuration for a given user
func getConfigurationForUser(ctx context.Context, userID int64) (*model.UserConfiguration, error) {
	// Get config service instance
	configService, err := service.NewConfigService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config service: %v", err)
	}
	
	// Use the service method that properly decrypts the API key
	return configService.GetConfigurationByUser(ctx, userID)
}
