package handlers

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
)

const dashboardCacheTTL = 5 * time.Minute

type dashboardCacheEntry struct {
	payload   gin.H
	expiresAt time.Time
}

// ReportHandler handles HTTP requests for report operations
type ReportHandler struct {
	reportService  interfaces.ReportServiceInterface
	dashboardCache map[int]*dashboardCacheEntry
	cacheMu        sync.Mutex
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(reportService interfaces.ReportServiceInterface) *ReportHandler {
	return &ReportHandler{
		reportService:  reportService,
		dashboardCache: make(map[int]*dashboardCacheEntry),
	}
}

// GetFinancialLedger handles GET /reports/ledger
func (h *ReportHandler) GetFinancialLedger(c *gin.Context) {
	orgID := c.GetInt("org_id")
	if orgID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing organisation context"})
		return
	}

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
		log.Printf("ERROR GetFinancialLedger org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate ledger report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetCollectionSummary handles GET /reports/collection-summary
func (h *ReportHandler) GetCollectionSummary(c *gin.Context) {
	orgID := c.GetInt("org_id")
	if orgID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing organisation context"})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.CollectionSummaryReport(orgID, startDate, endDate)
	if err != nil {
		log.Printf("ERROR GetCollectionSummary org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate collection summary"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetPaymentAnalysis handles GET /reports/payment-analysis
func (h *ReportHandler) GetPaymentAnalysis(c *gin.Context) {
	orgID := c.GetInt("org_id")
	if orgID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing organisation context"})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.PaymentAnalysisReport(orgID, startDate, endDate)
	if err != nil {
		log.Printf("ERROR GetPaymentAnalysis org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate payment analysis"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetTenantSummary handles GET /reports/tenant-summary
func (h *ReportHandler) GetTenantSummary(c *gin.Context) {
	orgID := c.GetInt("org_id")
	if orgID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing organisation context"})
		return
	}

	report, err := h.reportService.TenantSummaryReport(orgID)
	if err != nil {
		log.Printf("ERROR GetTenantSummary org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tenant summary"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetPropertyAnalytics handles GET /reports/property-analytics
func (h *ReportHandler) GetPropertyAnalytics(c *gin.Context) {
	orgID := c.GetInt("org_id")
	if orgID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing organisation context"})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.PropertyAnalyticsReport(orgID, startDate, endDate)
	if err != nil {
		log.Printf("ERROR GetPropertyAnalytics org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate property analytics"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetDashboardMetrics handles GET /reports/dashboard-metrics
func (h *ReportHandler) GetDashboardMetrics(c *gin.Context) {
	orgID := c.GetInt("org_id")
	if orgID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing organisation context"})
		return
	}

	h.cacheMu.Lock()
	if entry, ok := h.dashboardCache[orgID]; ok && time.Now().Before(entry.expiresAt) {
		h.cacheMu.Unlock()
		c.JSON(http.StatusOK, entry.payload)
		return
	}
	h.cacheMu.Unlock()

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	collectionReport, err := h.reportService.CollectionSummaryReport(orgID, startDate, endDate)
	if err != nil {
		log.Printf("ERROR GetDashboardMetrics collection org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard metrics"})
		return
	}

	paymentReport, err := h.reportService.PaymentAnalysisReport(orgID, startDate, endDate)
	if err != nil {
		log.Printf("ERROR GetDashboardMetrics payment org=%d: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard metrics"})
		return
	}

	payload := gin.H{
		"collection_summary": collectionReport,
		"payment_analysis":   paymentReport,
		"generated_at":       now,
	}

	h.cacheMu.Lock()
	h.dashboardCache[orgID] = &dashboardCacheEntry{payload: payload, expiresAt: now.Add(dashboardCacheTTL)}
	h.cacheMu.Unlock()

	c.JSON(http.StatusOK, payload)
}
