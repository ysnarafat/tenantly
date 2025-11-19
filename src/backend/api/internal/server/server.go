package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/config"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/handlers"
	"github.com/ysnarafat/tenantly/internal/middleware"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"
)

type Server struct {
	config *config.Config
	db     *sql.DB
	router *gin.Engine
}

func New(cfg *config.Config, db *sql.DB) *Server {
	s := &Server{
		config: cfg,
		db:     db,
		router: gin.Default(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Security headers
	s.router.Use(middleware.SecurityHeadersMiddleware())

	// CORS is handled by nginx proxy, no need for API-level CORS
	// s.router.Use(middleware.CORS())

	// Rate limiting (5 requests per second per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	auditService := database.NewAuditService(s.db)
	s.router.Use(middleware.RateLimitMiddleware(rateLimiter, auditService))

	// Request logging and audit
	s.router.Use(middleware.AuditMiddleware(auditService))

	// Request logging
	s.router.Use(gin.Logger())

	// Recovery middleware
	s.router.Use(gin.Recovery())
}

func (s *Server) setupRoutes() {
	// Initialize audit service
	auditService := database.NewAuditService(s.db)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(s.db)

	// Initialize services
	userService := services.NewUserService(userRepo, auditService, s.config.JWTSecret, s.config.JWTExpiration)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "tenantly-api",
		})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/reset-password", userHandler.ResetPassword)
			auth.POST("/confirm-reset-password", userHandler.ConfirmPasswordReset)
		}

		// Protected routes
		auditService := database.NewAuditService(s.db)
		protected := v1.Group("/")
		protected.Use(middleware.AuthRequired(s.config.JWTSecret, auditService))
		// Session timeout middleware can be added later if needed
		// protected.Use(middleware.SessionTimeoutMiddleware(auditService))
		{
			// Authentication routes (protected)
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", userHandler.Logout)
				authProtected.POST("/change-password", userHandler.ChangePassword)
			}
			// User management routes (role-based access)
			users := protected.Group("/users")
			{
				users.GET("", middleware.RequireAdminOrPropertyManager(), userHandler.GetUsers)
				users.POST("", middleware.RequireAdmin(), userHandler.CreateUser)
				users.GET("/:id", middleware.RequireAnyRole(), userHandler.GetUser)
				users.PUT("/:id", middleware.RequireAdmin(), userHandler.UpdateUser)
				users.DELETE("/:id", middleware.RequireAdmin(), userHandler.DeleteUser)
			}

			// Placeholder routes for other modules (will be implemented in later tasks)
			shops := protected.Group("/shops")
			{
				shops.GET("", s.handlePlaceholder("Get shops"))
				shops.POST("", s.handlePlaceholder("Create shop"))
				shops.GET("/:id", s.handlePlaceholder("Get shop"))
				shops.PUT("/:id", s.handlePlaceholder("Update shop"))
				shops.DELETE("/:id", s.handlePlaceholder("Delete shop"))
			}

			tenants := protected.Group("/tenants")
			{
				tenants.GET("", s.handlePlaceholder("Get tenants"))
				tenants.POST("", s.handlePlaceholder("Create tenant"))
				tenants.GET("/:id", s.handlePlaceholder("Get tenant"))
				tenants.PUT("/:id", s.handlePlaceholder("Update tenant"))
				tenants.DELETE("/:id", s.handlePlaceholder("Delete tenant"))
			}

			payments := protected.Group("/payments")
			{
				payments.GET("", s.handlePlaceholder("Get payments"))
				payments.POST("", s.handlePlaceholder("Create payment"))
				payments.GET("/:id", s.handlePlaceholder("Get payment"))
				payments.PUT("/:id", s.handlePlaceholder("Update payment"))
			}

			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/summary", s.handlePlaceholder("Dashboard summary"))
			}

			reports := protected.Group("/reports")
			{
				reports.GET("/ledger", s.handlePlaceholder("Ledger report"))
				reports.GET("/export", s.handlePlaceholder("Export report"))
			}
		}
	}
}

func (s *Server) handlePlaceholder(operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"message": operation + " - to be implemented in later tasks",
		})
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
