package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/models"
)

// ---------------------------------------------------------------------------
// Access Control Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_GetPayment_AccessDenied tests that users without permission get 403
func TestPaymentHandler_GetPayment_AccessDenied(t *testing.T) {
	t.Run("forbidden when user lacks access to payment", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentFn: func(id int) (*models.PaymentWithDetails, error) {
				return &models.PaymentWithDetails{
					Payment: models.Payment{
						ID:             id,
						OrganizationID: 99, // Different org
					},
				}, nil
			},
			canAccessPaymentFn: func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
				return false // Deny access
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}

		var resp gin.H
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["error"] == nil {
			t.Errorf("expected error message in response")
		}
	})

	t.Run("allows access when user has permission", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentFn: func(id int) (*models.PaymentWithDetails, error) {
				return &models.PaymentWithDetails{
					Payment: models.Payment{ID: id, OrganizationID: 1},
				}, nil
			},
			canAccessPaymentFn: func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
				return true // Allow access
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

// TestPaymentHandler_UpdatePayment_AccessDenied tests update access control
func TestPaymentHandler_UpdatePayment_AccessDenied(t *testing.T) {
	t.Run("forbidden when user cannot access payment for update", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentFn: func(id int) (*models.PaymentWithDetails, error) {
				return &models.PaymentWithDetails{
					Payment: models.Payment{ID: id, OrganizationID: 99},
				}, nil
			},
			canAccessPaymentFn: func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
				return false
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		body := `{"status": "Paid", "amount_paid": 5000}`
		req := httptest.NewRequest(http.MethodPut, "/payments/1", nil)
		req.Body = io.NopCloser(strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("accountants cannot update payments (read-only)", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentFn: func(id int) (*models.PaymentWithDetails, error) {
				return &models.PaymentWithDetails{
					Payment: models.Payment{ID: id, OrganizationID: 1},
				}, nil
			},
			canAccessPaymentFn: func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
				return true
			},
		}

		gin.SetMode(gin.TestMode)
		handler := NewPaymentHandler(svc)
		r := gin.New()

		// Set userRole to Accountant
		r.Use(func(c *gin.Context) {
			c.Set("userID", 1)
			c.Set("org_id", 1)
			c.Set("userRole", "Accountant")
			c.Next()
		})

		r.PUT("/payments/:id", handler.UpdatePayment)

		w := httptest.NewRecorder()
		body := `{"status": "Paid", "amount_paid": 5000}`
		req := httptest.NewRequest(http.MethodPut, "/payments/1", nil)
		req.Body = io.NopCloser(strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 for Accountant update, got %d", w.Code)
		}
	})
}

