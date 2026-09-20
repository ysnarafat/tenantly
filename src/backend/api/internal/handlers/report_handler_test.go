package handlers

import (
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

type mockReportService struct {
	financialLedgerReportFn   func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error)
	collectionSummaryReportFn func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error)
	paymentAnalysisReportFn   func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error)
	tenantSummaryReportFn     func(orgID int) (*models.TenantSummaryReport, error)
	propertyAnalyticsReportFn func(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error)
}

func (m *mockReportService) FinancialLedgerReport(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
	if m.financialLedgerReportFn == nil {
		return &models.FinancialLedgerReport{OrganizationID: orgID}, nil
	}
	return m.financialLedgerReportFn(orgID, filters, limit, offset)
}

func (m *mockReportService) CollectionSummaryReport(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
	if m.collectionSummaryReportFn == nil {
		return &models.CollectionSummaryReport{OrganizationID: orgID}, nil
	}
	return m.collectionSummaryReportFn(orgID, startDate, endDate)
}

func (m *mockReportService) PaymentAnalysisReport(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
	if m.paymentAnalysisReportFn == nil {
		return &models.PaymentAnalysisReport{OrganizationID: orgID}, nil
	}
	return m.paymentAnalysisReportFn(orgID, startDate, endDate)
}

func (m *mockReportService) TenantSummaryReport(orgID int) (*models.TenantSummaryReport, error) {
	if m.tenantSummaryReportFn == nil {
		return &models.TenantSummaryReport{OrganizationID: orgID}, nil
	}
	return m.tenantSummaryReportFn(orgID)
}

func (m *mockReportService) PropertyAnalyticsReport(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error) {
	if m.propertyAnalyticsReportFn == nil {
		return &models.PropertyAnalyticsReport{OrganizationID: orgID}, nil
	}
	return m.propertyAnalyticsReportFn(orgID, startDate, endDate)
}

// Ensure the mock satisfies the interface at compile time.
var _ interfaces.ReportServiceInterface = (*mockReportService)(nil)

// ---------------------------------------------------------------------------
// Router helper
// ---------------------------------------------------------------------------

func setupReportTestRouter(svc interfaces.ReportServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewReportHandler(svc)
	r := gin.New()

	// Middleware to inject auth context values used by handlers.
	r.Use(func(c *gin.Context) {
		c.Set("org_id", 1)
		c.Next()
	})

	reports := r.Group("/reports")
	reports.GET("/ledger", handler.GetFinancialLedger)
	reports.GET("/collection-summary", handler.GetCollectionSummary)
	reports.GET("/payment-analysis", handler.GetPaymentAnalysis)
	reports.GET("/tenant-summary", handler.GetTenantSummary)
	reports.GET("/property-analytics", handler.GetPropertyAnalytics)
	reports.GET("/dashboard-metrics", handler.GetDashboardMetrics)

	return r
}

// ---------------------------------------------------------------------------
// Sample data builders
// ---------------------------------------------------------------------------

func sampleFinancialLedgerReport() *models.FinancialLedgerReport {
	return &models.FinancialLedgerReport{
		OrganizationID: 1,
		Payments:       []*models.PaymentWithDetails{},
		Total:          10,
		TotalDue:       100000,
		TotalPaid:      80000,
		TotalPending:   20000,
		TotalOverdue:   5000,
		CollectionRate: 80.0,
		GeneratedAt:    time.Now(),
	}
}

func sampleCollectionSummaryReport() *models.CollectionSummaryReport {
	return &models.CollectionSummaryReport{
		OrganizationID: 1,
		CollectionRate: 75.5,
		TotalDue:       200000,
		TotalCollected: 151000,
		TotalPending:   49000,
		TotalOverdue:   10000,
		AgingBuckets:   map[string]float64{"0-30": 5000, "31-60": 3000, "60+": 2000},
		MonthlyTrend:   []*models.MonthlyCollectionTrend{},
		ReportPeriod:   "2026-05-01 to 2026-06-01",
		GeneratedAt:    time.Now(),
	}
}

func samplePaymentAnalysisReport() *models.PaymentAnalysisReport {
	return &models.PaymentAnalysisReport{
		OrganizationID:     1,
		PaymentMethods:     map[string]int64{"Bank Transfer": 5, "Cash": 3},
		StatusDistribution: map[string]int64{"Paid": 7, "Pending": 1},
		DailyTrend:         map[string]int64{"2026-05-01": 2, "2026-05-02": 3},
		TotalPayments:      8,
		ReportPeriod:       "2026-05-01 to 2026-06-01",
		GeneratedAt:        time.Now(),
	}
}

