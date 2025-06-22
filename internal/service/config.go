package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/divinecoid/oneagent/internal/db"
	"github.com/divinecoid/oneagent/internal/model"
	"github.com/divinecoid/oneagent/pkg/config/keys"
)

type ConfigService struct {
	keyManager *keys.KeyManager
}

func NewConfigService() (*ConfigService, error) {
	keyManager, err := keys.NewKeyManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key manager: %v", err)
	}
	return &ConfigService{keyManager: keyManager}, nil
}

// Encryption helpers
func (s *ConfigService) encrypt(text string) (string, error) {
	key := s.keyManager.GetCurrentKey()
	if key == nil {
		return "", fmt.Errorf("no encryption key available")
	}

	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Include key version in encrypted data
	data := map[string]string{
		"v": key.Version,
		"d": text,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, dataBytes, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *ConfigService) decrypt(encryptedText string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Try current key first
	text, err := s.decryptWithKey(ciphertext, s.keyManager.GetCurrentKey())
	if err == nil {
		return text, nil
	}

	// Try previous key if available
	if prevKey := s.keyManager.GetPreviousKey(); prevKey != nil {
		text, err := s.decryptWithKey(ciphertext, prevKey)
		if err == nil {
			// Re-encrypt with current key
			_, err := s.encrypt(text)
			if err != nil {
				return "", fmt.Errorf("failed to re-encrypt with current key: %v", err)
			}
			return text, nil
		}
	}

	return "", fmt.Errorf("failed to decrypt data with any available key")
}

func (s *ConfigService) decryptWithKey(ciphertext []byte, key *keys.EncryptionKey) (string, error) {
	if key == nil {
		return "", fmt.Errorf("no key provided")
	}

	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	// Parse versioned data
	var data map[string]string
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return "", err
	}

	if data["v"] != key.Version {
		return "", fmt.Errorf("key version mismatch")
	}

	return data["d"], nil
}

// CreateConfiguration creates a new configuration with encrypted tokens
func (s *ConfigService) CreateConfiguration(ctx context.Context, config *model.UserConfiguration) error {
	if config.OpenAIAPIKey != "" {
		encrypted, err := s.encrypt(config.OpenAIAPIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt OpenAI API key: %v", err)
		}
		config.OpenAIAPIKey = encrypted
	}
	if config.WhatsappToken != "" {
		encrypted, err := s.encrypt(config.WhatsappToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt WhatsApp token: %v", err)
		}
		config.WhatsappToken = encrypted
	}
	if config.OpenAIModel == "" {
		config.OpenAIModel = "gpt-4-turbo-preview"
	}
	if config.OpenAIEmbeddingModel == "" {
		config.OpenAIEmbeddingModel = "text-embedding-3-small"
	}
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now
	query := `
		INSERT INTO user_configurations (
			user_id, name, openai_api_key, whatsapp_token, whatsapp_number,
			basic_prompt, max_chat_reply_count, max_chat_reply_chars,
			openai_api_key_expires, whatsapp_token_expires,
			openai_model, openai_embedding_model,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`
	err := db.DB.QueryRow(
		ctx,
		query,
		config.UserID, config.Name, config.OpenAIAPIKey, config.WhatsappToken, config.WhatsappNumber,
		config.BasicPrompt, config.MaxChatReplyCount, config.MaxChatReplyChars,
		config.OpenAIAPIKeyExpires, config.WhatsappTokenExpires,
		config.OpenAIModel, config.OpenAIEmbeddingModel,
		config.CreatedAt, config.UpdatedAt, config.CreatedBy, config.UpdatedBy,
	).Scan(&config.ID)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %v", err)
	}
	return nil
}

// GetConfiguration retrieves a specific configuration by its ID and decrypts sensitive fields.
func (s *ConfigService) GetConfiguration(ctx context.Context, id int64) (*model.UserConfiguration, error) {
	config := &model.UserConfiguration{}
	query := `
		SELECT id, user_id, name, openai_api_key, whatsapp_token, whatsapp_number,
			   basic_prompt, max_chat_reply_count, max_chat_reply_chars,
			   openai_api_key_expires, whatsapp_token_expires,
			   openai_model, openai_embedding_model,
			   created_at, updated_at, created_by, updated_by
		FROM user_configurations
		WHERE id = $1`
	err := db.DB.QueryRow(ctx, query, id).Scan(
		&config.ID, &config.UserID, &config.Name, &config.OpenAIAPIKey, &config.WhatsappToken, &config.WhatsappNumber,
		&config.BasicPrompt, &config.MaxChatReplyCount, &config.MaxChatReplyChars,
		&config.OpenAIAPIKeyExpires, &config.WhatsappTokenExpires,
		&config.OpenAIModel, &config.OpenAIEmbeddingModel,
		&config.CreatedAt, &config.UpdatedAt, &config.CreatedBy, &config.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %v", err)
	}
	if config.OpenAIAPIKey != "" {
		decrypted, err := s.decrypt(config.OpenAIAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt OpenAI API key: %v", err)
		}
		config.OpenAIAPIKey = decrypted
	}
	if config.WhatsappToken != "" {
		decrypted, err := s.decrypt(config.WhatsappToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt WhatsApp token: %v", err)
		}
		config.WhatsappToken = decrypted
	}
	return config, nil
}

