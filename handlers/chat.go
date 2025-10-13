package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go-br-finance-api/config"
	"go-br-finance-api/models"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatRequest struct {
	Model       string          `json:"model"`
	Messages    []OllamaMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	Temperature float64         `json:"temperature"`
}

type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
}

// WebSearch performs a web search using a simple API (mock for demo)
func WebSearch(query string) string {
	// Mock web search - in production, use a real API like SerpAPI or Google Custom Search
	// For demo, return a fixed response
	return fmt.Sprintf("Web search results for '%s': [Mock] Latest financial news includes market trends, investment tips, etc.", query)
}

// ChatWithOllama godoc
// @Summary Chat with financial AI model
// @Description Send a message to the Ollama financial model and get a streamed response, with conversation saved in Redis
// @Tags chat
// @Accept  json
// @Produce  text/event-stream
// @Param request body ChatRequest true "Chat request"
// @Success 200 {string} string "Streamed response"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /chat [post]
func ChatWithOllama(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	ctx := context.Background()
	sessionKey := "chat:" + req.SessionID

	// Retrieve conversation from Redis (if available)
	var conversation models.Conversation
	if config.RedisClient != nil {
		conversationJSON, err := config.RedisClient.Get(ctx, sessionKey).Result()
		if err == nil {
			// Parse existing conversation
			json.Unmarshal([]byte(conversationJSON), &conversation)
		} else {
			// New conversation
			conversation = models.Conversation{
				SessionID: req.SessionID,
				Messages:  []models.Message{},
			}
		}
	} else {
		// No Redis, new conversation each time
		conversation = models.Conversation{
			SessionID: req.SessionID,
			Messages:  []models.Message{},
		}
	}

	// Append user message
	conversation.Messages = append(conversation.Messages, models.Message{
		Role:    "user",
		Content: req.Message,
	})

	// Check if web search is needed
	if strings.Contains(strings.ToLower(req.Message), "search") || strings.Contains(strings.ToLower(req.Message), "web") {
		searchResults := WebSearch(req.Message)
		// Append search results to the last user message
		conversation.Messages[len(conversation.Messages)-1].Content += "\n\nWeb Search Results:\n" + searchResults
	}

	// Prepare messages for Ollama
	var ollamaMessages []OllamaMessage
	// Add system message
	ollamaMessages = append(ollamaMessages, OllamaMessage{
		Role:    "system",
		Content: "Você é um consultor financeiro brasileiro. Forneça conselhos em português brasileiro, pesquisando na web as melhores informações para investir. Use fontes confiáveis e mostre seu raciocínio passo a passo. Mantenha as respostas concisas, com no máximo 300 caracteres. Responda sempre em português brasileiro, nunca em inglês.",
	})
	for _, msg := range conversation.Messages {
		ollamaMessages = append(ollamaMessages, OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Call Ollama API via HTTP with streaming
	ollamaReq := OllamaChatRequest{
		Model:       "gpt-oss:20b-cloud",
		Messages:    ollamaMessages,
		Stream:      true,
		Temperature: 0.1,
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	ollamaAPIKey := os.Getenv("OLLAMA_API_KEY")
	reqBody, _ := json.Marshal(ollamaReq)
	httpReq, err := http.NewRequest("POST", ollamaURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if ollamaAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+ollamaAPIKey)
	}
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
		return
	}
	defer resp.Body.Close()

	// Set up SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var chunk OllamaChatResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue // Skip invalid lines
		}

		content := chunk.Message.Content
		fullResponse.WriteString(content)

		// Send the content as is for real streaming
		event := fmt.Sprintf("data: %s\n\n", content)
		c.Writer.WriteString(event)
		c.Writer.Flush()
	}

	if err := scanner.Err(); err != nil {
		// Handle error, but since streaming, maybe just log
	}

	// Append full assistant response to conversation
	conversation.Messages = append(conversation.Messages, models.Message{
		Role:    "assistant",
		Content: fullResponse.String(),
	})

	// Save conversation back to Redis (if available)
	if config.RedisClient != nil {
		conversationJSONBytes, _ := json.Marshal(conversation)
		config.RedisClient.Set(ctx, sessionKey, string(conversationJSONBytes), 0) // No expiration
	}

	// End the stream
	c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
}

// GetChat godoc
// @Summary Get chat conversation
// @Description Retrieve the conversation for a session from Redis
// @Tags chat
// @Accept  json
// @Produce  json
// @Param session_id query string true "Session ID"
// @Success 200 {object} map[string][]models.Message
// @Failure 400 {object} map[string]string
// @Router /chat [get]
func GetChat(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	if config.RedisClient == nil {
		c.JSON(http.StatusOK, gin.H{"messages": []models.Message{}})
		return
	}

	ctx := context.Background()
	sessionKey := "chat:" + sessionID

	conversationJSON, err := config.RedisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"messages": []models.Message{}})
		return
	}

	var conversation models.Conversation
	json.Unmarshal([]byte(conversationJSON), &conversation)

	c.JSON(http.StatusOK, gin.H{"messages": conversation.Messages})
}

// DeleteChat godoc
// @Summary Delete chat conversation
// @Description Delete the conversation for a session from Redis
// @Tags chat
// @Accept  json
// @Produce  json
// @Param session_id query string true "Session ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /chat [delete]
func DeleteChat(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	ctx := context.Background()
	sessionKey := "chat:" + sessionID

	if config.RedisClient != nil {
		config.RedisClient.Del(ctx, sessionKey)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}
