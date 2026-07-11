package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type UserHandler struct {
	userService interfaces.UserServiceInterface
}

func NewUserHandler(userService interfaces.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.userService.Login(&req, clientIP, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"code":  "LOGIN_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	response, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"code":  "TOKEN_REFRESH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "NOT_AUTHENTICATED",
		})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.userService.Logout(userID.(int), clientIP, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
			"code":  "LOGOUT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "NOT_AUTHENTICATED",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	if err := h.userService.ChangePassword(userID.(int), req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "PASSWORD_CHANGE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.ResetPassword(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password reset",
			"code":  "PASSWORD_RESET_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

func (h *UserHandler) SetOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "NOT_AUTHENTICATED",
		})
		return
	}

	var req models.SetOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	response, err := h.userService.SetOrganization(userID.(int), &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
			"code":  "ORGANIZATION_SWITCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) ConfirmPasswordReset(c *gin.Context) {
	var req models.ConfirmPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.ConfirmPasswordReset(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "PASSWORD_RESET_CONFIRMATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerRoleRaw, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	callerRole := callerRoleRaw.(string)

	if !canCreateRole(callerRole, req.Role) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("role %s cannot create users with role %s", callerRole, req.Role),
			"code":  "INSUFFICIENT_PERMISSIONS",
		})
		return
	}

	// Non-SUPER_ADMIN callers must create users within their own organization
	if callerRole != "SUPER_ADMIN" {
		callerOrgID, _ := c.Get("organization_id")
		orgIDPtr, ok := callerOrgID.(*int)
		if !ok || orgIDPtr == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no organization context"})
			return
		}
		req.OrganizationID = orgIDPtr
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// canCreateRole enforces role hierarchy: callerRole determines what targetRole can be created
func canCreateRole(callerRole, targetRole string) bool {
	switch callerRole {
	case "SUPER_ADMIN":
		return true
	case "ORG_ADMIN":
		return targetRole == "Admin" || targetRole == "PropertyManager" || targetRole == "Accountant"
	case "Admin":
		return targetRole == "PropertyManager" || targetRole == "Accountant"
	default:
		return false
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	callerRole, _ := c.Get("role")
	callerOrgID, _ := c.Get("organization_id")

	// Default: active users only. Pass ?active=false to include inactive.
	activeOnly := c.DefaultQuery("active", "true") != "false"

	var users []*models.User
	var err error

	switch callerRole {
	case "SUPER_ADMIN":
		users, err = h.userService.GetAllUsers(activeOnly)
	default:
		orgIDPtr, ok := callerOrgID.(*int)
		if !ok || orgIDPtr == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "no organization context"})
			return
		}
		users, err = h.userService.GetUsersByOrganization(*orgIDPtr, activeOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// callerOrgID extracts the organization_id set by AuthRequired for the current
// caller. Returns ok=false if the caller has no organization context (e.g. SUPER_ADMIN).
func callerOrgID(c *gin.Context) (int, bool) {
	raw, exists := c.Get("organization_id")
	if !exists {
		return 0, false
	}
	orgIDPtr, ok := raw.(*int)
	if !ok || orgIDPtr == nil {
		return 0, false
	}
	return *orgIDPtr, true
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	callerRole, _ := c.Get("role")

	var user *models.User
	if callerRole == "SUPER_ADMIN" {
		user, err = h.userService.GetUserByID(id)
	} else {
		orgID, ok := callerOrgID(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "no organization context"})
			return
		}
		user, err = h.userService.GetUserByIDInOrganization(id, orgID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerRole, _ := c.Get("role")
	if callerRole == "SUPER_ADMIN" {
		err = h.userService.UpdateUser(id, &req)
	} else {
		orgID, ok := callerOrgID(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "no organization context"})
			return
		}
		err = h.userService.UpdateUserInOrganization(id, &req, orgID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	callerRole, _ := c.Get("role")
	if callerRole == "SUPER_ADMIN" {
		err = h.userService.DeleteUser(id)
	} else {
		orgID, ok := callerOrgID(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "no organization context"})
			return
		}
		err = h.userService.DeleteUserInOrganization(id, orgID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) RegisterWithInvitation(c *gin.Context) {
	var req models.RegisterWithInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.userService.RegisterWithInvitation(&req)
	if err != nil {
		// Determine appropriate status code based on error message
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid invitation: invalid or expired invitation token" ||
			err.Error() == "invalid invitation: invitation has already been accepted" ||
			err.Error() == "invalid invitation: invitation has expired" ||
			err.Error() == "email does not match invitation email" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "password validation failed" || err.Error() == "username validation failed" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
			"code":  "REGISTRATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}
