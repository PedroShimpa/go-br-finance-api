package models

type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

type Conversation struct {
	SessionID string    `json:"session_id"`
	Messages  []Message `json:"messages"`
}
