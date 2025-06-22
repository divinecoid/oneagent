package model

import (
	"time"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleSeller    Role = "seller"
	RoleCustomer  Role = "customer"
)

type User struct {
	ID                int64     `json:"id"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	Role              Role      `json:"role"`
	EmailVerified     bool      `json:"email_verified"`
	EmailVerifyToken  string    `json:"-"`
	ResetToken        string    `json:"-"`
	ResetTokenExpires time.Time `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}