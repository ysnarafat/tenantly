package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// ---------------------------------------------------------------------------
// Mock implementation
// ---------------------------------------------------------------------------

type mockPaymentService struct {
	createPaymentFn          func(req *models.CreatePaymentRequest, userID int) (*models.Payment, error)
	getPaymentFn             func(id, orgID int) (*models.PaymentWithDetails, error)
	updatePaymentFn          func(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error)
	getPaymentsFn            func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error)
	getPaymentsByBuildingFn  func(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error)
	getPaymentsByPropertyFn  func(propertyID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error)
	generateBuildingReportFn func(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error)
	generatePropertyReportFn func(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error)
	getDashboardSummaryFn    func() (*models.DashboardSummary, error)
	processBulkFn            func(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error)
	getAnalyticsFn           func(buildingID int, period string) (*models.BuildingPaymentAnalytics, error)
	canAccessPaymentFn       func(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool
	logAccessFn              func(userID int, action string, paymentID int, allowed bool)
	searchLeasesFn           func(orgID int, query string) (*models.LeaseSearchResponse, error)
}

func (m *mockPaymentService) CreatePayment(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
	if m.createPaymentFn == nil {
		return nil, errors.New("create payment not mocked")
	}
	return m.createPaymentFn(req, userID)
}

func (m *mockPaymentService) GetPayment(id, orgID int) (*models.PaymentWithDetails, error) {
	if m.getPaymentFn == nil {
		return &models.PaymentWithDetails{
			Payment: models.Payment{
				ID: id, TenantID: 1, BuildingID: 1, PropertyID: 1, UnitID: 1,
			},
		}, nil
	}
	return m.getPaymentFn(id, orgID)
}

func (m *mockPaymentService) UpdatePayment(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error) {
	if m.updatePaymentFn == nil {
		return nil, errors.New("update payment not mocked")
	}
	return m.updatePaymentFn(id, req, userID, orgID)
}

func (m *mockPaymentService) GetPayments(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	if m.getPaymentsFn == nil {
		return make([]*models.PaymentWithDetails, 0), 0, nil
	}
	return m.getPaymentsFn(page, pageSize, filters)
}

func (m *mockPaymentService) GetPaymentsByBuilding(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	if m.getPaymentsByBuildingFn == nil {
		return make([]*models.PaymentWithDetails, 0), 0, nil
	}
	return m.getPaymentsByBuildingFn(buildingID, page, pageSize, filters)
}

func (m *mockPaymentService) GetPaymentsByProperty(propertyID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	if m.getPaymentsByPropertyFn == nil {
		return make([]*models.PaymentWithDetails, 0), 0, nil
	}
	return m.getPaymentsByPropertyFn(propertyID, page, pageSize, filters)
}

func (m *mockPaymentService) GenerateBuildingPaymentReport(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
	if m.generateBuildingReportFn == nil {
		return nil, errors.New("generate building report not mocked")
	}
	return m.generateBuildingReportFn(buildingID, orgID, startDate, endDate)
}

func (m *mockPaymentService) GeneratePropertyPaymentReport(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error) {
	if m.generatePropertyReportFn == nil {
		return nil, errors.New("generate property report not mocked")
	}
	return m.generatePropertyReportFn(propertyID, orgID, startDate, endDate)
}

func (m *mockPaymentService) GetDashboardSummaryWithBuildingContext(orgID int) (*models.DashboardSummary, error) {
	if m.getDashboardSummaryFn == nil {
		return nil, errors.New("get dashboard summary not mocked")
	}
	return m.getDashboardSummaryFn()
}

func (m *mockPaymentService) ProcessBulkPayments(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
	if m.processBulkFn == nil {
		return make([]*models.Payment, 0), []error{}
	}
	return m.processBulkFn(requests, userID)
}

func (m *mockPaymentService) GetPaymentAnalyticsByBuilding(buildingID int, period string) (*models.BuildingPaymentAnalytics, error) {
	if m.getAnalyticsFn == nil {
		return nil, errors.New("get analytics not mocked")
	}
	return m.getAnalyticsFn(buildingID, period)
}

