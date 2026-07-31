package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/middleware"
	"github.com/ysnarafat/tenantly/internal/models"
)

const refreshCookieName = "refresh_token"
const refreshCookiePath = "/api/v1/auth"
const refreshCookieMaxAge = 7 * 24 * 60 * 60 // 7 days, in seconds — matches generateTokens' refresh token expiry

type UserHandler struct {
	userService interfaces.UserServiceInterface
	// accountLoginLimiter throttles login attempts per account (username),
	// independent of the per-IP limiter applied at the route level — mitigates
	// distributed credential stuffing against a single account from many IPs.
	accountLoginLimiter *middleware.RateLimiter
	cookieDomain        string
	cookieSecure        bool
}

func NewUserHandler(userService interfaces.UserServiceInterface, cookieDomain string, cookieSecure bool) *UserHandler {
	return &UserHandler{
		userService:         userService,
		accountLoginLimiter: middleware.NewRateLimiter(5, 15*time.Minute),
		cookieDomain:        cookieDomain,
		cookieSecure:        cookieSecure,
	}
}

// setRefreshCookie sets the refresh token as an httpOnly cookie rather than
// returning it in the JSON body — an XSS payload that can read localStorage
// cannot read an httpOnly cookie, so it can't steal the long-lived credential.
func (h *UserHandler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, token, refreshCookieMaxAge, refreshCookiePath, h.cookieDomain, h.cookieSecure, true)
}

func (h *UserHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, refreshCookiePath, h.cookieDomain, h.cookieSecure, true)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	// Per-account throttling, independent of the route's per-IP limiter — a
	// distributed credential-stuffing attempt against one account from many
	// IPs would otherwise never trip the per-IP limit.
	accountKey := strings.ToLower(strings.TrimSpace(req.Username))
	if !h.accountLoginLimiter.Allow(accountKey) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Too many login attempts for this account. Please try again later.",
			"code":  "ACCOUNT_RATE_LIMIT_EXCEEDED",
		})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.userService.Login(&req, clientIP, userAgent)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "LOGIN_FAILED", "Invalid credentials", err)
		return
	}

	h.setRefreshCookie(c, response.RefreshToken)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Refresh token missing",
			"code":  "TOKEN_REFRESH_FAILED",
		})
		return
	}

	response, err := h.userService.RefreshToken(refreshToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "TOKEN_REFRESH_FAILED", "Failed to refresh token", err)
		return
	}

	h.setRefreshCookie(c, response.RefreshToken)
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

	h.clearRefreshCookie(c)
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
		respondError(c, http.StatusBadRequest, "PASSWORD_CHANGE_FAILED", "Failed to change password", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
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
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	response, err := h.userService.SetOrganization(userID.(int), &req)
	if err != nil {
		respondError(c, http.StatusForbidden, "ORGANIZATION_SWITCH_FAILED", "Failed to switch organization", err)
		return
	}

	h.setRefreshCookie(c, response.RefreshToken)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) ConfirmPasswordReset(c *gin.Context) {
	var req models.ConfirmPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	if err := h.userService.ConfirmPasswordReset(req.Token, req.NewPassword); err != nil {
		respondError(c, http.StatusBadRequest, "PASSWORD_RESET_CONFIRMATION_FAILED", "Failed to reset password", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
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
		respondError(c, http.StatusBadRequest, "CREATE_USER_FAILED", "Failed to create user", err)
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
		respondError(c, http.StatusInternalServerError, "GET_USERS_FAILED", "Failed to retrieve users", err)
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
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
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
		respondError(c, http.StatusInternalServerError, "UPDATE_USER_FAILED", "Failed to update user", err)
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
		respondError(c, http.StatusInternalServerError, "DELETE_USER_FAILED", "Failed to delete user", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) RegisterWithInvitation(c *gin.Context) {
	var req models.RegisterWithInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
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

		respondError(c, statusCode, "REGISTRATION_FAILED", "Failed to complete registration", err)
		return
	}

	c.JSON(http.StatusCreated, response)
}
