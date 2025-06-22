package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080/api/v1"
	
	// Step 1: Register a user
	fmt.Println("=== Step 1: Registering user ===")
	registerURL := baseURL + "/auth/register"
	registerBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "seller",
	}
	
	registerJSON, _ := json.Marshal(registerBody)
	registerReq, _ := http.NewRequest("POST", registerURL, bytes.NewBuffer(registerJSON))
	registerReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	registerResp, err := client.Do(registerReq)
	if err != nil {
		fmt.Printf("Error registering: %v\n", err)
		return
	}
	defer registerResp.Body.Close()
	
	var registerResponse map[string]interface{}
	json.NewDecoder(registerResp.Body).Decode(&registerResponse)
	fmt.Printf("Register Status: %s\n", registerResp.Status)
	registerJSON, _ = json.MarshalIndent(registerResponse, "", "  ")
	fmt.Printf("Register Response: %s\n", string(registerJSON))
	
	// Step 2: Login
	fmt.Println("\n=== Step 2: Logging in ===")
	loginURL := baseURL + "/auth/login"
	loginBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	
	loginJSON, _ := json.Marshal(loginBody)
	loginReq, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginJSON))
	loginReq.Header.Set("Content-Type", "application/json")
	
	loginResp, err := client.Do(loginReq)
	if err != nil {
		fmt.Printf("Error logging in: %v\n", err)
		return
	}
	defer loginResp.Body.Close()
	
	var loginResponse map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	fmt.Printf("Login Status: %s\n", loginResp.Status)
	loginResponseJSON, _ := json.MarshalIndent(loginResponse, "", "  ")
	fmt.Printf("Login Response: %s\n", string(loginResponseJSON))
	
	// Extract session ID
	data, ok := loginResponse["data"].(map[string]interface{})
	if !ok {
		fmt.Println("No data in login response")
		return
	}
	
	sessionID, ok := data["session_id"].(string)
	if !ok {
		fmt.Println("No session_id in login response")
		return
	}
	
	fmt.Printf("Session ID: %s\n", sessionID)
	
	// Step 3: Create a configuration (required for chat)
	fmt.Println("\n=== Step 3: Creating configuration ===")
	configURL := baseURL + "/configs"
	configBody := map[string]interface{}{
		"name":                    "Test Config",
		"openai_api_key":          "sk-test-token",
		"whatsapp_token":          "whatsapp-token",
		"whatsapp_number":         "+1234567890",
		"basic_prompt":            "You are a helpful assistant.",
		"max_chat_reply_count":    5,
		"max_chat_reply_chars":    1000,
		"openai_model":            "gpt-4-turbo-preview",
		"openai_embedding_model":  "text-embedding-3-small",
	}
	
	configJSON, _ := json.Marshal(configBody)
	configReq, _ := http.NewRequest("POST", configURL, bytes.NewBuffer(configJSON))
	configReq.Header.Set("Content-Type", "application/json")
	configReq.Header.Set("X-Session-ID", sessionID)
	
	configResp, err := client.Do(configReq)
	if err != nil {
		fmt.Printf("Error creating config: %v\n", err)
		return
	}
	defer configResp.Body.Close()
	
	var configResponse map[string]interface{}
	json.NewDecoder(configResp.Body).Decode(&configResponse)
	fmt.Printf("Config Status: %s\n", configResp.Status)
	configResponseJSON, _ := json.MarshalIndent(configResponse, "", "  ")
	fmt.Printf("Config Response: %s\n", string(configResponseJSON))
	
	// Step 4: Test chat endpoint
	fmt.Println("\n=== Step 4: Testing chat endpoint ===")
	chatURL := baseURL + "/products/chat"
	chatBody := map[string]interface{}{
		"question": "What products do you have?",
	}
	
	chatJSON, _ := json.Marshal(chatBody)
	chatReq, _ := http.NewRequest("POST", chatURL, bytes.NewBuffer(chatJSON))
	chatReq.Header.Set("Content-Type", "application/json")
	chatReq.Header.Set("X-Session-ID", sessionID)
	
	chatResp, err := client.Do(chatReq)
	if err != nil {
		fmt.Printf("Error making chat request: %v\n", err)
		return
	}
	defer chatResp.Body.Close()
	
	fmt.Printf("Chat Status: %s\n", chatResp.Status)
	
	var chatResponse map[string]interface{}
	json.NewDecoder(chatResp.Body).Decode(&chatResponse)
	chatResponseJSON, _ := json.MarshalIndent(chatResponse, "", "  ")
	fmt.Printf("Chat Response: %s\n", string(chatResponseJSON))
} 