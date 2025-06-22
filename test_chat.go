package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080/api/v1"
	
	// Step 1: Login to get a session ID
	loginURL := baseURL + "/auth/login"
	loginBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	
	loginJSON, err := json.Marshal(loginBody)
	if err != nil {
		fmt.Printf("Error marshaling login JSON: %v\n", err)
		return
	}
	
	loginReq, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginJSON))
	if err != nil {
		fmt.Printf("Error creating login request: %v\n", err)
		return
	}
	
	loginReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	loginResp, err := client.Do(loginReq)
	if err != nil {
		fmt.Printf("Error making login request: %v\n", err)
		return
	}
	defer loginResp.Body.Close()
	
	fmt.Printf("Login Response Status: %s\n", loginResp.Status)
	
	var loginResponse map[string]interface{}
	if err := json.NewDecoder(loginResp.Body).Decode(&loginResponse); err != nil {
		fmt.Printf("Error decoding login response: %v\n", err)
		return
	}
	
	loginResponseJSON, _ := json.MarshalIndent(loginResponse, "", "  ")
	fmt.Printf("Login Response: %s\n", string(loginResponseJSON))
	
	// Extract session ID from response
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
	
	// Step 2: Test the chat endpoint with valid session
	chatURL := baseURL + "/products/chat"
	chatBody := map[string]interface{}{
		"question": "What products do you have?",
	}
	
	chatJSON, err := json.Marshal(chatBody)
	if err != nil {
		fmt.Printf("Error marshaling chat JSON: %v\n", err)
		return
	}
	
	chatReq, err := http.NewRequest("POST", chatURL, bytes.NewBuffer(chatJSON))
	if err != nil {
		fmt.Printf("Error creating chat request: %v\n", err)
		return
	}
	
	chatReq.Header.Set("Content-Type", "application/json")
	chatReq.Header.Set("X-Session-ID", sessionID)
	
	chatResp, err := client.Do(chatReq)
	if err != nil {
		fmt.Printf("Error making chat request: %v\n", err)
		return
	}
	defer chatResp.Body.Close()
	
	fmt.Printf("Chat Response Status: %s\n", chatResp.Status)
	
	var chatResponse map[string]interface{}
	if err := json.NewDecoder(chatResp.Body).Decode(&chatResponse); err != nil {
		fmt.Printf("Error decoding chat response: %v\n", err)
		return
	}
	
	chatResponseJSON, _ := json.MarshalIndent(chatResponse, "", "  ")
	fmt.Printf("Chat Response: %s\n", string(chatResponseJSON))
} 