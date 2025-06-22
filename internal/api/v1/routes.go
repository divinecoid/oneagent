package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/divinecoid/oneagent/internal/service"
	"github.com/divinecoid/oneagent/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	// Initialize services
	authService := service.NewAuthService()
	configService, err := service.NewConfigService()
	if err != nil {
		panic(err)
	}
	
	// Initialize handlers
	authHandler := NewAuthHandler(authService)
	configHandler := NewConfigHandler(configService)
	
	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Auth routes
	auth := rg.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authMiddleware.SessionAuth(), authHandler.Logout)
		auth.POST("/password-reset-request", authHandler.RequestPasswordReset)
		auth.POST("/password-reset", authHandler.ResetPassword)
	}

	// Protected routes
	api := rg.Group("/")
	api.Use(authMiddleware.SessionAuth())
	{
		// User routes
		api.GET("/users", authMiddleware.RequireRole("super_admin"), GetUsers)

		// Product routes
		products := api.Group("/products")
		{
			products.POST("/upload", authMiddleware.RequireRole("super_admin", "seller"), UploadProductExcel)
			products.POST("/update-embeddings", authMiddleware.RequireRole("super_admin", "seller"), UpdateProductEmbeddings)
			products.POST("/search", SearchProducts)
			products.POST("/chat", ChatWithProducts)
		}

		// Configuration routes
		configs := api.Group("/configs")
		configs.Use(authMiddleware.RequireRole("super_admin", "admin", "seller"))
		{
			configs.POST("", configHandler.CreateConfiguration)
			configs.GET("/:id", configHandler.GetConfiguration)
			configs.PUT("/:id", configHandler.UpdateConfiguration)
			configs.DELETE("/:id", configHandler.DeleteConfiguration)
		}
	}
}