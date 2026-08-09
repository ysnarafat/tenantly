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
	GetAll(activeOnly bool) ([]*models.User, error)
	GetByOrganizationID(orgID int, activeOnly bool) ([]*models.User, error)
	Update(id int, updates map[string]any) error
	Delete(id int) error
	CreateResetToken(token *models.ResetPasswordToken) error
	GetResetToken(token string) (*models.ResetPasswordToken, error)
	MarkResetTokenUsed(tokenID int) error
	CleanupExpiredTokens() error
}

// AuditServiceInterface defines the interface for audit service operations
type AuditServiceInterface interface {
	LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues any) error
	LogSystemAction(action, tableName string, recordID *int, oldValues, newValues any) error
}

// OrganizationRepositoryInterface defines the interface for organization repository operations
type OrganizationRepositoryInterface interface {
	Create(org *models.Organization) error
	GetByID(id int) (*models.Organization, error)
	GetBySlug(slug string) (*models.Organization, error)
	GetAll(activeOnly bool) ([]*models.Organization, error)
	Update(id int, updates map[string]any) error
	Delete(id int) error
}

// UserInvitationRepositoryInterface defines the interface for user invitation repository operations
type UserInvitationRepositoryInterface interface {
	Create(invitation *models.UserInvitation) error
	GetByID(id int) (*models.UserInvitation, error)
	GetByToken(token string) (*models.UserInvitation, error)
	GetByEmailAndOrg(email string, orgID int) (*models.UserInvitation, error)
	GetPendingByOrganization(orgID int) ([]*models.UserInvitation, error)
	GetByOrganization(orgID int) ([]*models.UserInvitation, error)
	AcceptInvitation(invitationID int, userID int) error
	Delete(id int) error
	CleanupExpiredInvitations() error
}

// UserOrganizationRoleRepositoryInterface defines operations for multi-org membership
type UserOrganizationRoleRepositoryInterface interface {
	GetByUserID(userID int) ([]models.UserOrganizationRole, error)
	GetByUserAndOrg(userID, orgID int) (*models.UserOrganizationRole, error)
	Upsert(userID, orgID int, role string) error
	Delete(userID, orgID int) error
}

// OrganizationServiceInterface defines the interface for organization service operations
type OrganizationServiceInterface interface {
	CreateOrganization(req *models.CreateOrganizationRequest, userID int) (*models.Organization, error)
	GetOrganization(id int) (*models.Organization, error)
	GetOrganizationBySlug(slug string) (*models.Organization, error)
	ListOrganizations(activeOnly bool) ([]*models.Organization, error)
	UpdateOrganization(id int, req *models.UpdateOrganizationRequest, userID int) error
	DeleteOrganization(id int, userID int) error
	InviteUserToOrganization(orgID int, req *models.InviteUserRequest, invitedByUserID int) (*models.UserInvitation, error)
	GetPendingInvitations(orgID int) ([]*models.UserInvitation, error)
	RevokeInvitation(invitationID int, revokedByUserID int) error
}

// BuildingRepositoryInterface defines the interface for building repository operations
type BuildingRepositoryInterface interface {
	Create(building *models.Building) error
	GetByID(id int) (*models.Building, error)
	GetByPropertyID(propertyID int) ([]*models.Building, error)
	GetByPropertyAndCode(propertyID int, code string) (*models.Building, error)
	Update(id int, updates map[string]any) error
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
	RegisterWithInvitation(req *models.RegisterWithInvitationRequest) (*models.LoginResponse, error)
	Logout(userID int, clientIP, userAgent string) error
	ChangePassword(userID int, currentPassword, newPassword string) error
	ResetPassword(email string) error
	ConfirmPasswordReset(token, newPassword string) error
	GetUserByID(id int) (*models.User, error)
	GetAllUsers(activeOnly bool) ([]*models.User, error)
	GetUsersByOrganization(orgID int, activeOnly bool) ([]*models.User, error)
	GetUserByIDInOrganization(userID int, orgID int) (*models.User, error)
	UpdateUser(id int, req *models.UpdateUserRequest) error
	DeleteUser(id int) error
	UpdateUserInOrganization(id int, req *models.UpdateUserRequest, orgID int) error
	DeleteUserInOrganization(id int, orgID int) error
	AdminResetPassword(id int, newPassword string) error
	AdminResetPasswordInOrganization(id int, newPassword string, orgID int) error
	SetOrganization(userID int, req *models.SetOrganizationRequest) (*models.SetOrganizationResponse, error)
}

