package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/services"
)

// MFAHandler exposes TOTP enrollment and verification endpoints.
type MFAHandler struct {
	mfaService *services.MFAService
}

func NewMFAHandler(mfaService *services.MFAService) *MFAHandler {
	return &MFAHandler{mfaService: mfaService}
}

// GetStatus reports whether the authenticated user has confirmed MFA.
func (h *MFAHandler) GetStatus(c *gin.Context) {
	userID := c.GetInt("user_id")
	enabled, err := h.mfaService.Status(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "MFA_STATUS_FAILED", "Failed to get MFA status", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

// Enroll generates a new TOTP secret and returns the provisioning URI + secret
// for the client to display as a QR code / manual key.
func (h *MFAHandler) Enroll(c *gin.Context) {
	userID := c.GetInt("user_id")
	account := c.GetString("username")
	if account == "" {
		account = "user"
	}

	secret, uri, err := h.mfaService.Enroll(userID, account)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "MFA_ENROLL_FAILED", "Failed to start MFA enrollment", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"secret": secret, "otpauth_uri": uri})
}

// Verify validates a submitted TOTP code and, on success, returns a short-lived
// step-up token to authorize sensitive actions.
func (h *MFAHandler) Verify(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "MFA_VERIFY_INVALID_BODY", "A verification code is required", err)
		return
	}

	token, expiresAt, err := h.mfaService.Verify(userID, req.Code)
	if err != nil {
		// Enrollment-missing and bad-code are both client errors; keep the public
		// message generic while logging the detail server-side.
		respondError(c, http.StatusUnauthorized, "MFA_VERIFY_FAILED", "Invalid verification code", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"step_up_token": token,
		"expires_at":    expiresAt,
	})
}
