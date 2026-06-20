package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
)

// ReportHandler handles HTTP requests for report operations
type ReportHandler struct {
	reportService interfaces.ReportServiceInterface
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(reportService interfaces.ReportServiceInterface) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GetFinancialLedger handles GET /reports/ledger
func (h *ReportHandler) GetFinancialLedger(c *gin.Context) {
	orgID := c.GetInt("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	filters := map[string]interface{}{
		"organization_id": orgID,
	}

	// Apply optional filters
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if month := c.Query("month"); month != "" {
		if m, err := strconv.Atoi(month); err == nil {
			filters["month"] = m
		}
	}
	if year := c.Query("year"); year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			filters["year"] = y
		}
	}

	report, err := h.reportService.FinancialLedgerReport(orgID, filters, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetCollectionSummary handles GET /reports/collection-summary
func (h *ReportHandler) GetCollectionSummary(c *gin.Context) {
	orgID := c.GetInt("org_id")

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.CollectionSummaryReport(orgID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetPaymentAnalysis handles GET /reports/payment-analysis
func (h *ReportHandler) GetPaymentAnalysis(c *gin.Context) {
	orgID := c.GetInt("org_id")

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.PaymentAnalysisReport(orgID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetTenantSummary handles GET /reports/tenant-summary
func (h *ReportHandler) GetTenantSummary(c *gin.Context) {
	orgID := c.GetInt("org_id")

	report, err := h.reportService.TenantSummaryReport(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetPropertyAnalytics handles GET /reports/property-analytics
func (h *ReportHandler) GetPropertyAnalytics(c *gin.Context) {
	orgID := c.GetInt("org_id")

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.PropertyAnalyticsReport(orgID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetDashboardMetrics handles GET /reports/dashboard-metrics
func (h *ReportHandler) GetDashboardMetrics(c *gin.Context) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	orgID := c.GetInt("org_id")

	collectionReport, err := h.reportService.CollectionSummaryReport(orgID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paymentReport, err := h.reportService.PaymentAnalysisReport(orgID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics := gin.H{
		"collection_summary": collectionReport,
		"payment_analysis":   paymentReport,
		"generated_at":       time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}
