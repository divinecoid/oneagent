package keys

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	KeyVersionSeparator = "."
	KeyLength          = 32 // 256 bits
	EnvKeyCurrent      = "ENCRYPTION_KEY_CURRENT"
	EnvKeyPrevious     = "ENCRYPTION_KEY_PREVIOUS"
	EnvAppEnvironment  = "APP_ENV"
)

// EncryptionKey represents a versioned encryption key
type EncryptionKey struct {
	Environment string    `json:"environment"`
	Version     string    `json:"version"`
	Key         []byte    `json:"key"`
	CreatedAt   time.Time `json:"created_at"`
}

// KeyManager handles encryption key operations
type KeyManager struct {
	currentKey  *EncryptionKey
	previousKey *EncryptionKey
	environment string
}

// NewKeyManager creates a new key manager instance
func NewKeyManager() (*KeyManager, error) {
	// Get environment
	env := os.Getenv(EnvAppEnvironment)
	if env == "" {
		env = "dev" // Default to development environment
	}

	// Initialize manager
	mgr := &KeyManager{environment: env}

	// Load current key
	currentKeyStr := os.Getenv(EnvKeyCurrent)
	if currentKeyStr == "" {
		return nil, fmt.Errorf("current encryption key not found in environment")
	}

	currentKey, err := parseKeyString(currentKeyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid current key format: %v", err)
	}

	// Validate environment
	if currentKey.Environment != env {
		return nil, fmt.Errorf("key environment mismatch: expected %s, got %s", env, currentKey.Environment)
	}

	mgr.currentKey = currentKey

	// Load previous key if available
	if previousKeyStr := os.Getenv(EnvKeyPrevious); previousKeyStr != "" {
		previousKey, err := parseKeyString(previousKeyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid previous key format: %v", err)
		}
		if previousKey.Environment != env {
			return nil, fmt.Errorf("previous key environment mismatch: expected %s, got %s", env, previousKey.Environment)
		}
		mgr.previousKey = previousKey
	}

	return mgr, nil
}

// GenerateKey creates a new encryption key
func GenerateKey(env, version string) (*EncryptionKey, error) {
	key := make([]byte, KeyLength)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %v", err)
	}

	return &EncryptionKey{
		Environment: env,
		Version:     version,
		Key:         key,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// String formats the key for environment variable storage
func (k *EncryptionKey) String() string {
	keyData := base64.StdEncoding.EncodeToString(k.Key)
	return fmt.Sprintf("%s%s%s%s%s",
		k.Environment,
		KeyVersionSeparator,
		k.Version,
		KeyVersionSeparator,
		keyData,
	)
}

// parseKeyString parses a key string into an EncryptionKey
func parseKeyString(keyStr string) (*EncryptionKey, error) {
	parts := strings.Split(keyStr, KeyVersionSeparator)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid key format")
	}

	keyData, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid key encoding: %v", err)
	}

	if len(keyData) != KeyLength {
		return nil, fmt.Errorf("invalid key length: expected %d bytes, got %d", KeyLength, len(keyData))
	}

	return &EncryptionKey{
		Environment: parts[0],
		Version:     parts[1],
		Key:         keyData,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// GetCurrentKey returns the current encryption key
func (m *KeyManager) GetCurrentKey() *EncryptionKey {
	return m.currentKey
}

// GetPreviousKey returns the previous encryption key if available
func (m *KeyManager) GetPreviousKey() *EncryptionKey {
	return m.previousKey
}

// GetEnvironment returns the current environment
func (m *KeyManager) GetEnvironment() string {
	return m.environment
}