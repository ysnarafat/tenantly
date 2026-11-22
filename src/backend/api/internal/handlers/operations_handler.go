package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"
)

type OperationsHandler struct {
	service *services.OperationsService
}

func NewOperationsHandler(service *services.OperationsService) *OperationsHandler {
	return &OperationsHandler{service: service}
}

func (h *OperationsHandler) ListPublicPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(true)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_PLANS_FAILED", "Failed to load plans", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *OperationsHandler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(false)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_PLANS_FAILED", "Failed to load plans", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *OperationsHandler) UpdatePlan(c *gin.Context) {
	var req models.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid plan data", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	plan, err := h.service.UpdatePlan(c.Param("code"), &req, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrPlanNotFound) {
			respondError(c, http.StatusNotFound, "PLAN_NOT_FOUND", "Plan not found", err)
			return
		}
		if errors.Is(err, services.ErrInvalidOperationsInput) {
			respondError(c, http.StatusBadRequest, "UPDATE_PLAN_FAILED", "Unable to update plan", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "UPDATE_PLAN_FAILED", "Unable to update plan", err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *OperationsHandler) ListOrganizations(c *gin.Context) {
	orgs, err := h.service.ListOrganizations()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_ORGANIZATIONS_FAILED", "Failed to load organizations", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"organizations": orgs, "total": len(orgs)})
}

func (h *OperationsHandler) OnboardOrganization(c *gin.Context) {
	var req models.CreateOrganizationOnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid organization onboarding data", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	onboarding, err := h.service.OnboardOrganization(&req, userID)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrOrganizationSlugUsed):
			respondError(c, http.StatusConflict, "ORGANIZATION_SLUG_EXISTS", "Organization slug is already in use", err)
		case errors.Is(err, repositories.ErrInvitationEmailUsed):
			respondError(c, http.StatusConflict, "INVITATION_EMAIL_EXISTS", "An account already exists for this email", err)
		case errors.Is(err, repositories.ErrPlanNotFound):
			respondError(c, http.StatusNotFound, "PLAN_NOT_AVAILABLE", "Selected plan is unavailable", err)
		case errors.Is(err, services.ErrInvalidOperationsInput):
			respondError(c, http.StatusBadRequest, "INVALID_ONBOARDING_DATA", "Invalid organization onboarding data", err)
		default:
			respondError(c, http.StatusInternalServerError, "ORGANIZATION_ONBOARDING_FAILED", "Failed to onboard organization", err)
		}
		return
	}
	c.JSON(http.StatusCreated, onboarding)
}

func (h *OperationsHandler) ListPendingOrganizationInvitations(c *gin.Context) {
	invitations, err := h.service.ListPendingOrganizationInvitations()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_INVITATIONS_FAILED", "Failed to load pending invitations", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"invitations": invitations, "total": len(invitations)})
}

func (h *OperationsHandler) ResendOrganizationInvitation(c *gin.Context) {
	invitationID, ok := parseOperationID(c, "invitation_id")
	if !ok {
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	invitation, err := h.service.ResendOrganizationInvitation(invitationID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrInvitationNotFound):
			respondError(c, http.StatusNotFound, "INVITATION_NOT_FOUND", "Pending invitation not found", err)
		case errors.Is(err, repositories.ErrInvitationAccepted):
			respondError(c, http.StatusConflict, "INVITATION_ALREADY_ACCEPTED", "Invitation has already been accepted", err)
		default:
			respondError(c, http.StatusInternalServerError, "RESEND_INVITATION_FAILED", "Failed to resend invitation", err)
		}
		return
	}
	c.JSON(http.StatusOK, invitation)
}

