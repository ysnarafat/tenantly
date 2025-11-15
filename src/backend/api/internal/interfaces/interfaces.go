package interfaces

import (
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id int) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(id int, updates map[string]interface{}) error
	Delete(id int) error
	CreateResetToken(token *models.ResetPasswordToken) error
	GetResetToken(token string) (*models.ResetPasswordToken, error)
	MarkResetTokenUsed(tokenID int) error
	CleanupExpiredTokens() error
}

// AuditServiceInterface defines the interface for audit service operations
type AuditServiceInterface interface {
	LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues interface{}) error
	LogSystemAction(action, tableName string, recordID *int, oldValues, newValues interface{}) error
}

// BuildingRepositoryInterface defines the interface for building repository operations
type BuildingRepositoryInterface interface {
	Create(building *models.Building) error
	GetByID(id int) (*models.Building, error)
	GetByPropertyID(propertyID int) ([]*models.Building, error)
	GetByPropertyAndCode(propertyID int, code string) (*models.Building, error)
	Update(id int, updates map[string]interface{}) error
	SoftDelete(id int) error
	GetWithStats(id int) (*models.BuildingWithStats, error)
	BulkCreate(buildings []*models.Building) error
	Search(filters *models.BuildingSearchFilters) ([]*models.Building, error)
	GetAnalytics(id int) (*models.BuildingAnalytics, error)

	// Enhanced property-building relationship operations
	CountByProperty(propertyID int, filters *models.BuildingSearchFilters) (int, error)
	GetByPropertyWithSorting(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string) ([]*models.Building, error)

	// Advanced search and filtering
	AdvancedSearch(req *models.BuildingSearchRequest) ([]*models.Building, int, error)
	GetBuildingUnits(buildingID int, offset, limit int) ([]*models.BuildingUnitSummary, int, error)
}

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	CreateUser(req *models.CreateUserRequest) (*models.User, error)
	Login(req *models.LoginRequest, clientIP, userAgent string) (*models.LoginResponse, error)
	RefreshToken(refreshToken string) (*models.LoginResponse, error)
	Logout(userID int, clientIP, userAgent string) error
	ChangePassword(userID int, currentPassword, newPassword string) error
	ResetPassword(email string) error
	ConfirmPasswordReset(token, newPassword string) error
	GetUserByID(id int) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(id int, req *models.UpdateUserRequest) error
	DeleteUser(id int) error
}

// MetadataValidatorInterface defines the interface for building metadata validation
type MetadataValidatorInterface interface {
	ValidateMetadata(buildingType models.BuildingType, metadata models.BuildingMetadata) error
	ValidateResidentialMetadata(metadata models.BuildingMetadata) error
	ValidateCommercialMetadata(metadata models.BuildingMetadata) error
	ValidateMixedMetadata(metadata models.BuildingMetadata) error
	GetMetadataSchema(buildingType models.BuildingType) map[string]interface{}
}

// BuildingValidationServiceInterface defines the interface for building validation operations
type BuildingValidationServiceInterface interface {
	// Core validation methods
	ValidateBuildingType(buildingType models.BuildingType) error
	ValidateBuildingCodeUniqueness(propertyID int, code string, excludeID *int) error
	ValidatePropertyAssociation(propertyID int) error
	ValidateMetadataSchema(buildingType models.BuildingType, metadata models.BuildingMetadata) error
	ValidateDeletionConstraints(buildingID int) error
	ValidateConstructionYear(year *int) error
	ValidateFloorCountAndElevatorRequirement(totalFloors int, hasElevator bool) error

	// Business rules validation
	ValidateBuildingCreation(req *models.CreateBuildingRequest) error
	ValidateBuildingUpdate(id int, req *models.UpdateBuildingRequest) error
	ValidateBuildingDeletion(id int) error

	// Advanced validation methods
	ValidateFloorRange(floorRange string) error
	ValidateBusinessHours(businessHours string) error
	ValidateElevatorRequirement(totalFloors int, hasElevator bool, buildingType models.BuildingType) error
	ValidateMetadataConsistency(buildingType models.BuildingType, metadata models.BuildingMetadata) error
}

