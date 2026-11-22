package services

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
)

var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)
var organizationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var ErrInvalidOperationsInput = errors.New("invalid operations input")

type OperationsService struct {
	repo                  *repositories.OperationsRepository
	auditService          interfaces.AuditServiceInterface
	invitationEmailSender InvitationEmailSender
}

func NewOperationsService(
	repo *repositories.OperationsRepository,
	auditService interfaces.AuditServiceInterface,
	invitationEmailSender ...InvitationEmailSender,
) *OperationsService {
	service := &OperationsService{repo: repo, auditService: auditService}
	if len(invitationEmailSender) > 0 {
		service.invitationEmailSender = invitationEmailSender[0]
	}
	return service
}

func (s *OperationsService) OnboardOrganization(
	req *models.CreateOrganizationOnboardingRequest,
	userID int,
) (*models.OrganizationOnboardingResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	req.PlanCode = strings.ToLower(strings.TrimSpace(req.PlanCode))
	req.AdminEmail = strings.ToLower(strings.TrimSpace(req.AdminEmail))
	if req.Name == "" || !organizationSlugPattern.MatchString(req.Slug) {
		return nil, fmt.Errorf("%w: organization name and a valid slug are required", ErrInvalidOperationsInput)
	}

	token, err := generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("generate initial admin invitation: %w", err)
	}
	expiresAt := time.Now().AddDate(0, 0, 7)
	onboarding, err := s.repo.CreateOrganizationOnboarding(req, userID, token, expiresAt)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(userID, "ONBOARD_ORGANIZATION", "organizations", &onboarding.Organization.ID, nil, onboarding)

	deliveryStatus := "failed"
	if s.invitationEmailSender == nil {
		log.Printf("initial admin invitation email is not configured for organization %d", onboarding.Organization.ID)
	} else if err := s.invitationEmailSender.SendInvitation(req.AdminEmail, req.Name, token); err != nil {
		log.Printf("failed to deliver initial admin invitation for organization %d: %v", onboarding.Organization.ID, err)
	} else {
		deliveryStatus = "sent"
	}
	if err := s.repo.UpdateOrganizationInvitationDeliveryStatus(onboarding.Invitation.ID, deliveryStatus); err != nil {
		return nil, fmt.Errorf("update initial admin invitation delivery status: %w", err)
	}
	onboarding.Invitation.EmailDeliveryStatus = deliveryStatus
	if deliveryStatus == "sent" {
		now := time.Now().UTC()
		onboarding.Invitation.LastSentAt = &now
	}
	return onboarding, nil
}

func (s *OperationsService) ListPendingOrganizationInvitations() ([]models.OrganizationInvitationSummary, error) {
	return s.repo.ListPendingOrganizationInvitations()
}

func (s *OperationsService) ResendOrganizationInvitation(invitationID, userID int) (*models.OrganizationInvitationSummary, error) {
	token, err := generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("generate replacement invitation token: %w", err)
	}
	invitation, err := s.repo.PrepareOrganizationInvitationResend(invitationID, token, time.Now().AddDate(0, 0, 7))
	if err != nil {
		return nil, err
	}

	deliveryStatus := "failed"
	if s.invitationEmailSender == nil {
		log.Printf("invitation email is not configured for invitation %d", invitationID)
	} else if err := s.invitationEmailSender.SendInvitation(invitation.Email, invitation.OrganizationName, token); err != nil {
		log.Printf("failed to resend invitation %d: %v", invitationID, err)
	} else {
		deliveryStatus = "sent"
	}
	if err := s.repo.UpdateOrganizationInvitationDeliveryStatus(invitationID, deliveryStatus); err != nil {
		return nil, fmt.Errorf("update resent invitation delivery status: %w", err)
	}
	invitation.EmailDeliveryStatus = deliveryStatus
	if deliveryStatus == "sent" {
		now := time.Now().UTC()
		invitation.LastSentAt = &now
	}
	_ = s.auditService.LogUserAction(userID, "RESEND_ORGANIZATION_INVITATION", "user_invitations", &invitationID, nil, map[string]string{"email_delivery_status": deliveryStatus})
	return invitation, nil
}

func (s *OperationsService) RevokeOrganizationInvitation(invitationID, userID int) error {
	if err := s.repo.RevokeOrganizationInvitation(invitationID); err != nil {
		return err
	}
	_ = s.auditService.LogUserAction(userID, "REVOKE_ORGANIZATION_INVITATION", "user_invitations", &invitationID, nil, nil)
	return nil
}