func sampleTenantSummaryReport() *models.TenantSummaryReport {
	return &models.TenantSummaryReport{
		OrganizationID: 1,
		Tenants: []*models.TenantReportEntry{
			{
				TenantID:    1,
				TenantName:  "John Doe",
				PhoneNumber: "01711000000",
				UnitNumber:  "101",
				MonthlyRent: 5000.0,
				LeaseActive: true,
				TotalDue:    5000.0,
				TotalPaid:   5000.0,
				BalanceDue:  0.0,
			},
		},
		Total:         1,
		ActiveTenants: 1,
		GeneratedAt:   time.Now(),
	}
}

func samplePropertyAnalyticsReport() *models.PropertyAnalyticsReport {
	return &models.PropertyAnalyticsReport{
		OrganizationID: 1,
		Properties: []*models.PropertyAnalyticsEntry{
			{
				PropertyID:   1,
				PropertyName: "Green Residency",
				PropertyCode: "GR-01",
				PropertyType: "residential",
			},
		},
		Total:        1,
		ReportPeriod: "2026-05-01 to 2026-06-01",
		GeneratedAt:  time.Now(),
	}
}

// ---------------------------------------------------------------------------
// TestReportHandler_GetFinancialLedger
// ---------------------------------------------------------------------------

func TestReportHandler_GetFinancialLedger(t *testing.T) {
	t.Run("success returns 200 with ledger report", func(t *testing.T) {
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.FinancialLedgerReport
		decodeBody(t, w, &got)
		if got.OrganizationID != 1 {
			t.Errorf("organization_id: got %d, want 1", got.OrganizationID)
		}
		if got.CollectionRate != 80.0 {
			t.Errorf("collection_rate: got %.2f, want 80.00", got.CollectionRate)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				return nil, errors.New("db query failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("default page and page_size are applied when omitted", func(t *testing.T) {
		var capturedLimit, capturedOffset int
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedLimit = limit
				capturedOffset = offset
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		// Default: page=1, page_size=50 → offset=0, limit=50
		if capturedLimit != 50 {
			t.Errorf("default limit: got %d, want 50", capturedLimit)
		}
		if capturedOffset != 0 {
			t.Errorf("default offset: got %d, want 0", capturedOffset)
		}
	})

	t.Run("explicit page and page_size compute correct offset and limit", func(t *testing.T) {
		var capturedLimit, capturedOffset int
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedLimit = limit
				capturedOffset = offset
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?page=3&page_size=20", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		// page=3, page_size=20 → offset=(3-1)*20=40
		if capturedLimit != 20 {
			t.Errorf("limit: got %d, want 20", capturedLimit)
		}
		if capturedOffset != 40 {
			t.Errorf("offset: got %d, want 40", capturedOffset)
		}
	})

	t.Run("page below 1 is clamped to 1", func(t *testing.T) {
		var capturedOffset int
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedOffset = offset
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?page=-5", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		// page clamped to 1 → offset=0
		if capturedOffset != 0 {
			t.Errorf("offset after page clamp: got %d, want 0", capturedOffset)
		}
	})

	t.Run("page_size above 100 is clamped to 50", func(t *testing.T) {
		var capturedLimit int
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedLimit = limit
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?page_size=200", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		// page_size 200 > 100 → clamped to 50
		if capturedLimit != 50 {
			t.Errorf("clamped limit: got %d, want 50", capturedLimit)
		}
	})

	t.Run("page_size of 0 is clamped to 50", func(t *testing.T) {
		var capturedLimit int
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedLimit = limit
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?page_size=0", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if capturedLimit != 50 {
			t.Errorf("clamped limit: got %d, want 50", capturedLimit)
		}
	})

	t.Run("status filter is forwarded in filters map", func(t *testing.T) {
		var capturedFilters map[string]interface{}
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedFilters = filters
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?status=Paid", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if capturedFilters["status"] != "Paid" {
			t.Errorf("status filter: got %v, want Paid", capturedFilters["status"])
		}
	})

	t.Run("month and year filters are parsed and forwarded as integers", func(t *testing.T) {
		var capturedFilters map[string]interface{}
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedFilters = filters
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?month=5&year=2026", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if capturedFilters["month"] != 5 {
			t.Errorf("month filter: got %v, want 5", capturedFilters["month"])
		}
		if capturedFilters["year"] != 2026 {
			t.Errorf("year filter: got %v, want 2026", capturedFilters["year"])
		}
	})

	t.Run("non-numeric month and year are ignored silently", func(t *testing.T) {
		var capturedFilters map[string]interface{}
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedFilters = filters
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger?month=abc&year=xyz", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if _, ok := capturedFilters["month"]; ok {
			t.Error("month filter should not be set for non-numeric value")
		}
		if _, ok := capturedFilters["year"]; ok {
			t.Error("year filter should not be set for non-numeric value")
		}
	})

	t.Run("org_id from context is always set in filters", func(t *testing.T) {
		var capturedFilters map[string]interface{}
		svc := &mockReportService{
			financialLedgerReportFn: func(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
				capturedFilters = filters
				return sampleFinancialLedgerReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/ledger", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if capturedFilters["organization_id"] != 1 {
			t.Errorf("organization_id filter: got %v, want 1", capturedFilters["organization_id"])
		}
	})
}