// UpdateConfiguration updates an existing configuration with encrypted tokens.
func (s *ConfigService) UpdateConfiguration(ctx context.Context, config *model.UserConfiguration) error {
	if config.OpenAIAPIKey != "" {
		encrypted, err := s.encrypt(config.OpenAIAPIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt OpenAI API key: %v", err)
		}
		config.OpenAIAPIKey = encrypted
	}
	if config.WhatsappToken != "" {
		encrypted, err := s.encrypt(config.WhatsappToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt WhatsApp token: %v", err)
		}
		config.WhatsappToken = encrypted
	}
	config.UpdatedAt = time.Now()
	query := `
		UPDATE user_configurations SET
			name = $1, openai_api_key = $2, whatsapp_token = $3, whatsapp_number = $4,
			basic_prompt = $5, max_chat_reply_count = $6, max_chat_reply_chars = $7,
			openai_api_key_expires = $8, whatsapp_token_expires = $9,
			openai_model = $10, openai_embedding_model = $11,
			updated_at = $12, updated_by = $13
		WHERE id = $14`
	_, err := db.DB.Exec(
		ctx,
		query,
		config.Name, config.OpenAIAPIKey, config.WhatsappToken, config.WhatsappNumber,
		config.BasicPrompt, config.MaxChatReplyCount, config.MaxChatReplyChars,
		config.OpenAIAPIKeyExpires, config.WhatsappTokenExpires,
		config.OpenAIModel, config.OpenAIEmbeddingModel,
		config.UpdatedAt, config.UpdatedBy, config.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update configuration: %v", err)
	}
	return nil
}

// DeleteConfiguration deletes a configuration by its ID.
// The userID is passed for potential audit logging or ownership checks in the future.
func (s *ConfigService) DeleteConfiguration(ctx context.Context, id int64, userID int64) error {
	// Fetch config for logging
	var oldAPIKey string
	err := db.DB.QueryRow(ctx, "SELECT openai_api_key FROM user_configurations WHERE id = $1", id).Scan(&oldAPIKey)
	if err == nil {
		_ = s.logDeletedConfig(ctx, userID, id, oldAPIKey, "Manual deletion via API")
	}
	query := `DELETE FROM user_configurations WHERE id = $1`
	result, err := db.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete configuration: %v", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no configuration found with ID %d to delete, or user %d does not own it", id, userID)
	}
	return nil
}

func (s *ConfigService) logDeletedConfig(ctx context.Context, userID, configID int64, oldAPIKey, reason string) error {
	query := `INSERT INTO deleted_configs_log (user_id, config_id, old_api_key, reason) VALUES ($1, $2, $3, $4)`
	_, err := db.DB.Exec(ctx, query, userID, configID, oldAPIKey, reason)
	return err
}

// GetConfigurationByUser retrieves the most recent configuration for a user and decrypts sensitive fields.
func (s *ConfigService) GetConfigurationByUser(ctx context.Context, userID int64) (*model.UserConfiguration, error) {
	config := &model.UserConfiguration{}
	query := `
		SELECT id, user_id, name, openai_api_key, whatsapp_token, whatsapp_number,
			   basic_prompt, max_chat_reply_count, max_chat_reply_chars,
			   openai_api_key_expires, whatsapp_token_expires,
			   openai_model, openai_embedding_model,
			   created_at, updated_at, created_by, updated_by
		FROM user_configurations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`
	err := db.DB.QueryRow(ctx, query, userID).Scan(
		&config.ID, &config.UserID, &config.Name, &config.OpenAIAPIKey, &config.WhatsappToken, &config.WhatsappNumber,
		&config.BasicPrompt, &config.MaxChatReplyCount, &config.MaxChatReplyChars,
		&config.OpenAIAPIKeyExpires, &config.WhatsappTokenExpires,
		&config.OpenAIModel, &config.OpenAIEmbeddingModel,
		&config.CreatedAt, &config.UpdatedAt, &config.CreatedBy, &config.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user configuration: %v", err)
	}
	
	// Decrypt sensitive fields
	if config.OpenAIAPIKey != "" {
		decrypted, err := s.decrypt(config.OpenAIAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt OpenAI API key: %v", err)
		}
		config.OpenAIAPIKey = decrypted
	}
	if config.WhatsappToken != "" {
		decrypted, err := s.decrypt(config.WhatsappToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt WhatsApp token: %v", err)
		}
		config.WhatsappToken = decrypted
	}
	return config, nil
}