// MetadataValidatorInterface defines the interface for building metadata validation
type MetadataValidatorInterface interface {
	ValidateMetadata(buildingType models.BuildingType, metadata models.BuildingMetadata) error
	ValidateResidentialMetadata(metadata models.BuildingMetadata) error
	ValidateCommercialMetadata(metadata models.BuildingMetadata) error
	ValidateMixedMetadata(metadata models.BuildingMetadata) error
	GetMetadataSchema(buildingType models.BuildingType) map[string]any
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
	GetBuilding(id, orgID int) (*models.Building, error)
	GetBuildingWithStats(id, orgID int) (*models.BuildingWithStats, error)
	UpdateBuilding(id int, req *models.UpdateBuildingRequest, orgID int) (*models.Building, error)
	DeleteBuilding(id, orgID int) error

	// Property-building relationship operations
	GetBuildingsByProperty(propertyID int) ([]*models.Building, error)
	GetBuildingsByPropertyWithStats(propertyID int) ([]*models.BuildingWithStats, error)
	GetBuildingByPropertyAndCode(propertyID int, code string) (*models.Building, error)

	// Bulk operations
	BulkCreateBuildings(req *models.BulkCreateBuildingsRequest) ([]*models.Building, error)

	// Search and filtering
	SearchBuildings(filters *models.BuildingSearchFilters) ([]*models.Building, error)

	// Building-level aggregation and analytics
	GetBuildingAnalytics(id, orgID int) (*models.BuildingAnalytics, error)
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
	GetBuildingUnits(buildingID, orgID int, page, pageSize int) (*models.BuildingUnitsResponse, error)
	GetBuildingMetadataSchema(buildingType models.BuildingType) (*models.MetadataSchemaResponse, error)
	ExportBuildingData(req *models.BuildingExportRequest) ([]byte, string, error)
	UpdateBuildingStatus(buildingID, orgID int, req *models.BuildingStatusRequest) (*models.Building, error)
}

// UnitRepositoryInterface defines the interface for unit repository operations
type UnitRepositoryInterface interface {
	Create(req *models.CreateUnitRequest, organizationID int) (*models.Unit, error)
	GetByID(id int) (*models.Unit, error)
	GetByIDWithDetails(id int) (*models.UnitWithDetails, error)
	Update(id int, req *models.UpdateUnitRequest) (*models.Unit, error)
	Delete(id int) error
	CheckUnitNumberExists(buildingID int, unitNumber string, excludeID int) (bool, error)
	HasActiveLeases(unitID int) (bool, error)
	GetByBuildingWithDetails(buildingID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error)
	GetByPropertyWithDetails(propertyID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error)
	GetBuildingOccupancyStats(buildingID int, startDate, endDate time.Time) (any, error)
	GetPropertyOccupancyStats(propertyID int, startDate, endDate time.Time) (any, error)
	GetSystemOccupancyStats(startDate, endDate time.Time) (any, error)
	GetBuildingUnitTypeDistribution(buildingID int) (any, error)
}

// UnitServiceInterface defines the interface for unit service operations
type UnitServiceInterface interface {
	CreateUnit(req *models.CreateUnitRequest, userID, orgID int) (*models.Unit, error)
	GetUnit(id, orgID int) (*models.UnitWithDetails, error)
	UpdateUnit(id int, req *models.UpdateUnitRequest, userID, orgID int) (*models.Unit, error)
	DeleteUnit(id, userID, orgID int) error
	ValidateBuildingUnitRelationship(buildingID, propertyID int) error
	ValidateUnitTypeForBuilding(buildingID int, unitType models.UnitType) error
	GetUnitsByBuilding(buildingID int, page, pageSize, orgID int) ([]*models.UnitWithDetails, int, error)
	GetUnitsByProperty(propertyID int, page, pageSize, orgID int) ([]*models.UnitWithDetails, int, error)
	ValidateHierarchyIntegrity(unitID int) error
	GetUnitHierarchyContext(unitID, orgID int) (map[string]any, error)
}

