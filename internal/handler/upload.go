package handler

import (
    "fmt"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
    "context"
    "github.com/divinecoid/oneagent/internal/model"
    "github.com/divinecoid/oneagent/internal/db"
)

func UploadProductExcel(c *gin.Context) {
    // Parse seller_id from multipart form
    sellerIDStr := c.PostForm("seller_id")
    if sellerIDStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "seller_id is required as a form field"})
        return
    }
    sellerID, err := strconv.ParseInt(sellerIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "seller_id must be a valid integer"})
        return
    }

    // Validate seller_id exists in users table with role 'seller'
    var exists bool
    err = db.DB.QueryRow(
        context.Background(),
        "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND role = 'seller')",
        sellerID,
    ).Scan(&exists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate seller_id", "details": err.Error()})
        return
    }
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "seller_id does not exist in users table with role 'seller'"})
        return
    }

    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
        return
    }

    f, err := file.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
        return
    }
    defer f.Close()

    xlsx, err := excelize.OpenReader(f)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file"})
        return
    }

    rows, err := xlsx.GetRows("Sheet1")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read sheet"})
        return
    }

    for i, row := range rows {
        if i == 0 {
            continue // skip header
        }

        if len(row) < 4 {
            continue
        }

        price, err := strconv.ParseFloat(row[2], 64)
        if err != nil {
            continue
        }

        product := model.Product{
            Name:        row[0],
            Category:    row[1],
            Price:       price,
            Description: row[3],
        }

        _, err = db.DB.Exec(context.Background(), `
            INSERT INTO products (name, category, price, description, seller_id)
            VALUES ($1, $2, $3, $4, $5)
        `, product.Name, product.Category, product.Price, product.Description, sellerID)

        if err != nil {
            fmt.Printf("DB error on row %d: %v\n", i, err)
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Upload successful"})
}