// BuildingServiceInterface defines the interface for building service operations
type BuildingServiceInterface interface {
	// Core CRUD operations
	CreateBuilding(req *models.CreateBuildingRequest) (*models.Building, error)
	GetBuilding(id int) (*models.Building, error)
	GetBuildingWithStats(id int) (*models.BuildingWithStats, error)
	UpdateBuilding(id int, req *models.UpdateBuildingRequest) (*models.Building, error)
	DeleteBuilding(id int) error

	// Property-building relationship operations
	GetBuildingsByProperty(propertyID int) ([]*models.Building, error)
	GetBuildingsByPropertyWithStats(propertyID int) ([]*models.BuildingWithStats, error)
	GetBuildingByPropertyAndCode(propertyID int, code string) (*models.Building, error)

	// Bulk operations
	BulkCreateBuildings(req *models.BulkCreateBuildingsRequest) ([]*models.Building, error)

	// Search and filtering
	SearchBuildings(filters *models.BuildingSearchFilters) ([]*models.Building, error)

	// Building-level aggregation and analytics
	GetBuildingAnalytics(id int) (*models.BuildingAnalytics, error)
	GetPropertyBuildingAnalytics(propertyID int) ([]*models.BuildingAnalytics, error)
	CalculateBuildingOccupancyRate(id int) (float64, error)
	CalculateBuildingRevenue(id int) (float64, error)
	GetBuildingUnitCounts(id int) (total int, occupied int, vacant int, err error)

	// Enhanced property-building relationship operations
	GetPropertyBuildingsWithPagination(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string, includeStats bool) (*models.BuildingListResponse, error)
	GetPropertyStatistics(propertyID int) (*models.PropertyStatistics, error)
	GetPropertyWithBuildingsHierarchy(propertyID int, includeStats bool) (*models.PropertyWithBuildings, error)

	// Validation and business rules
	ValidateBuildingCreation(req *models.CreateBuildingRequest) error
	ValidateBuildingUpdate(id int, req *models.UpdateBuildingRequest) error
	ValidateBuildingDeletion(id int) error
	ValidateBuildingCodeUniqueness(propertyID int, code string, excludeID *int) error

	// Advanced building API features
	AdvancedSearchBuildings(req *models.BuildingSearchRequest) (*models.BuildingListResponse, error)
	GetBuildingUnits(buildingID int, page, pageSize int) (*models.BuildingUnitsResponse, error)
	GetBuildingMetadataSchema(buildingType models.BuildingType) (*models.MetadataSchemaResponse, error)
	ExportBuildingData(req *models.BuildingExportRequest) ([]byte, string, error)
	UpdateBuildingStatus(buildingID int, req *models.BuildingStatusRequest) (*models.Building, error)
}

// BuildingAnalyticsServiceInterface defines the interface for building analytics operations
type BuildingAnalyticsServiceInterface interface {
	// Core analytics methods
	GetBuildingMetrics(buildingID int) (*models.BuildingMetrics, error)
	GetOccupancyAnalytics(buildingID int) (*models.OccupancyAnalytics, error)
	GetRevenueAnalytics(buildingID int, period string) (*models.RevenueAnalytics, error)
	CompareBuildingPerformance(propertyID int) (*models.BuildingPerformanceComparison, error)

	// Trend analysis and forecasting
	CalculatePerformanceScore(buildingID int) (float64, error)
	GenerateTrendAnalysis(buildingID int, period string) (*models.BuildingTrends, error)
	GenerateOccupancyForecast(buildingID int) (*models.OccupancyForecast, error)
	GenerateRevenueProjections(buildingID int) (*models.RevenueProjections, error)

	// Comparison and benchmarking
	CompareWithPropertyAverage(buildingID int) (*models.BuildingComparisons, error)
	CompareWithTypeAverage(buildingID, propertyID int, buildingType models.BuildingType) (*models.BuildingComparisons, error)
	GetPropertyBuildingRankings(propertyID int) ([]*models.BuildingRanking, error)

	// Caching and performance
	RefreshAnalyticsCache(buildingID int) error
	GetCachedMetrics(buildingID int) (*models.BuildingMetrics, bool)
	InvalidateCache(buildingID int) error
}

