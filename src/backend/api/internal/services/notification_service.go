package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type NotificationService struct {
	notificationRepo interfaces.NotificationRepositoryInterface
	unitRepo         interfaces.UnitRepositoryInterface
	buildingRepo     interfaces.BuildingRepositoryInterface
	propertyRepo     interfaces.PropertyRepositoryInterface
	tenantRepo       interfaces.TenantRepositoryInterface
	auditService     interfaces.AuditServiceInterface
}

func NewNotificationService(
	notificationRepo interfaces.NotificationRepositoryInterface,
	unitRepo interfaces.UnitRepositoryInterface,
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo interfaces.PropertyRepositoryInterface,
	tenantRepo interfaces.TenantRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		unitRepo:         unitRepo,
		buildingRepo:     buildingRepo,
		propertyRepo:     propertyRepo,
		tenantRepo:       tenantRepo,
		auditService:     auditService,
	}
}

// CreateNotification creates a notification with building context in tenant communications
func (s *NotificationService) CreateNotification(req *models.CreateNotificationRequest, userID int) (*models.NotificationQueue, error) {
	// Get unit details for building context
	unit, err := s.unitRepo.GetByID(req.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Get building information for context
	building, err := s.buildingRepo.GetByID(unit.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Get property information for context
	property, err := s.propertyRepo.GetByID(unit.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get tenant information
	tenant, err := s.tenantRepo.GetByID(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Enhance message with building context
	enhancedMessage := s.enhanceMessageWithBuildingContext(req.Message, building, property, unit, tenant)

	// Create enhanced notification request
	enhancedReq := &models.CreateNotificationRequest{
		TenantID:         req.TenantID,
		UnitID:           req.UnitID,
		Message:          enhancedMessage,
		NotificationType: req.NotificationType,
		Recipient:        req.Recipient,
		BuildingID:       unit.BuildingID,
		PropertyID:       unit.PropertyID,
	}

	// Create notification
	notification, err := s.notificationRepo.Create(enhancedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Log audit with building context
	s.auditService.LogUserAction(userID, "CREATE", "notifications", &notification.ID, nil, map[string]interface{}{
		"notification_id":   notification.ID,
		"tenant_id":         notification.TenantID,
		"unit_id":           req.UnitID,
		"building_id":       unit.BuildingID,
		"property_id":       unit.PropertyID,
		"building_name":     building.BuildingName,
		"building_code":     building.BuildingCode,
		"property_name":     property.PropertyName,
		"unit_number":       unit.UnitNumber,
		"notification_type": notification.NotificationType,
		"recipient":         notification.Recipient,
	})

	return notification, nil
}

// enhanceMessageWithBuildingContext adds building and property context to notification messages
func (s *NotificationService) enhanceMessageWithBuildingContext(
	message string,
	building *models.Building,
	property *models.Property,
	unit *models.Unit,
	tenant *models.Tenant,
) string {
	// Define context variables for message templating
	contextVars := map[string]string{
		"{tenant_name}":      tenant.Name,
		"{unit_number}":      unit.UnitNumber,
		"{unit_name}":        unit.UnitName,
		"{building_name}":    building.BuildingName,
		"{building_code}":    building.BuildingCode,
		"{property_name}":    property.PropertyName,
		"{property_code}":    property.PropertyCode,
		"{property_address}": property.Address,
		"{building_type}":    string(building.BuildingType),
		"{floor}":            fmt.Sprintf("%d", unit.Floor),
		"{section}":          unit.Section,
	}

	// Replace context variables in message
	enhancedMessage := message
	for placeholder, value := range contextVars {
		enhancedMessage = strings.ReplaceAll(enhancedMessage, placeholder, value)
	}

	// Add building context footer if not already present
	if !strings.Contains(enhancedMessage, building.BuildingName) {
		enhancedMessage += fmt.Sprintf("\n\nLocation: %s, %s\nBuilding: %s (%s)",
			property.PropertyName, property.Address, building.BuildingName, building.BuildingCode)
	}

	return enhancedMessage
}

// SendBuildingWideNotification sends notifications to all tenants in a building
func (s *NotificationService) SendBuildingWideNotification(
	buildingID int,
	message string,
	notificationType string,
	userID int,
) ([]*models.NotificationQueue, []error) {
	// Validate building exists
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, []error{fmt.Errorf("building not found: %w", err)}
	}

	// Get all active units in the building
	units, _, err := s.unitRepo.GetByBuildingWithDetails(buildingID, 1000, 0) // Get all units
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get building units: %w", err)}
	}

	notifications := make([]*models.NotificationQueue, 0)
	errors := make([]error, 0)

	// Create notifications for each tenant
	for _, unit := range units {
		if !unit.LeaseActive {
			continue // Skip units without active leases
		}

		// Get tenant for the unit
		tenant, err := s.tenantRepo.GetByUnitID(unit.ID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get tenant for unit %s: %w", unit.UnitNumber, err))
			continue
		}

		// Create notification request
		req := &models.CreateNotificationRequest{
			TenantID:         tenant.ID,
			UnitID:           unit.ID,
			Message:          message,
			NotificationType: notificationType,
			Recipient:        tenant.Email, // or tenant.Phone based on notification type
		}

		notification, err := s.CreateNotification(req, userID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create notification for unit %s: %w", unit.UnitNumber, err))
			continue
		}

		notifications = append(notifications, notification)
	}

	// Log building-wide notification audit
	s.auditService.LogUserAction(userID, "CREATE_BUILDING_WIDE", "notifications", nil, nil, map[string]interface{}{
		"building_id":          buildingID,
		"building_name":        building.BuildingName,
		"building_code":        building.BuildingCode,
		"notification_type":    notificationType,
		"total_notifications":  len(notifications),
		"failed_notifications": len(errors),
	})

	return notifications, errors
}

// SendPropertyWideNotification sends notifications to all tenants in a property with building context
func (s *NotificationService) SendPropertyWideNotification(
	propertyID int,
	message string,
	notificationType string,
	userID int,
) ([]*models.NotificationQueue, []error) {
	// Validate property exists
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, []error{fmt.Errorf("property not found: %w", err)}
	}

	// Get all buildings in the property
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get property buildings: %w", err)}
	}

	allNotifications := make([]*models.NotificationQueue, 0)
	allErrors := make([]error, 0)

	// Send notifications to each building
	for _, building := range buildings {
		if !building.ActiveStatus {
			continue // Skip inactive buildings
		}

		notifications, errors := s.SendBuildingWideNotification(building.ID, message, notificationType, userID)
		allNotifications = append(allNotifications, notifications...)
		allErrors = append(allErrors, errors...)
	}

	// Log property-wide notification audit
	s.auditService.LogUserAction(userID, "CREATE_PROPERTY_WIDE", "notifications", nil, nil, map[string]interface{}{
		"property_id":          propertyID,
		"property_name":        property.PropertyName,
		"property_code":        property.PropertyCode,
		"notification_type":    notificationType,
		"buildings_count":      len(buildings),
		"total_notifications":  len(allNotifications),
		"failed_notifications": len(allErrors),
	})

	return allNotifications, allErrors
}

// GetNotificationsByBuilding retrieves notifications for a specific building
func (s *NotificationService) GetNotificationsByBuilding(buildingID int, page, pageSize int) ([]*models.NotificationWithDetails, int, error) {
	// Validate building exists
	_, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, 0, fmt.Errorf("building not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	notifications, total, err := s.notificationRepo.GetByBuildingWithDetails(buildingID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications by building: %w", err)
	}

	return notifications, total, nil
}