// ---------------------------------------------------------------------------
// Role-Based Access Control Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_RoleBasedAccess tests different roles accessing payments
func TestPaymentHandler_RoleBasedAccess(t *testing.T) {
	testCases := []struct {
		role             string
		shouldHaveAccess bool
		description      string
	}{
		{"SUPER_ADMIN", true, "SUPER_ADMIN can access all payments"},
		{"ORG_ADMIN", true, "ORG_ADMIN can access org payments"},
		{"Admin", true, "Admin can access org payments"},
		{"PropertyManager", true, "PropertyManager can access managed properties"},
		{"Accountant", true, "Accountant has read-only access"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			svc := &mockPaymentService{
				getPaymentFn: func(id int) (*models.PaymentWithDetails, error) {
					return &models.PaymentWithDetails{
						Payment: models.Payment{ID: id, OrganizationID: 1, PropertyID: 1},
					}, nil
				},
				canAccessPaymentFn: func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
					// Simulate role-based access logic
					return userRole != "Unknown"
				},
			}

			gin.SetMode(gin.TestMode)
			handler := NewPaymentHandler(svc)
			r := gin.New()

			r.Use(func(c *gin.Context) {
				c.Set("userID", 1)
				c.Set("org_id", 1)
				c.Set("userRole", tc.role)
				c.Next()
			})

			r.GET("/payments/:id", handler.GetPayment)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments/1", nil)
			r.ServeHTTP(w, req)

			if tc.shouldHaveAccess && w.Code != http.StatusOK {
				t.Errorf("%s should have access, got status %d", tc.role, w.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Audit Logging Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_AuditLogging tests that payment access is logged
func TestPaymentHandler_AuditLogging(t *testing.T) {
	t.Run("successful payment creation is logged", func(t *testing.T) {
		var capturedAction string
		var capturedAllowed bool

		svc := &mockPaymentService{
			createPaymentFn: func(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
				return &models.Payment{ID: 1}, nil
			},
			logAccessFn: func(userID int, action string, paymentID int, allowed bool) {
				capturedAction = action
				capturedAllowed = allowed
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		body := `{
			"unit_id": 1, "tenant_id": 1, "building_id": 1, "property_id": 1,
			"month": 5, "year": 2026, "amount_due": 5000
		}`
		req := httptest.NewRequest(http.MethodPost, "/payments", nil)
		req.Body = io.NopCloser(strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if capturedAction != "CREATE" {
			t.Errorf("expected action CREATE, got %q", capturedAction)
		}
		if !capturedAllowed {
			t.Errorf("expected allowed=true, got false")
		}
	})

	t.Run("denied payment access is logged", func(t *testing.T) {
		var capturedAllowed *bool

		svc := &mockPaymentService{
			getPaymentFn: func(id int) (*models.PaymentWithDetails, error) {
				return &models.PaymentWithDetails{
					Payment: models.Payment{ID: id, OrganizationID: 99},
				}, nil
			},
			canAccessPaymentFn: func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
				return false
			},
			logAccessFn: func(userID int, action string, paymentID int, allowed bool) {
				capturedAllowed = &allowed
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/1", nil)
		router.ServeHTTP(w, req)

		if capturedAllowed == nil || *capturedAllowed {
			t.Errorf("expected denied access to be logged, got allowed=true or nil")
		}
	})
}

// ---------------------------------------------------------------------------
// Query Parameter Filtering Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_QueryFiltering tests query parameter handling
func TestPaymentHandler_QueryFiltering(t *testing.T) {
	t.Run("status filter is extracted from query params", func(t *testing.T) {
		var capturedFilters map[string]interface{}

		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedFilters = filters
				return make([]*models.PaymentWithDetails, 0), 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?status=Paid", nil)
		router.ServeHTTP(w, req)

		if status, ok := capturedFilters["status"]; !ok || status != "Paid" {
			t.Errorf("expected status filter, got %v", capturedFilters["status"])
		}
	})

	t.Run("month and year filters are converted to integers", func(t *testing.T) {
		var capturedFilters map[string]interface{}

		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedFilters = filters
				return make([]*models.PaymentWithDetails, 0), 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?month=5&year=2026", nil)
		router.ServeHTTP(w, req)

		if month, ok := capturedFilters["month"]; !ok {
			t.Errorf("month filter missing")
		} else if m, ok := month.(int); !ok || m != 5 {
			t.Errorf("month should be int 5, got %v (%T)", month, month)
		}

		if year, ok := capturedFilters["year"]; !ok {
			t.Errorf("year filter missing")
		} else if y, ok := year.(int); !ok || y != 2026 {
			t.Errorf("year should be int 2026, got %v (%T)", year, year)
		}
	})

	t.Run("invalid numeric filters are ignored", func(t *testing.T) {
		var capturedFilters map[string]interface{}

		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedFilters = filters
				return make([]*models.PaymentWithDetails, 0), 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?month=invalid&building_id=bad", nil)
		router.ServeHTTP(w, req)

		if _, ok := capturedFilters["month"]; ok {
			t.Errorf("invalid month should not be in filters")
		}
		if _, ok := capturedFilters["building_id"]; ok {
			t.Errorf("invalid building_id should not be in filters")
		}
	})
}

// ---------------------------------------------------------------------------
// Pagination Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_Pagination tests pagination behavior
func TestPaymentHandler_Pagination(t *testing.T) {
	t.Run("default pagination values are used", func(t *testing.T) {
		var capturedPage, capturedPageSize int

		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedPage = page
				capturedPageSize = pageSize
				return make([]*models.PaymentWithDetails, 0), 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments", nil)
		router.ServeHTTP(w, req)

		if capturedPage != 1 {
			t.Errorf("default page should be 1, got %d", capturedPage)
		}
		if capturedPageSize != 20 {
			t.Errorf("default pageSize should be 20, got %d", capturedPageSize)
		}
	})

	t.Run("custom pagination values are passed to service", func(t *testing.T) {
		var capturedPage, capturedPageSize int

		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedPage = page
				capturedPageSize = pageSize
				return make([]*models.PaymentWithDetails, 0), 50, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?page=3&page_size=50", nil)
		router.ServeHTTP(w, req)

		if capturedPage != 3 {
			t.Errorf("page should be 3, got %d", capturedPage)
		}
		if capturedPageSize != 50 {
			t.Errorf("pageSize should be 50, got %d", capturedPageSize)
		}
	})

	t.Run("pagination response includes correct metadata", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				return make([]*models.PaymentWithDetails, 0), 150, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?page=2&page_size=50", nil)
		router.ServeHTTP(w, req)

		var resp models.PaymentListResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.Page != 2 {
			t.Errorf("page in response should be 2, got %d", resp.Page)
		}
		if resp.PageSize != 50 {
			t.Errorf("pageSize in response should be 50, got %d", resp.PageSize)
		}
		if resp.Total != 150 {
			t.Errorf("total in response should be 150, got %d", resp.Total)
		}
		if resp.TotalPages != 3 {
			t.Errorf("totalPages should be 3, got %d", resp.TotalPages)
		}
	})
}

// ---------------------------------------------------------------------------
// Report Endpoint Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_DateRangeParsing tests date range query parameter parsing
func TestPaymentHandler_DateRangeParsing(t *testing.T) {
	t.Run("valid date range is parsed correctly", func(t *testing.T) {
		reportCalled := false

		svc := &mockPaymentService{
			generateBuildingReportFn: func(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
				reportCalled = true
				return &models.BuildingPaymentReport{}, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			"/payments/building/1/report?start_date=2026-01-01&end_date=2026-12-31",
			nil,
		)
		router.ServeHTTP(w, req)

		if w.Code == http.StatusBadRequest {
			t.Errorf("valid dates should not return 400: %v", w.Body.String())
		}
		if !reportCalled {
			t.Errorf("report generation function should have been called")
		}
	})

	t.Run("invalid date format returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			"/payments/building/1/report?start_date=01-01-2026",
			nil,
		)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid date should return 400, got %d", w.Code)
		}
	})
}

// ---------------------------------------------------------------------------
// Search Endpoint Tests
// ---------------------------------------------------------------------------

// TestPaymentHandler_SearchLeases tests lease search functionality
func TestPaymentHandler_SearchLeases(t *testing.T) {
	t.Run("missing query parameter returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/search", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("missing query parameter should return 400, got %d", w.Code)
		}
	})

	t.Run("search query is passed to service", func(t *testing.T) {
		var capturedQuery string
		var capturedOrgID int

		svc := &mockPaymentService{
			searchLeasesFn: func(orgID int, query string) (*models.LeaseSearchResponse, error) {
				capturedOrgID = orgID
				capturedQuery = query
				return &models.LeaseSearchResponse{Results: make([]*models.LeaseSearchResult, 0), Total: 0}, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/search?q=john+doe", nil)
		router.ServeHTTP(w, req)

		if capturedQuery != "john doe" {
			t.Errorf("query: got %q, want %q", capturedQuery, "john doe")
		}
		if capturedOrgID != 1 {
			t.Errorf("orgID should be 1, got %d", capturedOrgID)
		}
	})

	t.Run("search returns results", func(t *testing.T) {
		results := []*models.LeaseSearchResult{
			{LeaseID: 1, TenantName: "John Doe", PropertyName: "Test Property"},
		}

		svc := &mockPaymentService{
			searchLeasesFn: func(orgID int, query string) (*models.LeaseSearchResponse, error) {
				return &models.LeaseSearchResponse{Results: results, Total: 1}, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/search?q=john", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var resp models.LeaseSearchResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Total != 1 {
			t.Errorf("expected 1 result, got %d", resp.Total)
		}
	})
}