// UnitRepositoryInterface defines the interface for unit repository operations
type UnitRepositoryInterface interface {
	Create(req *models.CreateUnitRequest) (*models.Unit, error)
	GetByID(id int) (*models.Unit, error)
	GetByIDWithDetails(id int) (*models.UnitWithDetails, error)
	Update(id int, req *models.UpdateUnitRequest) (*models.Unit, error)
	Delete(id int) error
	CheckUnitNumberExists(buildingID int, unitNumber string, excludeID int) (bool, error)
	HasActiveLeases(unitID int) (bool, error)
	GetByBuildingWithDetails(buildingID int, limit, offset int) ([]*models.UnitWithDetails, int, error)
	GetByPropertyWithDetails(propertyID int, limit, offset int) ([]*models.UnitWithDetails, int, error)
	GetBuildingOccupancyStats(buildingID int, startDate, endDate time.Time) (interface{}, error)
	GetPropertyOccupancyStats(propertyID int, startDate, endDate time.Time) (interface{}, error)
	GetSystemOccupancyStats(startDate, endDate time.Time) (interface{}, error)
	GetBuildingUnitTypeDistribution(buildingID int) (interface{}, error)
}

// UnitServiceInterface defines the interface for unit service operations
type UnitServiceInterface interface {
	CreateUnit(req *models.CreateUnitRequest, userID int) (*models.Unit, error)
	GetUnit(id int) (*models.UnitWithDetails, error)
	UpdateUnit(id int, req *models.UpdateUnitRequest, userID int) (*models.Unit, error)
	DeleteUnit(id int, userID int) error
	ValidateBuildingUnitRelationship(buildingID, propertyID int) error
	ValidateUnitTypeForBuilding(buildingID int, unitType models.UnitType) error
	GetUnitsByBuilding(buildingID int, page, pageSize int) ([]*models.UnitWithDetails, int, error)
	GetUnitsByProperty(propertyID int, page, pageSize int) ([]*models.UnitWithDetails, int, error)
	ValidateHierarchyIntegrity(unitID int) error
	GetUnitHierarchyContext(unitID int) (map[string]interface{}, error)
}

// PaymentRepositoryInterface defines the interface for payment repository operations
type PaymentRepositoryInterface interface {
	Create(req *models.CreatePaymentRequest) (*models.Payment, error)
	GetByID(id int) (*models.Payment, error)
	GetByIDWithDetails(id int) (*models.PaymentWithDetails, error)
	Update(id int, req *models.UpdatePaymentRequest) (*models.Payment, error)
	GetWithDetailsAndFilters(filters map[string]interface{}, limit, offset int) ([]*models.PaymentWithDetails, int, error)
	GetBuildingPaymentStats(buildingID int, startDate, endDate time.Time) (interface{}, error)
	GetPropertyPaymentStats(propertyID int, startDate, endDate time.Time) (interface{}, error)
	GetSystemPaymentStats(startDate, endDate time.Time) (interface{}, error)
	GetBuildingPaymentsInPeriod(buildingID int, startDate, endDate time.Time, limit, offset int) ([]*models.PaymentWithDetails, int, error)
	GetDashboardSummary() (*models.DashboardSummary, error)
	GetBuildingLevelSummary() (map[string]interface{}, error)
	GetBuildingPaymentAnalytics(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentAnalytics, error)
}

