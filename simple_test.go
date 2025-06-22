package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080/api/v1"
	
	// Test 1: Check if server is running
	fmt.Println("=== Testing server connection ===")
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		fmt.Printf("Server not running: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Server is running, status: %s\n", resp.Status)
	
	// Test 2: Register user
	fmt.Println("\n=== Registering user ===")
	registerBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "seller",
	}
	
	registerJSON, _ := json.Marshal(registerBody)
	registerResp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(registerJSON))
	if err != nil {
		fmt.Printf("Error registering: %v\n", err)
		return
	}
	defer registerResp.Body.Close()
	
	var registerResponse map[string]interface{}
	json.NewDecoder(registerResp.Body).Decode(&registerResponse)
	fmt.Printf("Register Status: %s\n", registerResp.Status)
	if registerResponse["success"] == true {
		fmt.Println("User registered successfully")
	} else {
		fmt.Printf("Register failed: %v\n", registerResponse["message"])
	}
} 