func (h *OperationsHandler) RevokeOrganizationInvitation(c *gin.Context) {
	invitationID, ok := parseOperationID(c, "invitation_id")
	if !ok {
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.RevokeOrganizationInvitation(invitationID, userID); err != nil {
		if errors.Is(err, repositories.ErrInvitationNotFound) {
			respondError(c, http.StatusNotFound, "INVITATION_NOT_FOUND", "Pending invitation not found", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "REVOKE_INVITATION_FAILED", "Failed to revoke invitation", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OperationsHandler) GetSubscription(c *gin.Context) {
	organizationID, ok := parseOperationID(c, "organization_id")
	if !ok {
		return
	}
	subscription, err := h.service.GetSubscription(organizationID)
	if err != nil {
		if errors.Is(err, repositories.ErrSubscriptionNotFound) {
			respondError(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "GET_SUBSCRIPTION_FAILED", "Failed to load subscription", err)
		return
	}
	c.JSON(http.StatusOK, subscription)
}

func (h *OperationsHandler) UpdateSubscription(c *gin.Context) {
	organizationID, ok := parseOperationID(c, "organization_id")
	if !ok {
		return
	}
	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid subscription data", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	subscription, err := h.service.UpdateSubscription(organizationID, &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrOrganizationNotFound):
			respondError(c, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "Organization not found", err)
		case errors.Is(err, repositories.ErrPlanNotFound):
			respondError(c, http.StatusNotFound, "PLAN_NOT_FOUND", "Active plan not found", err)
		case errors.Is(err, services.ErrInvalidOperationsInput):
			respondError(c, http.StatusBadRequest, "UPDATE_SUBSCRIPTION_FAILED", "Unable to update subscription", err)
		default:
			respondError(c, http.StatusInternalServerError, "UPDATE_SUBSCRIPTION_FAILED", "Unable to update subscription", err)
		}
		return
	}
	c.JSON(http.StatusOK, subscription)
}

func (h *OperationsHandler) ListSupportCases(c *gin.Context) {
	cases, err := h.service.ListSupportCases()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_SUPPORT_CASES_FAILED", "Failed to load support cases", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases, "total": len(cases)})
}

func (h *OperationsHandler) ListTenantSupportCases(c *gin.Context) {
	organizationID, ok := authenticatedOrganizationID(c)
	if !ok {
		return
	}
	cases, err := h.service.ListTenantSupportCases(organizationID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_SUPPORT_CASES_FAILED", "Failed to load support cases", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases, "total": len(cases)})
}

func (h *OperationsHandler) GetTenantSupportCase(c *gin.Context) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	organizationID, ok := authenticatedOrganizationID(c)
	if !ok {
		return
	}
	supportCase, err := h.service.GetTenantSupportCase(caseID, organizationID)
	if err != nil {
		if errors.Is(err, repositories.ErrSupportCaseNotFound) {
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "GET_SUPPORT_CASE_FAILED", "Failed to load support case", err)
		return
	}
	c.JSON(http.StatusOK, supportCase)
}

func (h *OperationsHandler) CreateTenantSupportCase(c *gin.Context) {
	var req models.CreateTenantSupportCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid support case data", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	organizationID, ok := authenticatedOrganizationID(c)
	if !ok {
		return
	}
	supportCase, err := h.service.CreateTenantSupportCase(organizationID, userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOperationsInput) {
			respondError(c, http.StatusBadRequest, "CREATE_SUPPORT_CASE_FAILED", "Invalid support case", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "CREATE_SUPPORT_CASE_FAILED", "Failed to create support case", err)
		return
	}
	c.JSON(http.StatusCreated, supportCase)
}

func (h *OperationsHandler) CreateTenantSupportCaseMessage(c *gin.Context) {
	h.addSupportCaseMessage(c, true)
}

func (h *OperationsHandler) ListSupportCaseMessages(c *gin.Context) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	messages, err := h.service.ListSupportCaseMessages(caseID)
	if err != nil {
		if errors.Is(err, repositories.ErrSupportCaseNotFound) {
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "LIST_SUPPORT_MESSAGES_FAILED", "Failed to load case messages", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *OperationsHandler) CreateSupportCaseMessage(c *gin.Context) {
	h.addSupportCaseMessage(c, false)
}

func (h *OperationsHandler) addSupportCaseMessage(c *gin.Context, tenant bool) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	var req models.CreateSupportCaseMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid support message", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var message *models.SupportCaseMessage
	var err error
	if tenant {
		organizationID, ok := authenticatedOrganizationID(c)
		if !ok {
			return
		}
		message, err = h.service.AddTenantSupportCaseMessage(caseID, organizationID, userID, &req)
	} else {
		message, err = h.service.AddSupportCaseMessage(caseID, userID, &req)
	}
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrSupportCaseNotFound):
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
		case errors.Is(err, services.ErrInvalidOperationsInput):
			respondError(c, http.StatusBadRequest, "CREATE_SUPPORT_MESSAGE_FAILED", "Invalid support message", err)
		default:
			respondError(c, http.StatusInternalServerError, "CREATE_SUPPORT_MESSAGE_FAILED", "Failed to add support message", err)
		}
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *OperationsHandler) ListSupportAgents(c *gin.Context) {
	agents, err := h.service.ListSupportAgents()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_SUPPORT_AGENTS_FAILED", "Failed to load support agents", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

func (h *OperationsHandler) UpdateSupportCaseAssignment(c *gin.Context) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	var req models.UpdateSupportCaseAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid support case assignment", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	err := h.service.UpdateSupportCaseAssignment(caseID, userID, req.AssignedTo)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrSupportCaseNotFound):
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
		case errors.Is(err, repositories.ErrSupportAgentNotFound), errors.Is(err, services.ErrInvalidOperationsInput):
			respondError(c, http.StatusBadRequest, "UPDATE_SUPPORT_CASE_ASSIGNMENT_FAILED", "Invalid support agent", err)
		default:
			respondError(c, http.StatusInternalServerError, "UPDATE_SUPPORT_CASE_ASSIGNMENT_FAILED", "Failed to assign support case", err)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OperationsHandler) CreateSupportCase(c *gin.Context) {
	var req models.CreateSupportCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid support case data", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	supportCase, err := h.service.CreateSupportCase(&req, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrOrganizationNotFound) {
			respondError(c, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "Organization not found", err)
			return
		}
		if errors.Is(err, services.ErrInvalidOperationsInput) {
			respondError(c, http.StatusBadRequest, "CREATE_SUPPORT_CASE_FAILED", "Invalid support case", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "CREATE_SUPPORT_CASE_FAILED", "Failed to create support case", err)
		return
	}
	c.JSON(http.StatusCreated, supportCase)
}