// PaymentServiceInterface defines the interface for payment service operations
type PaymentServiceInterface interface {
	CreatePayment(req *models.CreatePaymentRequest, userID int) (*models.Payment, error)
	GetPayment(id int) (*models.PaymentWithDetails, error)
	UpdatePayment(id int, req *models.UpdatePaymentRequest, userID int) (*models.Payment, error)
	GetPaymentsByBuilding(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error)
	GetPaymentsByProperty(propertyID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error)
	GenerateBuildingPaymentReport(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error)
	GeneratePropertyPaymentReport(propertyID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error)
	GetDashboardSummaryWithBuildingContext() (*models.DashboardSummary, error)
	ProcessBulkPayments(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error)
	GetPaymentAnalyticsByBuilding(buildingID int, period string) (*models.BuildingPaymentAnalytics, error)
}

// NotificationRepositoryInterface defines the interface for notification repository operations
type NotificationRepositoryInterface interface {
	Create(req *models.CreateNotificationRequest) (*models.NotificationQueue, error)
	GetByID(id int) (*models.NotificationQueue, error)
	GetByIDWithDetails(id int) (*models.NotificationWithDetails, error)
	Update(id int, req *models.UpdateNotificationRequest) (*models.NotificationQueue, error)
	GetByBuildingWithDetails(buildingID int, limit, offset int) ([]*models.NotificationWithDetails, int, error)
	GetByPropertyWithDetails(propertyID int, limit, offset int) ([]*models.NotificationWithDetails, int, error)
	GetBuildingNotificationStats(buildingID int, startDate, endDate time.Time) (interface{}, error)
	GetPropertyNotificationStats(propertyID int, startDate, endDate time.Time) (interface{}, error)
	GetSystemNotificationStats(startDate, endDate time.Time) (interface{}, error)
}

// NotificationServiceInterface defines the interface for notification service operations
type NotificationServiceInterface interface {
	CreateNotification(req *models.CreateNotificationRequest, userID int) (*models.NotificationQueue, error)
	SendBuildingWideNotification(buildingID int, message string, notificationType string, userID int) ([]*models.NotificationQueue, []error)
	SendPropertyWideNotification(propertyID int, message string, notificationType string, userID int) ([]*models.NotificationQueue, []error)
	GetNotificationsByBuilding(buildingID int, page, pageSize int) ([]*models.NotificationWithDetails, int, error)
	GetNotificationsByProperty(propertyID int, page, pageSize int) ([]*models.NotificationWithDetails, int, error)
	UpdateNotification(id int, req *models.UpdateNotificationRequest, userID int) (*models.NotificationQueue, error)
	GenerateNotificationReport(propertyID *int, buildingID *int, startDate, endDate time.Time) (*models.NotificationReport, error)
}

// TenantRepositoryInterface defines the interface for tenant repository operations
type TenantRepositoryInterface interface {
	GetByID(id int) (*models.Tenant, error)
	GetByUnitID(unitID int) (*models.Tenant, error)
}

// PropertyRepositoryInterface defines the interface for property repository operations
type PropertyRepositoryInterface interface {
	GetByID(id int) (*models.Property, error)
	GetByIDWithStats(id int) (*models.PropertyWithStats, error)
	List(filters map[string]interface{}, limit, offset int) ([]*models.Property, int, error)
	GetBuildingAggregations(propertyID int) (map[string]interface{}, error)
	GetBuildingBreakdowns(propertyID int) (interface{}, error)
	GetBuildingTypeDistribution(propertyID int) (interface{}, error)
	GetBuildingCount(propertyID int) (int, error)
	GetBuildingSummary(propertyID int) (map[string]interface{}, error)
	HasActiveBuildings(propertyID int) (bool, error)
	HasActiveUnits(propertyID int) (bool, error)
}

// ReportingServiceInterface defines the interface for reporting service operations
type ReportingServiceInterface interface {
	GenerateComprehensiveReport(propertyID *int, buildingID *int, startDate, endDate time.Time, userID int) (*models.ComprehensiveReport, error)
	GenerateDashboardReport(filters map[string]interface{}, groupBy string, userID int) (*models.DashboardReport, error)
}
