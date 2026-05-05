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
	s.router.Use(middleware.CORS(s.config.Environment))

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
	tenantRepo := repositories.NewTenantRepository(s.db)
	paymentRepo := repositories.NewPaymentRepository(s.db)
	leaseRepo := repositories.NewLeaseRepository(s.db)
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
	leaseService := services.NewLeaseService(leaseRepo, tenantRepo, unitRepo, auditService)
	paymentService := services.NewPaymentService(paymentRepo, unitRepo, buildingRepo, propertyRepo, auditService)
	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	propertyHandler := handlers.NewPropertyHandler(propertyService)
	buildingHandler := handlers.NewBuildingHandler(buildingService)
	unitHandler := handlers.NewUnitHandler(unitService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	leaseHandler := handlers.NewLeaseHandler(leaseService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)

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
				users.DELETE("/:id", middleware.RequireSuperAdminOrAdmin(), userHandler.DeleteUser)
			}

			// Organization management routes (SUPER_ADMIN only)
			organizations := protected.Group("/organizations")
			organizations.Use(middleware.RequireSuperAdmin())
			{
				organizations.POST("", organizationHandler.CreateOrganization)
				organizations.GET("", organizationHandler.ListOrganizations)

				// User invitation routes within organization (register before generic :id route)
				organizations.POST("/:id/invitations", middleware.OrganizationValidationMiddleware(organizationRepo), middleware.RequireOrgAdmin(organizationRepo), organizationHandler.InviteUser)
				organizations.GET("/:id/invitations", middleware.OrganizationValidationMiddleware(organizationRepo), middleware.RequireOrgAdmin(organizationRepo), organizationHandler.GetPendingInvitations)
				organizations.DELETE("/:id/invitations/:invitation_id", middleware.RequireOrgAdmin(organizationRepo), organizationHandler.RevokeInvitation)

				// Generic organization routes (register after specific nested routes)
				organizations.GET("/:id", organizationHandler.GetOrganization)
				organizations.PUT("/:id", organizationHandler.UpdateOrganization)
				organizations.DELETE("/:id", organizationHandler.DeleteOrganization)
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
			tenants.Use(middleware.RequireOrgContext())
			{
				tenants.GET("", middleware.RequireAnyRole(), tenantHandler.GetAllTenants)
				tenants.POST("", middleware.RequireAdminOrPropertyManager(), tenantHandler.CreateTenant)
				tenants.GET("/:id", middleware.RequireAnyRole(), tenantHandler.GetTenantByID)
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
				payments.GET("/:id", middleware.RequireAnyRole(), paymentHandler.GetPayment)
				payments.PUT("/:id", middleware.RequireAdminOrPropertyManager(), paymentHandler.UpdatePayment)
				payments.GET("/building/:building_id/report", middleware.RequireAnyRole(), paymentHandler.GetBuildingPaymentReport)
				payments.GET("/property/:property_id/report", middleware.RequireAnyRole(), paymentHandler.GetPropertyPaymentReport)
			}

			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/summary", middleware.RequireAnyRole(), paymentHandler.GetDashboardSummary)
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
