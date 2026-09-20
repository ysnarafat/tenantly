package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupRequireOrganizationContext seeds the gin context the way AuthRequired
// (caller's own org, "organization_id") and OrganizationValidationMiddleware
// (requested org, "requested_organization_id") would, then runs RequireOrganization.
func setupRequireOrganizationContext(role string, callerOrgID *int, requestedOrgID int) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("role", role)
		c.Set("organization_id", callerOrgID)
		c.Set("requested_organization_id", requestedOrgID)
		c.Next()
	}, RequireOrganization(nil), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)
	return w
}

func TestRequireOrganization_DeniesCrossOrgAccess(t *testing.T) {
	callerOrg := 1
	requestedOrg := 2 // a different organization

	w := setupRequireOrganizationContext("Admin", &callerOrg, requestedOrg)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireOrganization_AllowsSameOrgAccess(t *testing.T) {
	callerOrg := 1
	requestedOrg := 1 // same organization

	w := setupRequireOrganizationContext("Admin", &callerOrg, requestedOrg)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireOrganization_SuperAdminBypassesOrgCheck(t *testing.T) {
	callerOrg := 1
	requestedOrg := 2 // different organization, but caller is SUPER_ADMIN

	w := setupRequireOrganizationContext("SUPER_ADMIN", &callerOrg, requestedOrg)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireOrganization_DeniesWhenCallerHasNoOrg(t *testing.T) {
	// SUPER_ADMIN accounts may have no organization_id in their JWT; a
	// non-SUPER_ADMIN caller with no org context must be denied, not allowed
	// through by a nil-pointer fallback.
	w := setupRequireOrganizationContext("Admin", nil, 2)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