// PaymentRepositoryInterface defines the interface for payment repository operations
type PaymentRepositoryInterface interface {
	Create(req *models.CreatePaymentRequest) (*models.Payment, error)
	GetByID(id int) (*models.Payment, error)
	GetByIDWithDetails(id int) (*models.PaymentWithDetails, error)
	Update(id int, req *models.UpdatePaymentRequest) (*models.Payment, error)
	GetWithDetailsAndFilters(filters map[string]any, limit, offset int) ([]*models.PaymentWithDetails, int, error)
	GetBuildingPaymentStats(buildingID int, startDate, endDate time.Time) (any, error)
	GetPropertyPaymentStats(propertyID int, startDate, endDate time.Time) (any, error)
	GetSystemPaymentStats(startDate, endDate time.Time) (any, error)
	GetBuildingPaymentsInPeriod(buildingID int, startDate, endDate time.Time, limit, offset int) ([]*models.PaymentWithDetails, int, error)
	GetDashboardSummary(orgID int) (*models.DashboardSummary, error)
	GetAgingBuckets(orgID int) (map[string]float64, error)
	GetMonthlyCollectionTrend(orgID int, months int) ([]*models.MonthlyCollectionTrend, error)
	GetTenantPaymentSummary(orgID int) ([]*models.TenantReportEntry, error)
	GetPaymentAnalyticsByPeriod(orgID int, startDate, endDate time.Time) (*models.PaymentAnalyticsResult, error)
	GetBuildingLevelSummary() (map[string]any, error)
	GetBuildingPaymentAnalytics(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentAnalytics, error)
	SearchLeases(orgID int, query string) ([]*models.LeaseSearchResult, error)
	GetActiveLeasesForPeriod(orgID, month, year int, buildingID *int) ([]*models.LeaseSearchResult, error)
	CheckPaymentExists(unitID, month, year int) (bool, error)
	GetBatchPropertyPaymentStats(propertyIDs []int, startDate, endDate time.Time) (map[int]any, error)
	NextReceiptNumber(orgID int, yearMonth string) (string, error)
}

