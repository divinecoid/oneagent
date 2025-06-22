package v1

import (
	"net/http"
	"time"

	"github.com/divinecoid/oneagent/internal/model"
	"github.com/divinecoid/oneagent/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=8"`
	Role     model.Role  `json:"role" binding:"required,oneof=customer seller"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Errors: gin.H{
				"validation_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Registration failed",
			Errors: gin.H{
				"registration_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Errors: gin.H{
				"validation_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	session, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "Login failed",
			Errors: gin.H{
				"login_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Login successful",
		Data: gin.H{
			"session_id": session.ID,
			"expires_at": session.ExpiresAt,
		},
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "No session ID provided",
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	err := h.authService.Logout(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Logout failed",
			Errors: gin.H{
				"logout_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Logged out successfully",
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Errors: gin.H{
				"validation_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	err := h.authService.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to request password reset",
			Errors: gin.H{
				"reset_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Password reset instructions sent",
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Errors: gin.H{
				"validation_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to reset password",
			Errors: gin.H{
				"reset_error": err.Error(),
			},
			Meta: MetaData{
				RequestID: c.GetHeader("X-Request-ID"),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Password reset successful",
		Meta: MetaData{
			RequestID: c.GetHeader("X-Request-ID"),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}