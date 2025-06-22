package model

import (
	"time"
)

type UserConfiguration struct {
	ID                    int64     `json:"id"`
	UserID                int64     `json:"user_id"`
	Name                  string    `json:"name"`
	OpenAIAPIKey          string    `json:"-"` // Encrypted, not exposed in JSON
	WhatsappToken         string    `json:"-"` // Encrypted, not exposed in JSON
	WhatsappNumber        string    `json:"whatsapp_number"`
	BasicPrompt           string    `json:"basic_prompt"`
	MaxChatReplyCount     int       `json:"max_chat_reply_count"`
	MaxChatReplyChars     int       `json:"max_chat_reply_chars"`
	OpenAIAPIKeyExpires   time.Time `json:"openai_api_key_expires,omitempty"`
	WhatsappTokenExpires  time.Time `json:"whatsapp_token_expires,omitempty"`
	OpenAIModel           string    `json:"openai_model"`
	OpenAIEmbeddingModel  string    `json:"openai_embedding_model"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	CreatedBy             int64     `json:"created_by"`
	UpdatedBy             int64     `json:"updated_by"`
}

type ProductKnowledgeData struct {
	ID              int64          `json:"id"`
	ConfigurationID int64          `json:"configuration_id"`
	Data            map[string]any `json:"data"` // Store Excel data as JSON
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CreatedBy       int64          `json:"created_by"`
	UpdatedBy       int64          `json:"updated_by"`
}

type ConfigurationHistory struct {
	ID              int64          `json:"id"`
	ConfigurationID int64          `json:"configuration_id"`
	ChangeType      string         `json:"change_type"` // CREATE, UPDATE, DELETE
	ChangedFields   map[string]any `json:"changed_fields"`
	ChangedAt       time.Time      `json:"changed_at"`
	ChangedBy       int64          `json:"changed_by"`
}

type DeletedConfigLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	ConfigID   int64     `json:"config_id"`
	OldAPIKey  string    `json:"old_api_key"`
	DeletedAt  time.Time `json:"deleted_at"`
	Reason     string    `json:"reason"`
}