func (m *mockPaymentService) CanUserAccessPayment(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
	if m.canAccessPaymentFn == nil {
		return true // Allow all by default in tests
	}
	return m.canAccessPaymentFn(userID, userRole, payment, userOrgID)
}

func (m *mockPaymentService) LogPaymentAccess(userID int, action string, paymentID int, allowed bool) {
	if m.logAccessFn != nil {
		m.logAccessFn(userID, action, paymentID, allowed)
	}
}

func (m *mockPaymentService) SearchLeases(orgID int, query string) (*models.LeaseSearchResponse, error) {
	if m.searchLeasesFn == nil {
		return &models.LeaseSearchResponse{Results: make([]*models.LeaseSearchResult, 0), Total: 0}, nil
	}
	return m.searchLeasesFn(orgID, query)
}

func (m *mockPaymentService) GenerateMonthlyPayments(req *models.GenerateMonthlyPaymentsRequest, orgID, userID int) (*models.GenerateMonthlyPaymentsResult, error) {
	return &models.GenerateMonthlyPaymentsResult{}, nil
}

// Ensure the mock satisfies the interface at compile time.
var _ interfaces.PaymentServiceInterface = (*mockPaymentService)(nil)

// ---------------------------------------------------------------------------
// Router helper
// ---------------------------------------------------------------------------

func setupPaymentTestRouter(svc interfaces.PaymentServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewPaymentHandler(svc)
	r := gin.New()

	// Middleware to inject auth context values used by handlers.
	r.Use(func(c *gin.Context) {
		c.Set("userID", 1)
		c.Set("org_id", 1)
		c.Next()
	})

	payments := r.Group("/payments")
	payments.GET("", handler.GetPayments)
	payments.POST("", handler.CreatePayment)
	payments.POST("/bulk", handler.BulkCreatePayments)
	payments.GET("/search", handler.SearchLeases)
	payments.GET("/:id", handler.GetPayment)
	payments.PUT("/:id", handler.UpdatePayment)
	payments.GET("/building/:building_id/report", handler.GetBuildingPaymentReport)
	payments.GET("/property/:property_id/report", handler.GetPropertyPaymentReport)

	r.GET("/dashboard/summary", handler.GetDashboardSummary)
	return r
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func checkStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status: got %d, want %d", got, want)
	}
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