func (h *OperationsHandler) UpdateSupportCase(c *gin.Context) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	var req models.UpdateSupportCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid support case status", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	err := h.service.UpdateSupportCaseStatus(caseID, req.Status, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrSupportCaseNotFound) {
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "UPDATE_SUPPORT_CASE_FAILED", "Failed to update support case", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OperationsHandler) ListSupportCaseNotes(c *gin.Context) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	notes, err := h.service.GetSupportCaseNotes(caseID)
	if err != nil {
		if errors.Is(err, repositories.ErrSupportCaseNotFound) {
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "LIST_SUPPORT_NOTES_FAILED", "Failed to load case notes", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func (h *OperationsHandler) CreateSupportCaseNote(c *gin.Context) {
	caseID, ok := parseOperationID(c, "case_id")
	if !ok {
		return
	}
	var req models.CreateSupportCaseNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid support note", err)
		return
	}
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	note, err := h.service.AddSupportCaseNote(caseID, userID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrSupportCaseNotFound) {
			respondError(c, http.StatusNotFound, "SUPPORT_CASE_NOT_FOUND", "Support case not found", err)
			return
		}
		if errors.Is(err, services.ErrInvalidOperationsInput) {
			respondError(c, http.StatusBadRequest, "CREATE_SUPPORT_NOTE_FAILED", "Invalid support note", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "CREATE_SUPPORT_NOTE_FAILED", "Failed to add support note", err)
		return
	}
	c.JSON(http.StatusCreated, note)
}

func authenticatedUserID(c *gin.Context) (int, bool) {
	value, exists := c.Get("user_id")
	userID, ok := value.(int)
	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return 0, false
	}
	return userID, true
}

func authenticatedOrganizationID(c *gin.Context) (int, bool) {
	organizationID, exists := c.Get("org_id")
	id, ok := organizationID.(int)
	if !exists || !ok || id < 1 {
		respondError(c, http.StatusForbidden, "ORG_CONTEXT_REQUIRED", "Organization context is required", nil)
		return 0, false
	}
	return id, true
}

func parseOperationID(c *gin.Context, parameter string) (int, bool) {
	id, err := strconv.Atoi(c.Param(parameter))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return id, true
}