// GetNotificationsByProperty retrieves notifications for a specific property with building context
func (s *NotificationService) GetNotificationsByProperty(propertyID int, page, pageSize int) ([]*models.NotificationWithDetails, int, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, 0, fmt.Errorf("property not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	notifications, total, err := s.notificationRepo.GetByPropertyWithDetails(propertyID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications by property: %w", err)
	}

	return notifications, total, nil
}

// UpdateNotification updates a notification with building context logging
func (s *NotificationService) UpdateNotification(id int, req *models.UpdateNotificationRequest, userID int) (*models.NotificationQueue, error) {
	// Get existing notification for audit
	existingNotification, err := s.notificationRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing notification: %w", err)
	}

	// Update notification
	updatedNotification, err := s.notificationRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}

	// Log audit with building context
	s.auditService.LogUserAction(userID, "UPDATE", "notifications", &id, map[string]interface{}{
		"notification_id": existingNotification.ID,
		"building_name":   existingNotification.BuildingName,
		"building_code":   existingNotification.BuildingCode,
		"property_name":   existingNotification.PropertyName,
		"unit_number":     existingNotification.UnitNumber,
		"old_status":      existingNotification.Status,
	}, map[string]interface{}{
		"notification_id": updatedNotification.ID,
		"new_status":      updatedNotification.Status,
		"retry_count":     updatedNotification.RetryCount,
	})

	return updatedNotification, nil
}

