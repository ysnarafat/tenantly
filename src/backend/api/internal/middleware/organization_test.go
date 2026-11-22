package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/ysnarafat/tenantly/internal/models"
)

type supportRoleRepositoryMock struct {
	role   string
	err    error
	userID int
	orgID  int
	called bool
}

func (m *supportRoleRepositoryMock) GetByUserID(int) ([]models.UserOrganizationRole, error) {
	return nil, nil
}

func (m *supportRoleRepositoryMock) GetByUserAndOrg(userID, orgID int) (*models.UserOrganizationRole, error) {
	m.called = true
	m.userID = userID
	m.orgID = orgID
	if m.err != nil {
		return nil, m.err
	}
	return &models.UserOrganizationRole{Role: m.role}, nil
}

func (m *supportRoleRepositoryMock) Upsert(int, int, string) error { return nil }
func (m *supportRoleRepositoryMock) Delete(int, int) error         { return nil }

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

func TestRequireOrgAdminContextChecksRoleAndMembership(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		membershipRole string
		membershipErr  error
		wantStatus     int
		wantRepoCall   bool
	}{
		{name: "allows matching organization admin membership", role: "ORG_ADMIN", membershipRole: "ORG_ADMIN", wantStatus: http.StatusOK, wantRepoCall: true},
		{name: "rejects other organization membership role", role: "ORG_ADMIN", membershipRole: "Admin", wantStatus: http.StatusForbidden, wantRepoCall: true},
		{name: "rejects non org admin", role: "Admin", wantStatus: http.StatusForbidden},
		{name: "rejects missing membership", role: "ORG_ADMIN", membershipErr: errors.New("membership missing"), wantStatus: http.StatusForbidden, wantRepoCall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &supportRoleRepositoryMock{role: tt.membershipRole, err: tt.membershipErr}
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/support", func(c *gin.Context) {
				c.Set("user_id", 7)
				c.Set("role", tt.role)
				c.Set("org_id", 42)
				c.Next()
			}, RequireOrgAdminContext(repo), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/support", nil))

			assert.Equal(t, tt.wantStatus, recorder.Code)
			assert.Equal(t, tt.wantRepoCall, repo.called)
			if tt.wantRepoCall {
				assert.Equal(t, 7, repo.userID)
				assert.Equal(t, 42, repo.orgID)
			}
		})
	}
}