func toJSON(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ---------------------------------------------------------------------------
// Sample data builders
// ---------------------------------------------------------------------------

func samplePayment() *models.Payment {
	return &models.Payment{
		ID:             1,
		UnitID:         10,
		TenantID:       20,
		BuildingID:     30,
		PropertyID:     40,
		OrganizationID: 1,
		Month:          5,
		Year:           2026,
		AmountDue:      5000.00,
		AmountPaid:     5000.00,
		Status:         models.PaymentStatusPaid,
		PaymentMethod:  "Bank Transfer",
	}
}

func samplePaymentWithDetails() *models.PaymentWithDetails {
	return &models.PaymentWithDetails{
		Payment:      *samplePayment(),
		PropertyName: "Green Residency",
		BuildingName: "Block A",
		BuildingCode: "BLK-A",
		UnitNumber:   "101",
		UnitType:     "1BHK",
		TenantName:   "John Doe",
	}
}

func sampleCreatePaymentRequest() map[string]interface{} {
	return map[string]interface{}{
		"unit_id":     10,
		"tenant_id":   20,
		"building_id": 30,
		"property_id": 40,
		"month":       5,
		"year":        2026,
		"amount_due":  5000.00,
		"due_date":    "2026-05-31",
	}
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_CreatePayment
// ---------------------------------------------------------------------------

func TestPaymentHandler_CreatePayment(t *testing.T) {
	t.Run("success returns 201 with created payment", func(t *testing.T) {
		svc := &mockPaymentService{
			createPaymentFn: func(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
				return samplePayment(), nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", toJSON(t, sampleCreatePaymentRequest()))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusCreated)

		var got models.Payment
		decodeBody(t, w, &got)
		if got.ID != 1 {
			t.Errorf("payment ID: got %d, want 1", got.ID)
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString("{bad json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("missing required fields returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		// Empty object â€” all required fields absent.
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", toJSON(t, map[string]interface{}{}))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("service error returns 400", func(t *testing.T) {
		svc := &mockPaymentService{
			createPaymentFn: func(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
				return nil, errors.New("duplicate payment")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", toJSON(t, sampleCreatePaymentRequest()))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("org_id is set from context not from request body", func(t *testing.T) {
		const contextOrgID = 1

		var capturedOrgID int
		svc := &mockPaymentService{
			createPaymentFn: func(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
				capturedOrgID = req.OrganizationID
				return samplePayment(), nil
			},
		}
		router := setupPaymentTestRouter(svc)

		// Include an organization_id in the body that differs from context.
		body := sampleCreatePaymentRequest()
		body["organization_id"] = 999

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", toJSON(t, body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusCreated)
		if capturedOrgID != contextOrgID {
			t.Errorf("org_id: got %d, want %d (from context)", capturedOrgID, contextOrgID)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_GetPayment
// ---------------------------------------------------------------------------

func TestPaymentHandler_GetPayment(t *testing.T) {
	t.Run("success returns 200 with payment details", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentFn: func(id, orgID int) (*models.PaymentWithDetails, error) {
				return samplePaymentWithDetails(), nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/1", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.PaymentWithDetails
		decodeBody(t, w, &got)
		if got.ID != 1 {
			t.Errorf("payment ID: got %d, want 1", got.ID)
		}
		if got.TenantName != "John Doe" {
			t.Errorf("tenant name: got %q, want %q", got.TenantName, "John Doe")
		}
	})

	t.Run("non-numeric ID returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/abc", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentFn: func(id, orgID int) (*models.PaymentWithDetails, error) {
				return nil, errors.New("payment not found")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/999", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusNotFound)
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_UpdatePayment
// ---------------------------------------------------------------------------

func TestPaymentHandler_UpdatePayment(t *testing.T) {
	amountPaid := 4500.00
	statusPaid := models.PaymentStatusPaid

	validUpdateBody := map[string]interface{}{
		"amount_paid": amountPaid,
		"status":      string(statusPaid),
	}

	t.Run("success returns 200 with updated payment", func(t *testing.T) {
		updated := samplePayment()
		updated.AmountPaid = amountPaid
		updated.Status = statusPaid

		// Sample existing payment for access verification
		existing := samplePaymentWithDetails()

		svc := &mockPaymentService{
			getPaymentFn: func(id, orgID int) (*models.PaymentWithDetails, error) {
				return existing, nil
			},
			updatePaymentFn: func(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error) {
				return updated, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/payments/1", toJSON(t, validUpdateBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.Payment
		decodeBody(t, w, &got)
		if got.AmountPaid != amountPaid {
			t.Errorf("amount_paid: got %.2f, want %.2f", got.AmountPaid, amountPaid)
		}
	})

	t.Run("non-numeric ID returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/payments/xyz", toJSON(t, validUpdateBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/payments/1", bytes.NewBufferString("{bad"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("service error returns 400", func(t *testing.T) {
		svc := &mockPaymentService{
			updatePaymentFn: func(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error) {
				return nil, errors.New("payment already finalised")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/payments/1", toJSON(t, validUpdateBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_GetPayments
// ---------------------------------------------------------------------------

func TestPaymentHandler_GetPayments(t *testing.T) {
	makePaymentsList := func(n int) []*models.PaymentWithDetails {
		list := make([]*models.PaymentWithDetails, n)
		for i := range list {
			p := samplePaymentWithDetails()
			p.ID = i + 1
			list[i] = p
		}
		return list
	}

	t.Run("no filters calls GetPayments with org_id in filters", func(t *testing.T) {
		var capturedFilters map[string]interface{}
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedFilters = filters
				return makePaymentsList(2), 2, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		if capturedFilters["organization_id"] != 1 {
			t.Errorf("organization_id filter: got %v, want 1", capturedFilters["organization_id"])
		}
	})

	t.Run("building_id filter routes to GetPaymentsByBuilding", func(t *testing.T) {
		var calledWithBuildingID int
		svc := &mockPaymentService{
			getPaymentsByBuildingFn: func(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				calledWithBuildingID = buildingID
				return makePaymentsList(1), 1, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?building_id=5", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if calledWithBuildingID != 5 {
			t.Errorf("building_id: got %d, want 5", calledWithBuildingID)
		}
	})

	t.Run("property_id filter routes to GetPaymentsByProperty", func(t *testing.T) {
		var calledWithPropertyID int
		svc := &mockPaymentService{
			getPaymentsByPropertyFn: func(propertyID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				calledWithPropertyID = propertyID
				return makePaymentsList(1), 1, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?property_id=3", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if calledWithPropertyID != 3 {
			t.Errorf("property_id: got %d, want 3", calledWithPropertyID)
		}
	})

	t.Run("status month year filters are populated in filters map", func(t *testing.T) {
		var capturedFilters map[string]interface{}
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				capturedFilters = filters
				return nil, 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?status=Paid&month=5&year=2026", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		if capturedFilters["status"] != "Paid" {
			t.Errorf("status filter: got %v, want Paid", capturedFilters["status"])
		}
		if capturedFilters["month"] != 5 {
			t.Errorf("month filter: got %v, want 5", capturedFilters["month"])
		}
		if capturedFilters["year"] != 2026 {
			t.Errorf("year filter: got %v, want 2026", capturedFilters["year"])
		}
	})

	t.Run("non-numeric building_id query param is ignored", func(t *testing.T) {
		// When building_id cannot be parsed the filter is not set, so
		// the request falls through to GetPayments (not GetPaymentsByBuilding).
		generalCalled := false
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				generalCalled = true
				return nil, 0, nil
			},
			getPaymentsByBuildingFn: func(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				t.Error("GetPaymentsByBuilding should not have been called")
				return nil, 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?building_id=notanumber", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if !generalCalled {
			t.Error("expected GetPayments to be called when building_id is non-numeric")
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				return nil, 0, errors.New("db connection lost")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("response contains correct pagination fields", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				return makePaymentsList(5), 45, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments?page=2&page_size=20", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var resp models.PaymentListResponse
		decodeBody(t, w, &resp)

		if resp.Page != 2 {
			t.Errorf("page: got %d, want 2", resp.Page)
		}
		if resp.PageSize != 20 {
			t.Errorf("page_size: got %d, want 20", resp.PageSize)
		}
		if resp.Total != 45 {
			t.Errorf("total: got %d, want 45", resp.Total)
		}
		// ceil(45/20) = 3
		if resp.TotalPages != 3 {
			t.Errorf("total_pages: got %d, want 3", resp.TotalPages)
		}
	})

	t.Run("total_pages is 1 when total is zero", func(t *testing.T) {
		svc := &mockPaymentService{
			getPaymentsFn: func(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
				return nil, 0, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var resp models.PaymentListResponse
		decodeBody(t, w, &resp)
		if resp.TotalPages != 1 {
			t.Errorf("total_pages: got %d, want 1 when total=0", resp.TotalPages)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_GetDashboardSummary
// ---------------------------------------------------------------------------

func TestPaymentHandler_GetDashboardSummary(t *testing.T) {
	t.Run("success returns 200 with summary", func(t *testing.T) {
		summary := &models.DashboardSummary{
			TotalDue:       100000.00,
			TotalPaid:      80000.00,
			CollectionRate: 80.0,
			PropertyCount:  3,
			BuildingCount:  6,
			UnitCount:      60,
			TenantCount:    55,
		}
		svc := &mockPaymentService{
			getDashboardSummaryFn: func() (*models.DashboardSummary, error) {
				return summary, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.DashboardSummary
		decodeBody(t, w, &got)
		if got.CollectionRate != 80.0 {
			t.Errorf("collection_rate: got %.2f, want 80.00", got.CollectionRate)
		}
		if got.BuildingCount != 6 {
			t.Errorf("building_count: got %d, want 6", got.BuildingCount)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockPaymentService{
			getDashboardSummaryFn: func() (*models.DashboardSummary, error) {
				return nil, errors.New("aggregation failed")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_GetBuildingPaymentReport
// ---------------------------------------------------------------------------

func TestPaymentHandler_GetBuildingPaymentReport(t *testing.T) {
	sampleReport := func() *models.BuildingPaymentReport {
		return &models.BuildingPaymentReport{
			BuildingID:   1,
			BuildingName: "Block A",
			BuildingCode: "BLK-A",
			PropertyID:   40,
			PropertyName: "Green Residency",
			ReportPeriod: "2026-04-01 to 2026-05-01",
			Payments:     []*models.PaymentWithDetails{samplePaymentWithDetails()},
			GeneratedAt:  time.Now(),
		}
	}

	t.Run("success returns 200 with report", func(t *testing.T) {
		svc := &mockPaymentService{
			generateBuildingReportFn: func(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
				return sampleReport(), nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/building/1/report?start_date=2026-04-01&end_date=2026-05-01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.BuildingPaymentReport
		decodeBody(t, w, &got)
		if got.BuildingID != 1 {
			t.Errorf("building_id: got %d, want 1", got.BuildingID)
		}
	})

	t.Run("non-numeric building_id returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/building/xyz/report", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("invalid start_date format returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/building/1/report?start_date=31-05-2026", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("invalid end_date format returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/building/1/report?start_date=2026-04-01&end_date=not-a-date", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("defaults are used when dates are omitted", func(t *testing.T) {
		var capturedStart, capturedEnd time.Time
		svc := &mockPaymentService{
			generateBuildingReportFn: func(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
				capturedStart = startDate
				capturedEnd = endDate
				return sampleReport(), nil
			},
		}
		router := setupPaymentTestRouter(svc)

		now := time.Now()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/building/1/report", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		// parseDateRange formats now-1month as YYYY-MM-DD then re-parses it,
		// so the result is midnight of that day. Compare against the truncated
		// expected value and allow Â±1 day to handle month-boundary rounding.
		expectedStartDate := now.AddDate(0, -1, 0).Format("2006-01-02")
		expectedStart, _ := time.Parse("2006-01-02", expectedStartDate)
		diff := capturedStart.Sub(expectedStart)
		if diff < -25*time.Hour || diff > 25*time.Hour {
			t.Errorf("default start_date unexpected: got %v, want ~%v (diff %v)", capturedStart, expectedStart, diff)
		}

		// Default end = today midnight + 23h59m59s (added by parseDateRange).
		// It should be strictly after the start and within ~25 hours of now.
		if !capturedEnd.After(capturedStart) {
			t.Errorf("default end_date should be after start_date; start=%v end=%v", capturedStart, capturedEnd)
		}
		endDiff := capturedEnd.Sub(now)
		if endDiff < -2*time.Minute || endDiff > 49*time.Hour {
			t.Errorf("default end_date out of expected range: %v", endDiff)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockPaymentService{
			generateBuildingReportFn: func(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
				return nil, errors.New("report generation failed")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/building/1/report", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_GetPropertyPaymentReport
// ---------------------------------------------------------------------------

func TestPaymentHandler_GetPropertyPaymentReport(t *testing.T) {
	samplePropertyReport := func() *models.PropertyPaymentReport {
		return &models.PropertyPaymentReport{
			PropertyID:   40,
			PropertyName: "Green Residency",
			PropertyCode: "GR-01",
			ReportPeriod: "2026-04-01 to 2026-05-01",
			GeneratedAt:  time.Now(),
		}
	}

	t.Run("success returns 200 with report", func(t *testing.T) {
		svc := &mockPaymentService{
			generatePropertyReportFn: func(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error) {
				return samplePropertyReport(), nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/property/40/report?start_date=2026-04-01&end_date=2026-05-01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.PropertyPaymentReport
		decodeBody(t, w, &got)
		if got.PropertyID != 40 {
			t.Errorf("property_id: got %d, want 40", got.PropertyID)
		}
	})

	t.Run("non-numeric property_id returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/property/notanumber/report", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("invalid date format returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/property/40/report?start_date=2026/04/01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockPaymentService{
			generatePropertyReportFn: func(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error) {
				return nil, errors.New("property report failed")
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/property/40/report", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})
}

// ---------------------------------------------------------------------------
// TestPaymentHandler_BulkCreatePayments
// ---------------------------------------------------------------------------

func TestPaymentHandler_BulkCreatePayments(t *testing.T) {
	makeBulkBody := func(n int) []map[string]interface{} {
		reqs := make([]map[string]interface{}, n)
		for i := range reqs {
			reqs[i] = sampleCreatePaymentRequest()
			reqs[i]["unit_id"] = i + 1
		}
		return reqs
	}

	t.Run("all success returns 207 with created payments", func(t *testing.T) {
		payments := []*models.Payment{samplePayment(), samplePayment()}

		svc := &mockPaymentService{
			processBulkFn: func(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
				return payments, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/bulk", toJSON(t, makeBulkBody(2)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusMultiStatus)

		var resp map[string]interface{}
		decodeBody(t, w, &resp)

		successVal, ok := resp["success"].(float64)
		if !ok {
			t.Fatalf("success field missing or wrong type: %T", resp["success"])
		}
		if int(successVal) != 2 {
			t.Errorf("success count: got %d, want 2", int(successVal))
		}

		failedVal, ok := resp["failed"].(float64)
		if !ok {
			t.Fatalf("failed field missing or wrong type: %T", resp["failed"])
		}
		if int(failedVal) != 0 {
			t.Errorf("failed count: got %d, want 0", int(failedVal))
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		svc := &mockPaymentService{}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/bulk", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("partial failures returns 207 with errors in response", func(t *testing.T) {
		svc := &mockPaymentService{
			processBulkFn: func(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
				// 1 success, 1 failure out of 2 requests.
				return []*models.Payment{samplePayment()}, []error{errors.New("unit not found")}
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/bulk", toJSON(t, makeBulkBody(2)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusMultiStatus)

		var resp map[string]interface{}
		decodeBody(t, w, &resp)

		successVal, ok := resp["success"].(float64)
		if !ok {
			t.Fatalf("success field missing or wrong type: %T", resp["success"])
		}
		if int(successVal) != 1 {
			t.Errorf("success count: got %d, want 1", int(successVal))
		}

		failedVal, ok := resp["failed"].(float64)
		if !ok {
			t.Fatalf("failed field missing or wrong type: %T", resp["failed"])
		}
		if int(failedVal) != 1 {
			t.Errorf("failed count: got %d, want 1", int(failedVal))
		}

		errs, ok := resp["errors"].([]interface{})
		if !ok {
			t.Fatalf("errors field missing or wrong type: %T", resp["errors"])
		}
		if len(errs) != 1 {
			t.Errorf("errors length: got %d, want 1", len(errs))
		}
	})

	t.Run("org_id is injected from context into every request", func(t *testing.T) {
		var capturedOrgIDs []int
		svc := &mockPaymentService{
			processBulkFn: func(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
				for _, r := range requests {
					capturedOrgIDs = append(capturedOrgIDs, r.OrganizationID)
				}
				return []*models.Payment{samplePayment(), samplePayment()}, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		body := makeBulkBody(2)
		// Include a wrong org_id in the body to ensure it gets overridden.
		for i := range body {
			body[i]["organization_id"] = 999
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/bulk", toJSON(t, body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusMultiStatus)

		for i, orgID := range capturedOrgIDs {
			if orgID != 1 {
				t.Errorf("request[%d] org_id: got %d, want 1 (from context)", i, orgID)
			}
		}
	})

	t.Run("total field reflects number of input requests", func(t *testing.T) {
		svc := &mockPaymentService{
			processBulkFn: func(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
				return []*models.Payment{}, nil
			},
		}
		router := setupPaymentTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments/bulk", toJSON(t, makeBulkBody(3)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusMultiStatus)

		var resp map[string]interface{}
		decodeBody(t, w, &resp)

		totalVal, ok := resp["total"].(float64)
		if !ok {
			t.Fatalf("total field missing or wrong type: %T", resp["total"])
		}
		if int(totalVal) != 3 {
			t.Errorf("total: got %d, want 3", int(totalVal))
		}
	})
}