func (s *OperationsService) ListPlans(publicOnly bool) ([]models.Plan, error) {
	return s.repo.ListPlans(publicOnly)
}

func (s *OperationsService) UpdatePlan(code string, req *models.UpdatePlanRequest, userID int) (*models.Plan, error) {
	if req.PriceMinor != nil && *req.PriceMinor < 0 {
		return nil, fmt.Errorf("%w: price_minor must not be negative", ErrInvalidOperationsInput)
	}
	if req.MaxProperties != nil && *req.MaxProperties < 1 {
		return nil, fmt.Errorf("%w: max_properties must be positive or omitted", ErrInvalidOperationsInput)
	}
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if !currencyCodePattern.MatchString(req.Currency) {
		return nil, fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidOperationsInput)
	}
	req.NameEn = strings.TrimSpace(req.NameEn)
	req.NameBn = strings.TrimSpace(req.NameBn)
	req.DescriptionEn = strings.TrimSpace(req.DescriptionEn)
	req.DescriptionBn = strings.TrimSpace(req.DescriptionBn)
	if req.NameEn == "" || req.NameBn == "" {
		return nil, fmt.Errorf("%w: plan names are required in both languages", ErrInvalidOperationsInput)
	}
	featuresEn, err := sanitizePlanFeatures(req.FeaturesEn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOperationsInput, err)
	}
	featuresBn, err := sanitizePlanFeatures(req.FeaturesBn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOperationsInput, err)
	}
	req.FeaturesEn = featuresEn
	req.FeaturesBn = featuresBn

	plan, err := s.repo.UpdatePlan(code, req, userID)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(userID, "UPDATE_PLAN", "plans", &plan.ID, nil, plan)
	return plan, nil
}

func sanitizePlanFeatures(features []string) ([]string, error) {
	if len(features) > 20 {
		return nil, fmt.Errorf("a plan may contain at most 20 features")
	}
	cleaned := make([]string, 0, len(features))
	for _, feature := range features {
		feature = strings.TrimSpace(feature)
		if feature == "" {
			continue
		}
		if utf8.RuneCountInString(feature) > 160 {
			return nil, fmt.Errorf("each plan feature must be 160 characters or fewer")
		}
		cleaned = append(cleaned, feature)
	}
	return cleaned, nil
}

func (s *OperationsService) ListOrganizations() ([]models.OrganizationCustomer, error) {
	return s.repo.ListOrganizations()
}

func (s *OperationsService) GetSubscription(organizationID int) (*models.OrganizationSubscription, error) {
	return s.repo.GetSubscription(organizationID)
}

func (s *OperationsService) UpdateSubscription(organizationID int, req *models.UpdateSubscriptionRequest, userID int) (*models.OrganizationSubscription, error) {
	if err := s.ValidateSubscriptionDates(req); err != nil {
		return nil, err
	}
	subscription, err := s.repo.UpdateSubscription(organizationID, req, userID)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(userID, "UPDATE_SUBSCRIPTION", "organization_subscriptions", &subscription.ID, nil, subscription)
	return subscription, nil
}

func (s *OperationsService) ListSupportCases() ([]models.SupportCase, error) {
	return s.repo.ListSupportCases()
}

func (s *OperationsService) ListTenantSupportCases(organizationID int) ([]models.TenantSupportCase, error) {
	return s.repo.ListTenantSupportCases(organizationID)
}

func (s *OperationsService) GetTenantSupportCase(caseID, organizationID int) (*models.TenantSupportCase, error) {
	return s.repo.GetTenantSupportCase(caseID, organizationID)
}

func (s *OperationsService) CreateTenantSupportCase(
	organizationID, userID int,
	req *models.CreateTenantSupportCaseRequest,
) (*models.TenantSupportCase, error) {
	req.Subject = strings.TrimSpace(req.Subject)
	req.Description = strings.TrimSpace(req.Description)
	if req.Subject == "" || req.Description == "" {
		return nil, fmt.Errorf("%w: support subject and description are required", ErrInvalidOperationsInput)
	}
	supportCase, err := s.repo.CreateTenantSupportCase(organizationID, userID, req)
	if err != nil {
		return nil, err
	}
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "CREATE_SUPPORT_CASE", "support_cases", &supportCase.ID, nil, supportCase)
	}
	return supportCase, nil
}

