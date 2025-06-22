package v1

import (
	"net/http"
	"strconv"
	"time"
	"fmt"
	"github.com/divinecoid/oneagent/internal/model"
	"github.com/divinecoid/oneagent/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

type CreateConfigRequest struct {
	Name                  string    `json:"name" binding:"required"`
	OpenAIAPIKey          string    `json:"openai_api_key" binding:"required"`
	WhatsappToken         string    `json:"whatsapp_token" binding:"required"`
	WhatsappNumber        string    `json:"whatsapp_number" binding:"required"`
	BasicPrompt           string    `json:"basic_prompt" binding:"required"`
	MaxChatReplyCount     int       `json:"max_chat_reply_count" binding:"required"`
	MaxChatReplyChars     int       `json:"max_chat_reply_chars" binding:"required"`
	OpenAIAPIKeyExpires   time.Time `json:"openai_api_key_expires,omitempty"`
	WhatsappTokenExpires  time.Time `json:"whatsapp_token_expires,omitempty"`
	OpenAIModel           string    `json:"openai_model,omitempty"`
	OpenAIEmbeddingModel  string    `json:"openai_embedding_model,omitempty"`
}

type UpdateConfigRequest struct {
	Name                  string    `json:"name" binding:"required"`
	OpenAIAPIKey          string    `json:"openai_api_key,omitempty"`
	WhatsappToken         string    `json:"whatsapp_token,omitempty"`
	WhatsappNumber        string    `json:"whatsapp_number" binding:"required"`
	BasicPrompt           string    `json:"basic_prompt" binding:"required"`
	MaxChatReplyCount     int       `json:"max_chat_reply_count" binding:"required"`
	MaxChatReplyChars     int       `json:"max_chat_reply_chars" binding:"required"`
	OpenAIAPIKeyExpires   time.Time `json:"openai_api_key_expires,omitempty"`
	WhatsappTokenExpires  time.Time `json:"whatsapp_token_expires,omitempty"`
	OpenAIModel           string    `json:"openai_model,omitempty"`
	OpenAIEmbeddingModel  string    `json:"openai_embedding_model,omitempty"`
}

func (h *ConfigHandler) CreateConfiguration(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Errors: gin.H{"validation_error": err.Error()},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	// Validate API key format
	if len(req.OpenAIAPIKey) < 20 || len(req.OpenAIAPIKey) > 208 || len(req.OpenAIAPIKey) == 0 || req.OpenAIAPIKey[:3] != "sk-" {
		fmt.Println("Invalid OpenAI API key format", req.OpenAIAPIKey)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid OpenAI API key format",
			Errors: gin.H{
				"validation_error": fmt.Sprintf("API key must start with sk- and be at least 20 characters. Provided: %q", req.OpenAIAPIKey[:3]),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	userID := c.MustGet("user_id").(int64)

	config := &model.UserConfiguration{
		UserID:                userID,
		Name:                  req.Name,
		OpenAIAPIKey:          req.OpenAIAPIKey,
		WhatsappToken:         req.WhatsappToken,
		WhatsappNumber:        req.WhatsappNumber,
		BasicPrompt:           req.BasicPrompt,
		MaxChatReplyCount:     req.MaxChatReplyCount,
		MaxChatReplyChars:     req.MaxChatReplyChars,
		OpenAIAPIKeyExpires:   req.OpenAIAPIKeyExpires,
		WhatsappTokenExpires:  req.WhatsappTokenExpires,
		OpenAIModel:           req.OpenAIModel,
		OpenAIEmbeddingModel:  req.OpenAIEmbeddingModel,
		CreatedBy:             userID,
		UpdatedBy:             userID,
	}

	err := h.configService.CreateConfiguration(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create configuration",
			Errors: gin.H{"error": err.Error()},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Configuration created successfully",
		Data:    config,
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *ConfigHandler) GetConfiguration(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid configuration ID",
			Errors: gin.H{"error": "configuration ID must be a number"},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	config, err := h.configService.GetConfiguration(c.Request.Context(), configID)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Configuration not found",
			Errors: gin.H{"error": err.Error()},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Configuration retrieved successfully",
		Data:    config,
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *ConfigHandler) UpdateConfiguration(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid configuration ID",
			Errors: gin.H{"error": "configuration ID must be a number"},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Errors: gin.H{"validation_error": err.Error()},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	if req.OpenAIAPIKey != "" && (len(req.OpenAIAPIKey) < 20 || len(req.OpenAIAPIKey) > 128 || req.OpenAIAPIKey[:3] != "sk-") {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid OpenAI API key format",
			Errors: gin.H{"validation_error": "API key must start with sk- and be at least 20 characters"},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	userID := c.MustGet("user_id").(int64)

	config := &model.UserConfiguration{
		ID:                    configID,
		Name:                  req.Name,
		OpenAIAPIKey:          req.OpenAIAPIKey,
		WhatsappToken:         req.WhatsappToken,
		WhatsappNumber:        req.WhatsappNumber,
		BasicPrompt:           req.BasicPrompt,
		MaxChatReplyCount:     req.MaxChatReplyCount,
		MaxChatReplyChars:     req.MaxChatReplyChars,
		OpenAIAPIKeyExpires:   req.OpenAIAPIKeyExpires,
		WhatsappTokenExpires:  req.WhatsappTokenExpires,
		OpenAIModel:           req.OpenAIModel,
		OpenAIEmbeddingModel:  req.OpenAIEmbeddingModel,
		UpdatedBy:             userID,
	}

	err = h.configService.UpdateConfiguration(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update configuration",
			Errors: gin.H{"error": err.Error()},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Configuration updated successfully",
		Data:    config,
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *ConfigHandler) DeleteConfiguration(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid configuration ID",
			Errors: gin.H{"error": "configuration ID must be a number"},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	// Get user ID from session
	userID := c.MustGet("user_id").(int64)

	err = h.configService.DeleteConfiguration(c.Request.Context(), configID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete configuration",
			Errors: gin.H{"error": err.Error()},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Configuration deleted successfully",
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}