package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type OpenRouterRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("themes/sub.html"))
	tmpl.Execute(w, nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	welcomeMsg := Message{
		Type:    "bot",
		Content: "Hello! I'm your cybersecurity consultant. How can I help you today?",
	}
	conn.WriteJSON(welcomeMsg)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		response, err := getOpenRouterResponse(msg.Content)
		if err != nil {
			log.Printf("Error getting AI response: %v", err)
			response = "I apologize, but I'm having trouble processing your request right now."
		}

		responseMsg := Message{
			Type:    "bot",
			Content: response,
		}

		err = conn.WriteJSON(responseMsg)
		if err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}

func getOpenRouterResponse(userMessage string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	siteName := os.Getenv("YOUR_SITE_NAME")
	siteURL := os.Getenv("YOUR_SITE_URL")

	requestBody := OpenRouterRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a cybersecurity consultant. Provide clear, accurate, and helpful advice about cybersecurity topics.",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if siteName != "" {
		req.Header.Set("X-Title", siteName)
	}
	if siteURL != "" {
		req.Header.Set("HTTP-Referer", siteURL)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return openRouterResp.Choices[0].Message.Content, nil
}
