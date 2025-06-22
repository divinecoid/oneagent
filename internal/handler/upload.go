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
            INSERT INTO products (name, category, price, description)
            VALUES ($1, $2, $3, $4)
        `, product.Name, product.Category, product.Price, product.Description)

        if err != nil {
            fmt.Printf("DB error on row %d: %v\n", i, err)
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Upload successful"})
}