// PaymentServiceInterface defines the interface for payment service operations
type PaymentServiceInterface interface {
	CreatePayment(req *models.CreatePaymentRequest, userID int) (*models.Payment, error)
	GetPayment(id, orgID int) (*models.PaymentWithDetails, error)
	UpdatePayment(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error)
	GetPayments(page, pageSize int, filters map[string]any) ([]*models.PaymentWithDetails, int, error)
	GetPaymentsByBuilding(buildingID int, page, pageSize int, filters map[string]any) ([]*models.PaymentWithDetails, int, error)
	GetPaymentsByProperty(propertyID int, page, pageSize int, filters map[string]any) ([]*models.PaymentWithDetails, int, error)
	GenerateBuildingPaymentReport(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error)
	GeneratePropertyPaymentReport(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error)
	GetDashboardSummaryWithBuildingContext(orgID int) (*models.DashboardSummary, error)
	ProcessBulkPayments(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error)
	GetPaymentAnalyticsByBuilding(buildingID int, period string) (*models.BuildingPaymentAnalytics, error)
	CanUserAccessPayment(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool
	LogPaymentAccess(userID int, action string, paymentID int, allowed bool)
	SearchLeases(orgID int, query string) (*models.LeaseSearchResponse, error)
	GenerateMonthlyPayments(req *models.GenerateMonthlyPaymentsRequest, orgID, userID int) (*models.GenerateMonthlyPaymentsResult, error)
	GenerateReceiptPDF(paymentID, orgID int) ([]byte, error)
}

// TenantRepositoryInterface defines the interface for tenant repository operations
type TenantRepositoryInterface interface {
	Create(req *models.CreateTenantRequest) (*models.Tenant, error)
	CheckEmailExists(email string, excludeID int) (bool, error)
	CheckNIDExists(nid string, excludeID int) (bool, error)
	GetByID(id int) (*models.Tenant, error)
	GetDecryptedNID(id int) (string, int, error)
	GetByUnitID(unitID int) (*models.Tenant, error)
	GetAll(page, pageSize, orgID int) ([]*models.Tenant, int, error)
	Update(id int, updates map[string]any) error
}

// TenantServiceInterface defines the interface for tenant service operations
type TenantServiceInterface interface {
	CreateTenant(req *models.CreateTenantRequest, userID int) (*models.TenantResponse, error)
	GetAllTenants(page, pageSize, orgID int) (*models.TenantListResponse, error)
	GetTenantByID(id int, orgID int) (*models.TenantWithLeases, error)
	UpdateTenant(id int, req *models.UpdateTenantRequest, userID, orgID int) (*models.TenantResponse, error)
	DeleteTenant(id int, userID, orgID int) error
	RevealNID(id, orgID, userID int) (string, error)
}

// PropertyRepositoryInterface defines the interface for property repository operations
type PropertyRepositoryInterface interface {
	GetByID(id int) (*models.Property, error)
	GetByIDWithStats(id int) (*models.PropertyWithStats, error)
	List(filters map[string]any, limit, offset int) ([]*models.Property, int, error)
	GetBuildingAggregations(propertyID int) (map[string]any, error)
	GetBuildingBreakdowns(propertyID int) (any, error)
	GetBuildingTypeDistribution(propertyID int) (any, error)
	GetBuildingCount(propertyID int) (int, error)
	GetBuildingSummary(propertyID int) (map[string]any, error)
	HasActiveBuildings(propertyID int) (bool, error)
	HasActiveUnits(propertyID int) (bool, error)
}

// LeaseRepositoryInterface defines the interface for lease repository operations
type LeaseRepositoryInterface interface {
	Create(req *models.CreateLeaseRequest) (*models.Lease, error)
	GetByID(id int) (*models.Lease, error)
	GetByIDWithDetails(id int) (*models.LeaseWithDetails, error)
	GetAll(page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error)
	GetByUnitID(unitID int, page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error)
	GetByTenantID(tenantID int, page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error)
	Update(id int, req *models.UpdateLeaseRequest) (*models.Lease, error)
	Delete(id int) error
	SoftDelete(id int) error
	HasActiveLeaseOnUnit(unitID int, excludeLeaseID *int) (bool, error)
	HasActiveLeaseForTenant(tenantID int) (bool, error)
	GetLeasesDueForMonth(orgID int) ([]models.LeaseDue, error)
	GetDueSummary(orgID int) (*models.DueSummary, error)
}

// LeaseServiceInterface defines the interface for lease service operations
type LeaseServiceInterface interface {
	CreateLease(req *models.CreateLeaseRequest, userID int) (*models.LeaseWithDetails, error)
	GetLeaseByID(id int, orgID int) (*models.LeaseWithDetails, error)
	GetAllLeases(page, pageSize, orgID int) (*models.LeaseListResponse, error)
	UpdateLease(id int, req *models.UpdateLeaseRequest, userID, orgID int) (*models.LeaseWithDetails, error)
	DeleteLease(id int, userID, orgID int) error
	TerminateLease(id int, userID, orgID int, terminationDate string) error
	GetLeasesByUnit(unitID int, page, pageSize, orgID int) (*models.LeaseListResponse, error)
	GetLeasesByTenant(tenantID int, page, pageSize, orgID int) (*models.LeaseListResponse, error)
	GetLeasesDue(orgID int) ([]models.LeaseDue, error)
	GetDueSummary(orgID int) (*models.DueSummary, error)
}

// ReportServiceInterface defines the interface for report service operations
type ReportServiceInterface interface {
	FinancialLedgerReport(orgID int, filters map[string]any, limit, offset int) (*models.FinancialLedgerReport, error)
	CollectionSummaryReport(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error)
	PaymentAnalysisReport(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error)
	TenantSummaryReport(orgID int) (*models.TenantSummaryReport, error)
	PropertyAnalyticsReport(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error)
}
