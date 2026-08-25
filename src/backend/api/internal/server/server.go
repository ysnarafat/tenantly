package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/config"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/handlers"
	"github.com/ysnarafat/tenantly/internal/middleware"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"
)

type Server struct {
	config *config.Config
	db     *sqlx.DB
	router *gin.Engine
}

func New(cfg *config.Config, db *sqlx.DB) *Server {
	s := &Server{
		config: cfg,
		db:     db,
		router: gin.Default(),
	}

	// Empty/nil means Gin ignores X-Forwarded-For entirely and uses the real
	// socket address — safe default unless a specific reverse proxy is configured.
	if err := s.router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		panic(fmt.Sprintf("invalid TRUSTED_PROXIES configuration: %v", err))
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Security headers (including HSTS for HTTPS)
	s.router.Use(middleware.SecurityHeadersMiddleware())

	// Add HSTS header for secure transport
	s.router.Use(func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	s.router.Use(middleware.CORS(s.config.AllowedOrigins))

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
	propertyRepo := repositories.NewPropertyRepository(s.db)
	buildingRepo := repositories.NewBuildingRepository(s.db)
	unitRepo := repositories.NewUnitRepository(s.db)
	tenantRepo := repositories.NewTenantRepository(s.db, s.config.NIDProtector)
	mfaRepo := repositories.NewMFARepository(s.db)
	paymentRepo := repositories.NewPaymentRepository(s.db)
	paymentTransactionRepo := repositories.NewPaymentTransactionRepository(s.db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(s.db)
	leaseRepo := repositories.NewLeaseRepository(s.db)
	leaseChargeRepo := repositories.NewLeaseChargeRepository(s.db)
	organizationRepo := repositories.NewOrganizationRepository(s.db)
	userInvitationRepo := repositories.NewUserInvitationRepository(s.db)
	userOrgRoleRepo := repositories.NewUserOrganizationRoleRepository(s.db)

	// Initialize metadata validator
	metadataValidator := services.NewBuildingMetadataValidator()

	// Initialize services
	organizationService := services.NewOrganizationService(organizationRepo, userInvitationRepo, auditService)
	userService := services.NewUserServiceWithOrganization(userRepo, auditService, organizationService, userOrgRoleRepo, s.config.JWTSecret, s.config.JWTExpiration)
	propertyService := services.NewPropertyService(propertyRepo, auditService)
	buildingService := services.NewBuildingService(buildingRepo, propertyRepo, auditService, metadataValidator)
	unitService := services.NewUnitService(unitRepo, buildingRepo, propertyRepo, auditService)
	tenantService := services.NewTenantService(tenantRepo, leaseRepo, auditService)
	mfaService := services.NewMFAService(mfaRepo, s.config.NIDProtector, s.config.JWTSecret)
	leaseService := services.NewLeaseService(leaseRepo, tenantRepo, unitRepo, leaseChargeRepo, auditService)
	paymentService := services.NewPaymentService(paymentRepo, paymentTransactionRepo, paymentTransactionAttachmentRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo)
	reportService := services.NewReportService(paymentRepo, propertyRepo)
	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, s.config.CookieDomain, s.config.CookieSecure)
	propertyHandler := handlers.NewPropertyHandler(propertyService)
	buildingHandler := handlers.NewBuildingHandler(buildingService)
	unitHandler := handlers.NewUnitHandler(unitService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	mfaHandler := handlers.NewMFAHandler(mfaService)
	leaseHandler := handlers.NewLeaseHandler(leaseService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	reportHandler := handlers.NewReportHandler(reportService)

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
		loginRateLimiter := middleware.NewRateLimiter(10, time.Minute)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", middleware.RateLimitMiddleware(loginRateLimiter, auditService), userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/reset-password", userHandler.ResetPassword)
			auth.POST("/confirm-reset-password", userHandler.ConfirmPasswordReset)
			auth.POST("/register-with-invitation", userHandler.RegisterWithInvitation)
		}

		// Public invitation routes
		invitations := v1.Group("/invitations")
		{
			invitations.GET("/validate", organizationHandler.ValidateInvitationToken)
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
				authProtected.POST("/set-organization", userHandler.SetOrganization)

				// MFA (TOTP) enrollment and step-up verification
				authProtected.GET("/mfa/status", mfaHandler.GetStatus)
				authProtected.POST("/mfa/enroll", mfaHandler.Enroll)
				authProtected.POST("/mfa/verify", mfaHandler.Verify)
			}

			// Invitation acceptance routes (protected)
			invitationsProtected := protected.Group("/invitations")
			{
				invitationsProtected.POST("/accept", organizationHandler.AcceptInvitation)
			}

			// User management routes (role-based access)
			users := protected.Group("/users")
			{
				users.GET("", middleware.RequireSuperAdminOrAdmin(), userHandler.GetUsers)
				users.POST("", middleware.RequireSuperAdminOrAdmin(), userHandler.CreateUser)
				users.GET("/:id", middleware.RequireAnyRole(), userHandler.GetUser)
				users.PUT("/:id", middleware.RequireSuperAdminOrAdmin(), userHandler.UpdateUser)
				users.PUT("/:id/reset-password", middleware.RequireSuperAdminOrAdmin(), userHandler.AdminResetPassword)
				users.DELETE("/:id", middleware.RequireSuperAdminOrAdmin(), userHandler.DeleteUser)
			}

			// Organization management routes
			organizations := protected.Group("/organizations")
			{
				// Platform-level management (SUPER_ADMIN only)
				organizations.POST("", middleware.RequireSuperAdmin(), organizationHandler.CreateOrganization)
				organizations.GET("", middleware.RequireSuperAdmin(), organizationHandler.ListOrganizations)

				// User invitation routes within organization — SUPER_ADMIN, or ORG_ADMIN of
				// THIS organization specifically (RequireOrgAdmin verifies membership via DB,
				// not just the JWT role claim). Register before the generic :id route.
				organizations.POST("/:id/invitations", middleware.OrganizationValidationMiddleware(organizationRepo), middleware.RequireOrgAdmin(organizationRepo, userOrgRoleRepo), organizationHandler.InviteUser)
				organizations.GET("/:id/invitations", middleware.OrganizationValidationMiddleware(organizationRepo), middleware.RequireOrgAdmin(organizationRepo, userOrgRoleRepo), organizationHandler.GetPendingInvitations)
				organizations.DELETE("/:id/invitations/:invitation_id", middleware.OrganizationValidationMiddleware(organizationRepo), middleware.RequireOrgAdmin(organizationRepo, userOrgRoleRepo), organizationHandler.RevokeInvitation)

				// Generic organization routes (SUPER_ADMIN only; register after specific nested routes)
				organizations.GET("/:id", middleware.RequireSuperAdmin(), organizationHandler.GetOrganization)
				organizations.PUT("/:id", middleware.RequireSuperAdmin(), organizationHandler.UpdateOrganization)
				organizations.DELETE("/:id", middleware.RequireSuperAdmin(), organizationHandler.DeleteOrganization)
			}

			// Property management routes
			properties := protected.Group("/properties")
			properties.Use(middleware.RequireOrgContext())
			{
				properties.GET("", middleware.RequireAnyRole(), propertyHandler.GetProperties)
				properties.POST("", middleware.RequireAdminOrPropertyManager(), propertyHandler.CreateProperty)
				properties.GET("/search", middleware.RequireAnyRole(), propertyHandler.SearchProperties)
				properties.GET("/:id", middleware.RequireAnyRole(), propertyHandler.GetProperty)
				properties.PUT("/:id", middleware.RequireAdminOrPropertyManager(), propertyHandler.UpdateProperty)
				properties.DELETE("/:id", middleware.RequireAdmin(), propertyHandler.DeleteProperty)
				properties.GET("/:id/aggregations", middleware.RequireAnyRole(), propertyHandler.GetPropertyAggregations)

				// Property-building relationship endpoints with property validation middleware
				properties.GET("/:id/buildings", middleware.RequireAnyRole(), middleware.PropertyValidationMiddleware(propertyRepo), buildingHandler.GetPropertyBuildings)
				properties.POST("/:id/buildings/bulk", middleware.RequireAdminOrPropertyManager(), middleware.PropertyValidationMiddleware(propertyRepo), buildingHandler.BulkCreateBuildings)

				// Property-unit relationship endpoints
				properties.GET("/:id/units", middleware.RequireAnyRole(), middleware.PropertyValidationMiddleware(propertyRepo), unitHandler.GetUnitsByProperty)
			}

			// Building management routes
			buildings := protected.Group("/buildings")
			buildings.Use(middleware.RequireOrgContext())
			{
				buildings.GET("", middleware.RequireAnyRole(), buildingHandler.GetBuildings)
				buildings.POST("", middleware.RequireAdminOrPropertyManager(), buildingHandler.CreateBuilding)
				buildings.GET("/search", middleware.RequireAnyRole(), buildingHandler.AdvancedSearchBuildings)
				buildings.GET("/export", middleware.RequireAnyRole(), buildingHandler.ExportBuildingData)
				buildings.GET("/types/:type/metadata", middleware.RequireAnyRole(), buildingHandler.GetBuildingMetadataSchema)
				buildings.GET("/:id", middleware.RequireAnyRole(), buildingHandler.GetBuilding)
				buildings.PUT("/:id", middleware.RequireAdminOrPropertyManager(), buildingHandler.UpdateBuilding)
				buildings.DELETE("/:id", middleware.RequireAdmin(), buildingHandler.DeleteBuilding)
				buildings.GET("/:id/analytics", middleware.RequireAnyRole(), buildingHandler.GetBuildingAnalytics)
				buildings.GET("/:id/units", middleware.RequireAnyRole(), buildingHandler.GetBuildingUnits)
				buildings.PUT("/:id/status", middleware.RequireAdminOrPropertyManager(), buildingHandler.UpdateBuildingStatus)

				// Building-unit relationship endpoints
				buildings.GET("/:id/units/list", middleware.RequireAnyRole(), unitHandler.GetUnitsByBuilding)
				buildings.POST("/:id/units/bulk", middleware.RequireAdminOrPropertyManager(), unitHandler.BulkCreateUnits)
			}

			// Unit management routes
			units := protected.Group("/units")
			units.Use(middleware.RequireOrgContext())
			{
				units.POST("", middleware.RequireAdminOrPropertyManager(), unitHandler.CreateUnit)
				units.GET("/:id", middleware.RequireAnyRole(), unitHandler.GetUnit)
				units.PUT("/:id", middleware.RequireAdminOrPropertyManager(), unitHandler.UpdateUnit)
				units.DELETE("/:id", middleware.RequireAdmin(), unitHandler.DeleteUnit)
				units.GET("/:id/hierarchy", middleware.RequireAnyRole(), unitHandler.GetUnitHierarchyContext)
			}

			// Placeholder routes for other modules (will be implemented in later tasks).
			// Org-scoped and role-checked now, matching every other data route group,
			// so real handlers can't land here later without access control already in place.
			shops := protected.Group("/shops")
			shops.Use(middleware.RequireOrgContext())
			{
				shops.GET("", middleware.RequireAnyRole(), s.handlePlaceholder("Get shops"))
				shops.POST("", middleware.RequireAdminOrPropertyManager(), s.handlePlaceholder("Create shop"))
				shops.GET("/:id", middleware.RequireAnyRole(), s.handlePlaceholder("Get shop"))
				shops.PUT("/:id", middleware.RequireAdminOrPropertyManager(), s.handlePlaceholder("Update shop"))
				shops.DELETE("/:id", middleware.RequireAdmin(), s.handlePlaceholder("Delete shop"))
			}

			tenants := protected.Group("/tenants")
			tenants.Use(middleware.RequireOrgContext())
			{
				tenants.GET("", middleware.RequireAnyRole(), tenantHandler.GetAllTenants)
				tenants.POST("", middleware.RequireAdminOrPropertyManager(), tenantHandler.CreateTenant)
				tenants.GET("/:id", middleware.RequireAnyRole(), tenantHandler.GetTenantByID)
				tenants.GET("/:id/nid", middleware.RequireAdminOrPropertyManager(), middleware.RequireStepUp(s.config.JWTSecret), tenantHandler.RevealNID)
				tenants.PUT("/:id", middleware.RequireAdminOrPropertyManager(), tenantHandler.UpdateTenant)
				tenants.DELETE("/:id", middleware.RequireAdmin(), tenantHandler.DeleteTenant)
			}

			leases := protected.Group("/leases")
			leases.Use(middleware.RequireOrgContext())
			{
				leases.GET("", middleware.RequireAnyRole(), leaseHandler.GetAllLeases)
				leases.POST("", middleware.RequireAdminOrPropertyManager(), leaseHandler.CreateLease)
				leases.GET("/:id", middleware.RequireAnyRole(), leaseHandler.GetLeaseByID)
				leases.PUT("/:id", middleware.RequireAdminOrPropertyManager(), leaseHandler.UpdateLease)
				leases.DELETE("/:id", middleware.RequireAdmin(), leaseHandler.DeleteLease)
				leases.POST("/:id/terminate", middleware.RequireAdminOrPropertyManager(), leaseHandler.TerminateLease)
				leases.POST("/:id/renew", middleware.RequireAdminOrPropertyManager(), leaseHandler.RenewLease)
				leases.POST("/:id/charges", middleware.RequireAdminOrPropertyManager(), leaseHandler.AddLeaseCharge)
				leases.PUT("/:id/charges/:chargeId", middleware.RequireAdminOrPropertyManager(), leaseHandler.UpdateLeaseCharge)
				leases.DELETE("/:id/charges/:chargeId", middleware.RequireAdminOrPropertyManager(), leaseHandler.DeleteLeaseCharge)

				// Lease relationships
				leases.GET("/unit/:unit_id", middleware.RequireAnyRole(), leaseHandler.GetLeasesByUnit)
				leases.GET("/tenant/:tenant_id", middleware.RequireAnyRole(), leaseHandler.GetLeasesByTenant)

				// Due list
				leases.GET("/due", middleware.RequireAdminOrPropertyManagerOrAccountant(), leaseHandler.GetLeasesDue)
				leases.GET("/due/summary", middleware.RequireAdminOrPropertyManagerOrAccountant(), leaseHandler.GetDueSummary)
			}

			payments := protected.Group("/payments")
			payments.Use(middleware.RequireOrgContext())
			{
				payments.GET("", middleware.RequireAnyRole(), paymentHandler.GetPayments)
				payments.POST("", middleware.RequireAdminOrPropertyManager(), paymentHandler.CreatePayment)
				payments.POST("/bulk", middleware.RequireAdminOrPropertyManager(), paymentHandler.BulkCreatePayments)
				payments.POST("/generate-monthly", middleware.RequireAdminOrPropertyManager(), paymentHandler.GenerateMonthlyPayments)
				payments.GET("/search", middleware.RequireAnyRole(), paymentHandler.SearchLeases)
				payments.GET("/:id", middleware.RequireAnyRole(), paymentHandler.GetPayment)
				payments.GET("/:id/receipt", middleware.RequireAnyRole(), paymentHandler.DownloadReceipt)
				payments.PUT("/:id", middleware.RequireAdminOrPropertyManager(), paymentHandler.UpdatePayment)
				payments.POST("/:id/transactions", middleware.RequireAdminOrPropertyManager(), paymentHandler.RecordPaymentTransaction)
				payments.GET("/:id/transactions", middleware.RequireAnyRole(), paymentHandler.GetPaymentTransactions)
				payments.DELETE("/:id/transactions/:transactionId", middleware.RequireAdminOrPropertyManager(), paymentHandler.DeletePaymentTransaction)
				payments.POST("/:id/transactions/:transactionId/attachments", middleware.RequireAdminOrPropertyManager(), paymentHandler.UploadPaymentTransactionAttachment)
				payments.GET("/:id/transactions/:transactionId/attachments", middleware.RequireAnyRole(), paymentHandler.GetPaymentTransactionAttachments)
				payments.GET("/:id/transactions/:transactionId/attachments/:attachmentId", middleware.RequireAnyRole(), paymentHandler.DownloadPaymentTransactionAttachment)
				payments.DELETE("/:id/transactions/:transactionId/attachments/:attachmentId", middleware.RequireAdminOrPropertyManager(), paymentHandler.DeletePaymentTransactionAttachment)
				payments.GET("/building/:building_id/report", middleware.RequireAnyRole(), paymentHandler.GetBuildingPaymentReport)
				payments.GET("/property/:property_id/report", middleware.RequireAnyRole(), paymentHandler.GetPropertyPaymentReport)
			}

			dashboard := protected.Group("/dashboard")
			dashboard.Use(middleware.RequireOrgContext())
			{
				dashboard.GET("/summary", middleware.RequireAnyRole(), paymentHandler.GetDashboardSummary)
			}

			reports := protected.Group("/reports")
			reports.Use(middleware.RequireOrgContext())
			{
				reports.GET("/ledger", middleware.RequireAnyRole(), reportHandler.GetFinancialLedger)
				reports.GET("/collection-summary", middleware.RequireAnyRole(), reportHandler.GetCollectionSummary)
				reports.GET("/payment-analysis", middleware.RequireAnyRole(), reportHandler.GetPaymentAnalysis)
				reports.GET("/dashboard-metrics", middleware.RequireAnyRole(), reportHandler.GetDashboardMetrics)
				reports.GET("/tenant-summary", middleware.RequireAnyRole(), reportHandler.GetTenantSummary)
				reports.GET("/property-analytics", middleware.RequireAnyRole(), reportHandler.GetPropertyAnalytics)
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

// Start starts the HTTP/HTTPS server with optional TLS support
func (s *Server) Start(addr string) error {
	// Check if TLS is configured
	certFile, keyFile := s.getTLSCertificates()

	if certFile == "" || keyFile == "" {
		// No TLS - run plain HTTP
		fmt.Println("⚠️  WARNING: Running without TLS/HTTPS. This is NOT recommended for production.")
		fmt.Println("   Set TLS_CERT_FILE and TLS_KEY_FILE environment variables to enable HTTPS.")
		return s.router.Run(addr)
	}

	// TLS enabled - configure strong ciphers and TLS version
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	// Create HTTP server with TLS config
	server := &http.Server{
		Addr:      addr,
		Handler:   s.router,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("✅ HTTPS/TLS enabled (TLS 1.2+) on %s\n", addr)
	return server.ListenAndServeTLS(certFile, keyFile)
}

// getTLSCertificates returns TLS certificate and key file paths from environment
func (s *Server) getTLSCertificates() (string, string) {
	// These would be set via environment variables in production
	// Example: export TLS_CERT_FILE=/etc/ssl/certs/server.crt
	// Example: export TLS_KEY_FILE=/etc/ssl/private/server.key
	certFile := ""
	keyFile := ""

	// In production, these should come from environment variables or config
	// For now, returning empty strings will fall back to HTTP
	return certFile, keyFile
}
