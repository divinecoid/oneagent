package main

import (
    "github.com/gin-gonic/gin"
    "github.com/divinecoid/oneagent/internal/db"
    "log"
    apiv1 "github.com/divinecoid/oneagent/internal/api/v1"
)

const apiKey = "sk-132d16a281454332ac905f8abcc703bc" // Replace with your actual DeepSeek API key
const apiURL = "https://api.deepseek.com/v1/chat/completions" // Adjust if endpoint differs

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
}

type ChatResponse struct {
    Choices []struct {
        Message Message `json:"message"`
    } `json:"choices"`
}

func main() {
    r := gin.Default()
    r.Use(apiv1.CORSMiddleware()) 
    r.SetTrustedProxies(nil) 
    v1 := r.Group("/api/v1")
    apiv1.RegisterRoutes(v1)
    
    if err := db.Connect(); err != nil {
        log.Fatal("DB connection failed:", err)
    }

    r.Run(":8080")


    // Start the chat loop



    // reader := bufio.NewReader(os.Stdin)
    // messages := []Message{
    //     {Role: "system", Content: "Anda adalah customer service bakmi Air Mangkok. Anda akan membantu customer dengan menjawab pertanyaan mereka terkait bakmi Air Mangkok."},
    // }

    // for {
    //     fmt.Print("You: ")
    //     input, _ := reader.ReadString('\n')
    //     input = strings.TrimSpace(input)

    //     if input == "exit" {
    //         break
    //     }

    //     messages = append(messages, Message{Role: "user", Content: input})

    //     requestBody := ChatRequest{
    //         Model:    "deepseek-chat", // Replace with the correct model name from DeepSeek
    //         Messages: messages,
    //     }

    //     bodyBytes, err := json.Marshal(requestBody)
    //     if err != nil {
    //         fmt.Println("Error encoding JSON:", err)
    //         continue
    //     }

    //     req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
    //     if err != nil {
    //         fmt.Println("Error creating request:", err)
    //         continue
    //     }

    //     req.Header.Set("Authorization", "Bearer "+apiKey)
    //     req.Header.Set("Content-Type", "application/json")

    //     client := &http.Client{}
    //     resp, err := client.Do(req)
    //     if err != nil {
    //         fmt.Println("Error sending request:", err)
    //         continue
    //     }
    //     defer resp.Body.Close()

    //     var chatResp ChatResponse
    //     if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
    //         fmt.Println("Error decoding response:", err)
    //         continue
    //     }

    //     if len(chatResp.Choices) > 0 {
    //         reply := chatResp.Choices[0].Message.Content
    //         fmt.Println("DeepSeek:", reply)
    //         messages = append(messages, Message{Role: "assistant", Content: reply})
    //     } else {
    //         fmt.Println("DeepSeek returned no response.")
    //     }
    // }
}