// GenerateNotificationReport generates notification report with building breakdowns
func (s *NotificationService) GenerateNotificationReport(
	propertyID *int,
	buildingID *int,
	startDate, endDate time.Time,
) (*models.NotificationReport, error) {
	var report *models.NotificationReport
	var err error

	if buildingID != nil {
		// Generate building-specific report
		report, err = s.generateBuildingNotificationReport(*buildingID, startDate, endDate)
	} else if propertyID != nil {
		// Generate property-wide report with building breakdowns
		report, err = s.generatePropertyNotificationReport(*propertyID, startDate, endDate)
	} else {
		// Generate system-wide report
		report, err = s.generateSystemNotificationReport(startDate, endDate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate notification report: %w", err)
	}

	return report, nil
}

// generateBuildingNotificationReport generates notification report for a specific building
func (s *NotificationService) generateBuildingNotificationReport(buildingID int, startDate, endDate time.Time) (*models.NotificationReport, error) {
	// Get building details
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Get property details
	property, err := s.propertyRepo.GetByID(building.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get notification statistics
	stats, err := s.notificationRepo.GetBuildingNotificationStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building notification stats: %w", err)
	}

	report := &models.NotificationReport{
		ReportType:   "building",
		BuildingID:   &buildingID,
		BuildingName: &building.BuildingName,
		PropertyID:   &building.PropertyID,
		PropertyName: &property.PropertyName,
		StartDate:    startDate,
		EndDate:      endDate,
		Statistics:   stats,
		GeneratedAt:  time.Now(),
	}

	return report, nil
}

// generatePropertyNotificationReport generates notification report for a property with building breakdowns
func (s *NotificationService) generatePropertyNotificationReport(propertyID int, startDate, endDate time.Time) (*models.NotificationReport, error) {
	// Get property details
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get overall property statistics
	propertyStats, err := s.notificationRepo.GetPropertyNotificationStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property notification stats: %w", err)
	}

	// Get building breakdowns
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property buildings: %w", err)
	}

	buildingBreakdowns := make([]*models.BuildingNotificationBreakdown, 0, len(buildings))
	for _, building := range buildings {
		buildingStats, err := s.notificationRepo.GetBuildingNotificationStats(building.ID, startDate, endDate)
		if err != nil {
			continue // Skip buildings with errors
		}

		breakdown := &models.BuildingNotificationBreakdown{
			BuildingID:   building.ID,
			BuildingName: building.BuildingName,
			BuildingCode: building.BuildingCode,
			Statistics:   buildingStats,
		}
		buildingBreakdowns = append(buildingBreakdowns, breakdown)
	}

	report := &models.NotificationReport{
		ReportType:         "property",
		PropertyID:         &propertyID,
		PropertyName:       &property.PropertyName,
		StartDate:          startDate,
		EndDate:            endDate,
		Statistics:         propertyStats,
		BuildingBreakdowns: buildingBreakdowns,
		GeneratedAt:        time.Now(),
	}

	return report, nil
}

// generateSystemNotificationReport generates system-wide notification report with property and building context
func (s *NotificationService) generateSystemNotificationReport(startDate, endDate time.Time) (*models.NotificationReport, error) {
	// Get system-wide statistics
	systemStats, err := s.notificationRepo.GetSystemNotificationStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get system notification stats: %w", err)
	}

	report := &models.NotificationReport{
		ReportType:  "system",
		StartDate:   startDate,
		EndDate:     endDate,
		Statistics:  systemStats,
		GeneratedAt: time.Now(),
	}

	return report, nil
}