func (s *OperationsService) ListSupportCaseMessages(caseID int) ([]models.SupportCaseMessage, error) {
	return s.repo.ListSupportCaseMessages(caseID)
}

func (s *OperationsService) AddSupportCaseMessage(
	caseID, authorID int,
	req *models.CreateSupportCaseMessageRequest,
) (*models.SupportCaseMessage, error) {
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		return nil, fmt.Errorf("%w: message body is required", ErrInvalidOperationsInput)
	}
	message, err := s.repo.AddSupportCaseMessage(caseID, authorID, req.Body)
	if err != nil {
		return nil, err
	}
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(authorID, "ADD_SUPPORT_CASE_MESSAGE", "support_case_messages", &message.ID, nil, map[string]int{"case_id": caseID})
	}
	return message, nil
}

func (s *OperationsService) ListSupportAgents() ([]models.SupportAgent, error) {
	return s.repo.ListSupportAgents()
}

func (s *OperationsService) UpdateSupportCaseAssignment(caseID, userID int, assignedTo *int) error {
	if assignedTo != nil && *assignedTo < 1 {
		return fmt.Errorf("%w: assigned_to must be a positive user ID or null", ErrInvalidOperationsInput)
	}
	if err := s.repo.UpdateSupportCaseAssignment(caseID, assignedTo); err != nil {
		return err
	}
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "ASSIGN_SUPPORT_CASE", "support_cases", &caseID, nil, map[string]*int{"assigned_to": assignedTo})
	}
	return nil
}

func (s *OperationsService) CreateSupportCase(req *models.CreateSupportCaseRequest, userID int) (*models.SupportCase, error) {
	req.Subject = strings.TrimSpace(req.Subject)
	req.Description = strings.TrimSpace(req.Description)
	if req.Subject == "" || req.Description == "" {
		return nil, fmt.Errorf("%w: support subject and description are required", ErrInvalidOperationsInput)
	}
	exists, err := s.repo.HasOrganization(req.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repositories.ErrOrganizationNotFound
	}
	supportCase, err := s.repo.CreateSupportCase(req, userID)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(userID, "CREATE_SUPPORT_CASE", "support_cases", &supportCase.ID, nil, supportCase)
	return supportCase, nil
}

func (s *OperationsService) UpdateSupportCaseStatus(caseID int, status string, userID int) error {
	if err := s.repo.UpdateSupportCaseStatus(caseID, status); err != nil {
		return err
	}
	_ = s.auditService.LogUserAction(userID, "UPDATE_SUPPORT_CASE", "support_cases", &caseID, nil, map[string]string{"status": status})
	return nil
}

func (s *OperationsService) GetSupportCaseNotes(caseID int) ([]models.SupportCaseNote, error) {
	return s.repo.GetSupportCaseNotes(caseID)
}

func (s *OperationsService) AddSupportCaseNote(caseID, authorID int, req *models.CreateSupportCaseNoteRequest) (*models.SupportCaseNote, error) {
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		return nil, fmt.Errorf("%w: note body is required", ErrInvalidOperationsInput)
	}
	note, err := s.repo.AddSupportCaseNote(caseID, authorID, req.Body)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(authorID, "ADD_SUPPORT_CASE_NOTE", "support_case_notes", &note.ID, nil, map[string]int{"case_id": caseID})
	return note, nil
}

func (s *OperationsService) AddTenantSupportCaseMessage(
	caseID, organizationID, authorID int,
	req *models.CreateSupportCaseMessageRequest,
) (*models.SupportCaseMessage, error) {
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		return nil, fmt.Errorf("%w: message body is required", ErrInvalidOperationsInput)
	}
	message, err := s.repo.AddTenantSupportCaseMessage(caseID, organizationID, authorID, req.Body)
	if err != nil {
		return nil, err
	}
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(authorID, "ADD_SUPPORT_CASE_MESSAGE", "support_case_messages", &message.ID, nil, map[string]int{"case_id": caseID})
	}
	return message, nil
}

func (s *OperationsService) ValidateSubscriptionDates(req *models.UpdateSubscriptionRequest) error {
	if req.StartedAt != nil && req.StartedAt.After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("%w: subscription start date cannot be in the future", ErrInvalidOperationsInput)
	}
	if req.RenewsAt != nil && req.StartedAt != nil && req.RenewsAt.Before(*req.StartedAt) {
		return fmt.Errorf("%w: renewal date cannot be before the subscription start date", ErrInvalidOperationsInput)
	}
	return nil
}
