# Reports & Analytics

Developer reference for the reporting and analytics system in Tenantly.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Backend](#backend)
   - [Models](#models)
   - [Services](#services)
   - [API Endpoints](#api-endpoints)
4. [Frontend](#frontend)
   - [ReportService](#reportservice)
   - [ReportAnalysis Component](#reportanalysis-component)
   - [Report Templates](#report-templates)
5. [Adding a New Report — End-to-End Checklist](#adding-a-new-report--end-to-end-checklist)

---

## Overview

The reporting system covers two distinct concerns:

| Concern | What it does | Entry point |
|---|---|---|
| **Reports** | Structured, filterable exports of payment data for a date range | `GET /api/v1/reports/*` |
| **Analytics** | Aggregated live metrics at building / property / system level | `GET /api/v1/buildings/:id/analytics`, `GET /api/v1/payments/*/report` |

**Who can access:** All authenticated roles (`RequireAnyRole`). Reports are always scoped to the caller's `org_id` from the JWT — no cross-organisation data leaks.

---

## Architecture

```
ReportAnalysis component (Angular)
  │
  ├── ReportService          → /api/v1/reports/*
  │     getFinancialLedger()        → GET /reports/ledger
  │     getCollectionSummary()      → GET /reports/collection-summary
  │     getPaymentAnalysis()        → GET /reports/payment-analysis
  │     getDashboardMetrics()       → GET /reports/dashboard-metrics
  │     getTenantSummary()          → GET /reports/tenant-summary
  │     getPropertyAnalytics()      → GET /reports/property-analytics
  │
  └── PaymentService         → /api/v1/payments/*
        getBuildingReport()         → GET /payments/building/:id/report
        getPropertyReport()         → GET /payments/property/:id/report

                    ▼ HTTP ▼

ReportHandler               (handlers/report_handler.go)
PaymentHandler              (handlers/payment_handler.go)

                    ▼ DI ▼

ReportService               (services/report_service.go)
  └── PaymentRepositoryInterface
        GetWithDetailsAndFilters()
        GetDashboardSummary(orgID)      ← org-scoped
        GetAgingBuckets(orgID)
        GetMonthlyCollectionTrend(orgID, months)
        GetTenantPaymentSummary(orgID)
        GetPaymentAnalyticsByPeriod(orgID, start, end)
  └── PropertyRepositoryInterface
        List(filters, limit, offset)

BuildingAnalyticsService    (services/building_analytics_service.go)
  └── BuildingRepositoryInterface (with in-memory cache)
```

---

## Backend

### Models

#### `models/report.go` — Financial report types

```
FinancialLedgerReport
  organization_id    int
  payments           []*PaymentWithDetails   paginated transaction list
  total              int                     total record count
  total_due          int64
  total_paid         int64
  total_pending      int64
  total_overdue      int64
  collection_rate    float64                 percentage (0–100)
  generated_at       time.Time

CollectionSummaryReport
  organization_id    int
  collection_rate    float64
  total_due          int64
  total_collected    int64
  total_pending      int64
  total_overdue      int64
  aging_buckets      map[string]int64        keys: "current", "30d", "60d", "90d+"
                                            values: outstanding balance (amount_due - amount_paid)
                                            bucketed by months overdue (year*12+month arithmetic)
  monthly_trend      []*MonthlyCollectionTrend   real per-month DB aggregation, last 6 months
  report_period      string                  "YYYY-MM-DD to YYYY-MM-DD"
  generated_at       time.Time

MonthlyCollectionTrend
  month              time.Time
  collection_rate    float64                 computed from amount_collected / amount_due * 100
  amount_due         int64
  amount_collected   int64

PaymentAnalysisReport
  organization_id       int
  payment_methods       map[string]int      e.g. {"cash":12,"bKash":8}  — DB GROUP BY, not in-memory
  status_distribution   map[string]int64    e.g. {"Paid":20,"Due":5}
  daily_trend           map[string]int64    key: "YYYY-MM-DD", uses payment_date if set else created_at
  total_payments        int64
  report_period         string
  generated_at          time.Time

TenantReportEntry
  tenant_id, tenant_name, phone_number, email
  unit_number, building_name, property_name
  lease_start *string, lease_end *string    ISO date or null
  monthly_rent float64
  lease_active bool
  total_due, total_paid, balance_due float64

TenantSummaryReport
  organization_id    int
  tenants            []*TenantReportEntry
  total              int
  active_tenants     int
  generated_at       time.Time

PropertyAnalyticsEntry
  property_id, property_name, property_code, property_type
  payment_stats      interface{}    raw stats from GetPropertyPaymentStats

PropertyAnalyticsReport
  organization_id    int
  properties         []*PropertyAnalyticsEntry
  total              int
  report_period      string
  generated_at       time.Time

PaymentAnalyticsResult   (intermediate, not exposed directly in API)
  MethodCounts    map[string]int
  StatusCounts    map[string]int64
  DailyTrend      map[string]int64
  TotalPayments   int64
```

#### `models/reports.go` — Hierarchical / analytics types

```
BuildingPaymentReport        payment records + stats for one building over a date range
PropertyPaymentReport        property-level stats + per-building breakdown
BuildingPaymentAnalytics     occupancy rate, revenue, trend analysis for one building
ComprehensiveReport          system/property/building-scoped report with groupings
DashboardReport              dashboard view with property/building/type groupings
```

---

### Services

#### `ReportService` — `services/report_service.go`

Depends on `PaymentRepositoryInterface` + `PropertyRepositoryInterface`.

| Method | What it does |
|---|---|
| `FinancialLedgerReport(orgID, filters, limit, offset)` | Paginates payments; summary totals from org-scoped `GetDashboardSummary` |
| `CollectionSummaryReport(orgID, startDate, endDate)` | Collection rate, real aging buckets via `GetAgingBuckets`, real monthly trend via `GetMonthlyCollectionTrend` |
| `PaymentAnalysisReport(orgID, startDate, endDate)` | All aggregations in DB via `GetPaymentAnalyticsByPeriod` — no in-memory row scanning |
| `TenantSummaryReport(orgID)` | Single JOIN query across tenants/leases/payments via `GetTenantPaymentSummary` |
| `PropertyAnalyticsReport(orgID, startDate, endDate)` | Lists org's properties then calls `GetPropertyPaymentStats` per property |

#### `BuildingAnalyticsService` — `services/building_analytics_service.go`

Live building performance metrics with an **in-memory cache** to avoid repeated aggregation queries.

| Method | Returns |
|---|---|
| `GetBuildingMetrics(buildingID)` | Occupancy %, revenue, performance score |
| `GetOccupancyAnalytics(buildingID)` | Vacancy tracking over time |
| `GetRevenueAnalytics(buildingID, period)` | Revenue breakdown by period |
| `CalculatePerformanceScore(buildingID)` | Composite 0–100 score |
| `GenerateTrendAnalysis(buildingID, months)` | Month-by-month trend |
| `CompareWithPropertyAverage(buildingID)` | Delta vs. property peers |

---

### API Endpoints

All endpoints require a valid JWT. `org_id` is extracted from the token via `RequireOrgContext()` middleware — never passed as a query parameter.

#### Reports (`/api/v1/reports/*`)

| Method | Path | Query params | Response |
|---|---|---|---|
| `GET` | `/reports/ledger` | `page`, `page_size`, `status`, `month`, `year` | `FinancialLedgerReport` |
| `GET` | `/reports/collection-summary` | `start_date`, `end_date` (YYYY-MM-DD) | `CollectionSummaryReport` |
| `GET` | `/reports/payment-analysis` | `start_date`, `end_date` | `PaymentAnalysisReport` |
| `GET` | `/reports/dashboard-metrics` | — | `{ collection_summary, payment_analysis, generated_at }` |
| `GET` | `/reports/tenant-summary` | — | `TenantSummaryReport` |
| `GET` | `/reports/property-analytics` | `start_date`, `end_date` | `PropertyAnalyticsReport` |

#### Payment reports (`/api/v1/payments/*`)

| Method | Path | Query params | Response |
|---|---|---|---|
| `GET` | `/payments/building/:building_id/report` | `start_date`, `end_date` | `BuildingPaymentReport` |
| `GET` | `/payments/property/:property_id/report` | `start_date`, `end_date` | `PropertyPaymentReport` |

#### Building analytics

| Method | Path | Query params | Response |
|---|---|---|---|
| `GET` | `/buildings/:id/analytics` | `period` (`month`/`quarter`/`year`) | `BuildingPaymentAnalytics` |

**Default date range** when `start_date`/`end_date` are omitted: last 30 days.

---

## Frontend

### ReportService

`src/frontend/src/app/core/services/report.service.ts`

```typescript
private reportService = inject(ReportService);

getFinancialLedger(page?, pageSize?, filters?): Observable<FinancialLedgerReport>
getCollectionSummary(startDate?, endDate?): Observable<CollectionSummaryReport>
getPaymentAnalysis(startDate?, endDate?): Observable<PaymentAnalysisReport>
getDashboardMetrics(): Observable<DashboardMetrics>
getTenantSummary(): Observable<TenantSummaryReport>
getPropertyAnalytics(startDate?, endDate?): Observable<PropertyAnalyticsReport>
```

All methods return `Observable`. Interfaces are defined in `report.service.ts`.

---

### ReportAnalysis Component

`src/frontend/src/app/features/reports/report-analysis.ts`

Standalone component. Route: `/reports`.

**State:**

| Property | Type | Purpose |
|---|---|---|
| `dashboardMetrics` | `DashboardMetrics \| null` | Loaded on init; feeds Overview tab |
| `collectionReport` | `CollectionSummaryReport \| null` | From dashboard or generated |
| `paymentReport` | `PaymentAnalysisReport \| null` | From dashboard or generated |
| `ledgerReport` | `FinancialLedgerReport \| null` | Set after ledger generation |
| `tenantReport` | `TenantSummaryReport \| null` | Set after tenant report generation |
| `propertyAnalyticsReport` | `PropertyAnalyticsReport \| null` | Set after property analytics generation |
| `buildingReport` | `unknown \| null` | Set after building performance generation |
| `buildings` | `Building[]` | Loaded on init for building filter dropdown |
| `properties` | `Property[]` | Loaded on init for property filter dropdown |
| `quickMetrics` | getter | Computed live from `dashboardMetrics` |

**Lifecycle:**
```
ngOnInit
  ├── initForm()
  ├── loadDashboardMetrics()   → getDashboardMetrics()
  ├── loadBuildings()          → BuildingService.getBuildings({ active: true })
  └── loadProperties()         → PropertyService.getProperties({ active: true })
```

**CSV export** — `exportReport('csv')` produces a browser download for: ledger, tenant, collection summary, property analytics, building performance. PDF/XLSX show "coming soon".

---

### Report Templates

| id | Name | Category | Backend call |
|---|---|---|---|
| `ledger` | Financial Ledger | `financial` | `getFinancialLedger()` |
| `collection_summary` | Collection Summary | `collections` | `getCollectionSummary()` |
| `property_analytics` | Property Analytics | `operational` | `getPropertyAnalytics()` |
| `tenant_report` | Tenant Report | `tenant` | `getTenantSummary()` |
| `building_performance` | Building Performance | `operational` | `PaymentService.getBuildingReport(buildingId, ...)` |
| `payment_analysis` | Payment Analysis | `financial` | `getPaymentAnalysis()` |

---

## Adding a New Report — End-to-End Checklist

### 1. Backend — Model

Add a struct to `internal/models/report.go`:

```go
type MyNewReport struct {
    OrganizationID int       `json:"organization_id"`
    // ... fields
    ReportPeriod   string    `json:"report_period"`
    GeneratedAt    time.Time `json:"generated_at"`
}
```

### 2. Backend — Repository

For aggregation queries, add to `payment_repository_raw.go` (or create `{entity}_repository_analytics.go`). Use DB-level `GROUP BY` — never fetch all rows and aggregate in Go.

Add the method signature to `interfaces/interfaces.go`. **Important:** also add a stub to `MockPaymentRepo` in `internal/services/payment_service_test.go` or the build breaks immediately.

### 3. Backend — Service

Add to `ReportService`:

```go
func (s *ReportService) MyNewReport(orgID int, startDate, endDate time.Time) (*models.MyNewReport, error) {
    // call repo, build struct
}
```

Add to `ReportServiceInterface` in `interfaces/interfaces.go`.

### 4. Backend — Handler

Add to `report_handler.go`:

```go
func (h *ReportHandler) GetMyNewReport(c *gin.Context) {
    orgID := c.GetInt("org_id")
    startDate, endDate, err := parseDateRange(c)  // defined in payment_handler.go
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    report, err := h.reportService.MyNewReport(orgID, startDate, endDate)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, report)
}
```

### 5. Backend — Route

In `server.go` under the `reports` group (which already has `RequireOrgContext()`):

```go
reports.GET("/my-new-report", middleware.RequireAnyRole(), reportHandler.GetMyNewReport)
```

### 6. Frontend — Service

Add interface + method to `report.service.ts`:

```typescript
export interface MyNewReport { organization_id: number; /* ... */ }

getMyNewReport(startDate?: string, endDate?: string): Observable<MyNewReport> {
  let params = new HttpParams();
  if (startDate) params = params.set('start_date', startDate);
  if (endDate) params = params.set('end_date', endDate);
  return this.http.get<MyNewReport>(`${this.apiUrl}/my-new-report`, { params });
}
```

### 7. Frontend — Report Template

Add to the `reports` array in `report-analysis.ts`:

```typescript
{
  id: 'my_new_report',
  name: 'My New Report',
  description: 'One-line description',
  icon: 'insert_chart',
  category: 'financial',   // financial | operational | tenant | collections
  requires: ['org_id', 'date_range'],
}
```

### 8. Frontend — Wire generation and display

In `generateReport()` add a `case 'my_new_report': this.loadMyNewReport(); break;`

Add a `loadMyNewReport()` method and a result section in `report-analysis.html` (follow the `ledger-results` pattern already in the template).

---

### Summary of file changes per new report

| Layer | File |
|---|---|
| Model | `internal/models/report.go` |
| Repository method | `internal/repositories/payment_repository_raw.go` |
| Repository interface | `internal/interfaces/interfaces.go` |
| Repository mock | `internal/services/payment_service_test.go` (MockPaymentRepo stub) |
| Service method | `internal/services/report_service.go` |
| Service interface | `internal/interfaces/interfaces.go` (ReportServiceInterface) |
| Handler | `internal/handlers/report_handler.go` |
| Route | `internal/server/server.go` |
| Frontend interface+method | `src/frontend/src/app/core/services/report.service.ts` |
| Frontend component | `src/frontend/src/app/features/reports/report-analysis.ts` |
| Frontend template | `src/frontend/src/app/features/reports/report-analysis.html` |
