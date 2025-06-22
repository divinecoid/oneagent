package model

type Product struct {
    ID          int64   `json:"id"`
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
    Embedding   []float32 `json:"-"`
}