// ---------------------------------------------------------------------------
// TestReportHandler_GetCollectionSummary
// ---------------------------------------------------------------------------

func TestReportHandler_GetCollectionSummary(t *testing.T) {
	t.Run("success returns 200 with collection summary", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return sampleCollectionSummaryReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/collection-summary?start_date=2026-05-01&end_date=2026-06-01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.CollectionSummaryReport
		decodeBody(t, w, &got)
		if got.OrganizationID != 1 {
			t.Errorf("organization_id: got %d, want 1", got.OrganizationID)
		}
		if got.CollectionRate != 75.5 {
			t.Errorf("collection_rate: got %.2f, want 75.50", got.CollectionRate)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return nil, errors.New("aggregation failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/collection-summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("malformed start_date returns 400", func(t *testing.T) {
		svc := &mockReportService{}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/collection-summary?start_date=not-a-date", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("malformed end_date returns 400", func(t *testing.T) {
		svc := &mockReportService{}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/collection-summary?start_date=2026-05-01&end_date=31/05/2026", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("default date range is used when dates are omitted", func(t *testing.T) {
		var capturedStart, capturedEnd time.Time
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				capturedStart = startDate
				capturedEnd = endDate
				return sampleCollectionSummaryReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		now := time.Now()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/collection-summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		expectedStartStr := now.AddDate(0, -1, 0).Format("2006-01-02")
		expectedStart, _ := time.Parse("2006-01-02", expectedStartStr)
		diff := capturedStart.Sub(expectedStart)
		if diff < -25*time.Hour || diff > 25*time.Hour {
			t.Errorf("default start_date unexpected: got %v, want ~%v (diff %v)", capturedStart, expectedStart, diff)
		}
		if !capturedEnd.After(capturedStart) {
			t.Errorf("default end_date should be after start_date; start=%v end=%v", capturedStart, capturedEnd)
		}
	})

	t.Run("parsed dates are forwarded to service correctly", func(t *testing.T) {
		var capturedStart, capturedEnd time.Time
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				capturedStart = startDate
				capturedEnd = endDate
				return sampleCollectionSummaryReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/collection-summary?start_date=2026-05-01&end_date=2026-05-31", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		wantStart, _ := time.Parse("2006-01-02", "2026-05-01")
		if !capturedStart.Equal(wantStart) {
			t.Errorf("start_date: got %v, want %v", capturedStart, wantStart)
		}
		if capturedEnd.Before(capturedStart) {
			t.Errorf("end_date should not be before start_date; start=%v end=%v", capturedStart, capturedEnd)
		}
	})
}

// ---------------------------------------------------------------------------
// TestReportHandler_GetPaymentAnalysis
// ---------------------------------------------------------------------------

func TestReportHandler_GetPaymentAnalysis(t *testing.T) {
	t.Run("success returns 200 with payment analysis", func(t *testing.T) {
		svc := &mockReportService{
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				return samplePaymentAnalysisReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/payment-analysis?start_date=2026-05-01&end_date=2026-06-01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.PaymentAnalysisReport
		decodeBody(t, w, &got)
		if got.OrganizationID != 1 {
			t.Errorf("organization_id: got %d, want 1", got.OrganizationID)
		}
		if got.TotalPayments != 8 {
			t.Errorf("total_payments: got %d, want 8", got.TotalPayments)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				return nil, errors.New("analysis query failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/payment-analysis", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("malformed start_date returns 400", func(t *testing.T) {
		svc := &mockReportService{}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/payment-analysis?start_date=not-a-date", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("malformed end_date returns 400", func(t *testing.T) {
		svc := &mockReportService{}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/payment-analysis?start_date=2026-05-01&end_date=2026/06/01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("default date range is used when dates are omitted", func(t *testing.T) {
		var capturedStart, capturedEnd time.Time
		svc := &mockReportService{
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				capturedStart = startDate
				capturedEnd = endDate
				return samplePaymentAnalysisReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		now := time.Now()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/payment-analysis", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		expectedStartStr := now.AddDate(0, -1, 0).Format("2006-01-02")
		expectedStart, _ := time.Parse("2006-01-02", expectedStartStr)
		diff := capturedStart.Sub(expectedStart)
		if diff < -25*time.Hour || diff > 25*time.Hour {
			t.Errorf("default start_date unexpected: got %v, want ~%v (diff %v)", capturedStart, expectedStart, diff)
		}
		if !capturedEnd.After(capturedStart) {
			t.Errorf("default end_date should be after start_date; start=%v end=%v", capturedStart, capturedEnd)
		}
	})
}

// ---------------------------------------------------------------------------
// TestReportHandler_GetTenantSummary
// ---------------------------------------------------------------------------

func TestReportHandler_GetTenantSummary(t *testing.T) {
	t.Run("success returns 200 with tenant summary", func(t *testing.T) {
		svc := &mockReportService{
			tenantSummaryReportFn: func(orgID int) (*models.TenantSummaryReport, error) {
				return sampleTenantSummaryReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/tenant-summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.TenantSummaryReport
		decodeBody(t, w, &got)
		if got.OrganizationID != 1 {
			t.Errorf("organization_id: got %d, want 1", got.OrganizationID)
		}
		if got.Total != 1 {
			t.Errorf("total: got %d, want 1", got.Total)
		}
		if got.ActiveTenants != 1 {
			t.Errorf("active_tenants: got %d, want 1", got.ActiveTenants)
		}
		if len(got.Tenants) != 1 {
			t.Errorf("tenants length: got %d, want 1", len(got.Tenants))
		}
		if got.Tenants[0].TenantName != "John Doe" {
			t.Errorf("tenant name: got %q, want %q", got.Tenants[0].TenantName, "John Doe")
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			tenantSummaryReportFn: func(orgID int) (*models.TenantSummaryReport, error) {
				return nil, errors.New("tenant query failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/tenant-summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("org_id from context is forwarded to service", func(t *testing.T) {
		var capturedOrgID int
		svc := &mockReportService{
			tenantSummaryReportFn: func(orgID int) (*models.TenantSummaryReport, error) {
				capturedOrgID = orgID
				return sampleTenantSummaryReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/tenant-summary", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)
		if capturedOrgID != 1 {
			t.Errorf("org_id: got %d, want 1", capturedOrgID)
		}
	})
}

// ---------------------------------------------------------------------------
// TestReportHandler_GetPropertyAnalytics
// ---------------------------------------------------------------------------

func TestReportHandler_GetPropertyAnalytics(t *testing.T) {
	t.Run("success returns 200 with property analytics", func(t *testing.T) {
		svc := &mockReportService{
			propertyAnalyticsReportFn: func(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error) {
				return samplePropertyAnalyticsReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/property-analytics?start_date=2026-05-01&end_date=2026-06-01", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var got models.PropertyAnalyticsReport
		decodeBody(t, w, &got)
		if got.OrganizationID != 1 {
			t.Errorf("organization_id: got %d, want 1", got.OrganizationID)
		}
		if got.Total != 1 {
			t.Errorf("total: got %d, want 1", got.Total)
		}
		if len(got.Properties) != 1 {
			t.Errorf("properties length: got %d, want 1", len(got.Properties))
		}
		if got.Properties[0].PropertyName != "Green Residency" {
			t.Errorf("property name: got %q, want %q", got.Properties[0].PropertyName, "Green Residency")
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			propertyAnalyticsReportFn: func(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error) {
				return nil, errors.New("property analytics failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/property-analytics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("malformed start_date returns 400", func(t *testing.T) {
		svc := &mockReportService{}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/property-analytics?start_date=not-a-date", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("malformed end_date returns 400", func(t *testing.T) {
		svc := &mockReportService{}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/property-analytics?start_date=2026-05-01&end_date=01-06-2026", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusBadRequest)
	})

	t.Run("default date range is used when dates are omitted", func(t *testing.T) {
		var capturedStart, capturedEnd time.Time
		svc := &mockReportService{
			propertyAnalyticsReportFn: func(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error) {
				capturedStart = startDate
				capturedEnd = endDate
				return samplePropertyAnalyticsReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		now := time.Now()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/property-analytics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		expectedStartStr := now.AddDate(0, -1, 0).Format("2006-01-02")
		expectedStart, _ := time.Parse("2006-01-02", expectedStartStr)
		diff := capturedStart.Sub(expectedStart)
		if diff < -25*time.Hour || diff > 25*time.Hour {
			t.Errorf("default start_date unexpected: got %v, want ~%v (diff %v)", capturedStart, expectedStart, diff)
		}
		if !capturedEnd.After(capturedStart) {
			t.Errorf("default end_date should be after start_date; start=%v end=%v", capturedStart, capturedEnd)
		}
	})
}

// ---------------------------------------------------------------------------
// TestReportHandler_GetDashboardMetrics
// ---------------------------------------------------------------------------

func TestReportHandler_GetDashboardMetrics(t *testing.T) {
	t.Run("success returns 200 with both collection summary and payment analysis", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return sampleCollectionSummaryReport(), nil
			},
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				return samplePaymentAnalysisReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/dashboard-metrics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var resp map[string]interface{}
		decodeBody(t, w, &resp)

		if _, ok := resp["collection_summary"]; !ok {
			t.Error("response missing 'collection_summary' key")
		}
		if _, ok := resp["payment_analysis"]; !ok {
			t.Error("response missing 'payment_analysis' key")
		}
		if _, ok := resp["generated_at"]; !ok {
			t.Error("response missing 'generated_at' key")
		}
	})

	t.Run("collection summary service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return nil, errors.New("collection query failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/dashboard-metrics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("payment analysis service error returns 500", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return sampleCollectionSummaryReport(), nil
			},
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				return nil, errors.New("payment analysis query failed")
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/dashboard-metrics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusInternalServerError)
	})

	t.Run("date range is current-month start to now", func(t *testing.T) {
		var capturedCollectionStart time.Time
		var capturedPaymentStart time.Time

		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				capturedCollectionStart = startDate
				return sampleCollectionSummaryReport(), nil
			},
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				capturedPaymentStart = startDate
				return samplePaymentAnalysisReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		now := time.Now()
		expectedStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/dashboard-metrics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		if !capturedCollectionStart.Equal(expectedStart) {
			t.Errorf("collection start_date: got %v, want %v", capturedCollectionStart, expectedStart)
		}
		if !capturedPaymentStart.Equal(expectedStart) {
			t.Errorf("payment start_date: got %v, want %v", capturedPaymentStart, expectedStart)
		}
	})

	t.Run("collection_summary nested fields are present in response", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return sampleCollectionSummaryReport(), nil
			},
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				return samplePaymentAnalysisReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/dashboard-metrics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var resp map[string]interface{}
		decodeBody(t, w, &resp)

		collectionRaw, ok := resp["collection_summary"]
		if !ok {
			t.Fatal("response missing 'collection_summary' key")
		}
		collectionBytes, _ := json.Marshal(collectionRaw)
		var collection models.CollectionSummaryReport
		if err := json.Unmarshal(collectionBytes, &collection); err != nil {
			t.Fatalf("failed to decode collection_summary: %v", err)
		}
		if collection.CollectionRate != 75.5 {
			t.Errorf("collection_rate inside dashboard metrics: got %.2f, want 75.50", collection.CollectionRate)
		}
	})

	t.Run("payment_analysis nested fields are present in response", func(t *testing.T) {
		svc := &mockReportService{
			collectionSummaryReportFn: func(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
				return sampleCollectionSummaryReport(), nil
			},
			paymentAnalysisReportFn: func(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
				return samplePaymentAnalysisReport(), nil
			},
		}
		router := setupReportTestRouter(svc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reports/dashboard-metrics", nil)
		router.ServeHTTP(w, req)

		checkStatus(t, w.Code, http.StatusOK)

		var resp map[string]interface{}
		decodeBody(t, w, &resp)

		paymentRaw, ok := resp["payment_analysis"]
		if !ok {
			t.Fatal("response missing 'payment_analysis' key")
		}
		paymentBytes, _ := json.Marshal(paymentRaw)
		var payment models.PaymentAnalysisReport
		if err := json.Unmarshal(paymentBytes, &payment); err != nil {
			t.Fatalf("failed to decode payment_analysis: %v", err)
		}
		if payment.TotalPayments != 8 {
			t.Errorf("total_payments inside dashboard metrics: got %d, want 8", payment.TotalPayments)
		}
	})
}
