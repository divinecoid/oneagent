package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/divinecoid/oneagent/internal/db"
	"github.com/divinecoid/oneagent/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(ctx context.Context, email, password string, role model.Role) (*model.User, error) {
	// Check if user exists
	var existingUser model.User
	err := db.DB.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&existingUser.ID)
	if err == nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	user := &model.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert user into database
	err = db.DB.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.Session, error) {
	// Get user
	var user model.User
	err := db.DB.QueryRow(ctx,
		"SELECT id, email, password_hash, role FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate session
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %v", err)
	}

	// Create session data
	sessionData := map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	}

	dataBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session data: %v", err)
	}

	session := &model.Session{
		ID:        sessionID,
		UserID:    user.ID,
		Data:      dataBytes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	// Save session to database
	_, err = db.DB.Exec(ctx,
		"INSERT INTO sessions (id, user_id, data, created_at, expires_at) VALUES ($1, $2, $3, $4, $5)",
		session.ID, session.UserID, session.Data, session.CreatedAt, session.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	_, err := db.DB.Exec(ctx, "DELETE FROM sessions WHERE id = $1", sessionID)
	return err
}

func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	var session model.Session
	err := db.DB.QueryRow(ctx,
		"SELECT id, user_id, data, created_at, expires_at FROM sessions WHERE id = $1 AND expires_at > NOW()",
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.Data, &session.CreatedAt, &session.ExpiresAt)

	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	return &session, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	// Generate reset token
	resetToken, err := generateSessionID() // reusing session ID generator for token
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %v", err)
	}

	// Update user with reset token
	_, err = db.DB.Exec(ctx,
		"UPDATE users SET reset_token = $1, reset_token_expires = $2 WHERE email = $3",
		resetToken,
		time.Now().Add(1*time.Hour), // Token expires in 1 hour
		email,
	)

	if err != nil {
		return fmt.Errorf("failed to set reset token: %v", err)
	}

	// TODO: Send reset token via email
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password and clear reset token
	result, err := db.DB.Exec(ctx,
		"UPDATE users SET password_hash = $1, reset_token = NULL, reset_token_expires = NULL WHERE reset_token = $2 AND reset_token_expires > NOW()",
		string(hashedPassword),
		token,
	)

	if err != nil {
		return fmt.Errorf("failed to reset password: %v", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("invalid or expired reset token")
	}

	return nil
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}