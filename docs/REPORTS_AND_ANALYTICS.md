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
        GetDashboardSummary()

ReportingService            (services/reporting_service.go)
  └── PaymentRepositoryInterface + BuildingRepositoryInterface

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
```

```
CollectionSummaryReport
  organization_id    int
  collection_rate    float64
  total_due          int64
  total_collected    int64
  total_pending      int64
  total_overdue      int64
  aging_buckets      map[string]int64        keys: "current", "30d", "60d", "90d+"
  monthly_trend      []*MonthlyCollectionTrend
  report_period      string                  "YYYY-MM-DD to YYYY-MM-DD"
  generated_at       time.Time

MonthlyCollectionTrend
  month              time.Time
  collection_rate    float64
  amount_due         int64
  amount_collected   int64
```

```
PaymentAnalysisReport
  organization_id       int
  payment_methods       map[string]int      e.g. {"Cash":12,"bKash":8}
  status_distribution   map[string]int64    e.g. {"Paid":20,"Due":5}
  daily_trend           map[string]int64    key: "YYYY-MM-DD"
  total_payments        int64
  report_period         string
  generated_at          time.Time
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

Thin orchestration layer. Depends only on `PaymentRepositoryInterface`.

| Method | What it does |
|---|---|
| `FinancialLedgerReport(orgID, filters, limit, offset)` | Paginates payments and attaches summary totals |
| `CollectionSummaryReport(orgID, startDate, endDate)` | Returns collection rate, aging buckets, 6-month trend |
| `PaymentAnalysisReport(orgID, startDate, endDate)` | Groups payments by method, status, and day |

> **Note:** `CollectionSummaryReport` currently returns the same `CollectionRate` for every month in the trend (uses the all-time summary). Per-month trending from the DB is a future improvement.

#### `ReportingService` — `services/reporting_service.go`

Comprehensive multi-level reports. Depends on payment and building repositories.

| Method | Scope |
|---|---|
| `GenerateComprehensiveReport(scope, id, start, end)` | `"building"`, `"property"`, or `"system"` |
| `GenerateDashboardReport(orgID, groupBy, buildingID?)` | Grouped by `"property"`, `"building"`, or `"type"` |

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

All endpoints require a valid JWT. `org_id` is extracted from the token — never passed as a query parameter.

#### Reports (`/api/v1/reports/*`)

| Method | Path | Query params | Response |
|---|---|---|---|
| `GET` | `/reports/ledger` | `page`, `page_size`, `status`, `month`, `year` | `FinancialLedgerReport` |
| `GET` | `/reports/collection-summary` | `start_date`, `end_date` (YYYY-MM-DD) | `CollectionSummaryReport` |
| `GET` | `/reports/payment-analysis` | `start_date`, `end_date` | `PaymentAnalysisReport` |
| `GET` | `/reports/dashboard-metrics` | — | `{ collection_summary, payment_analysis, generated_at }` |

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
// Inject anywhere needed
private reportService = inject(ReportService);

// Paginated transaction list
getFinancialLedger(page?, pageSize?, filters?): Observable<FinancialLedgerReport>

// Collection metrics for a date range
getCollectionSummary(startDate?, endDate?): Observable<CollectionSummaryReport>

// Payment method and status breakdown
getPaymentAnalysis(startDate?, endDate?): Observable<PaymentAnalysisReport>

// Combined current-month snapshot (used by dashboard tab)
getDashboardMetrics(): Observable<DashboardMetrics>
```

All methods return `Observable` — subscribe in the component or use `async` pipe in the template.

**Interfaces** (defined in `report.service.ts`, not a separate model file):

```typescript
FinancialLedgerReport    — mirrors backend model exactly
CollectionSummaryReport  — aging_buckets: Record<string, number>
PaymentAnalysisReport    — payment_methods / status_distribution / daily_trend as Record<string, number>
DashboardMetrics         — { collection_summary, payment_analysis, generated_at }
```

---

### ReportAnalysis Component

`src/frontend/src/app/features/reports/report-analysis.ts`

Standalone component. Route: `/reports` (verify in `app.routes.ts`).

**State:**

| Signal / property | Purpose |
|---|---|
| `selectedReport` | Currently active `ReportTemplate` |
| `dashboardMetrics` | Loaded on `ngOnInit` from `getDashboardMetrics()` |
| `collectionReport` | Extracted from `dashboardMetrics.collection_summary` |
| `paymentReport` | Extracted from `dashboardMetrics.payment_analysis` |
| `reportForm` | `reportType`, `startDate`, `endDate`, `propertyFilter`, `buildingFilter`, `exportFormat` |
| `activeTab` | Tab index (0 = Overview, 1 = Generate, ...) |

**Lifecycle:**

```
ngOnInit
  └── initForm()            build reactive form with defaults
  └── loadDashboardMetrics()
        └── reportService.getDashboardMetrics()
              → sets dashboardMetrics, collectionReport, paymentReport
```

---

### Report Templates

Defined as `ReportTemplate[]` inside the component. Each template maps to a `ReportService` call.

| id | Name | Category | Required params | Backend call |
|---|---|---|---|---|
| `ledger` | Financial Ledger | `financial` | `org_id` | `getFinancialLedger()` |
| `collection_summary` | Collection Summary | `collections` | `org_id`, date range | `getCollectionSummary()` |
| `property_analytics` | Property Analytics | `operational` | `org_id`, date range | *(not yet wired)* |
| `tenant_report` | Tenant Report | `tenant` | `org_id` | *(not yet wired)* |
| `building_performance` | Building Performance | `operational` | `org_id`, date range | `getBuildingReport()` via `PaymentService` |
| `payment_analysis` | Payment Analysis | `financial` | `org_id`, date range | `getPaymentAnalysis()` |

> `property_analytics` and `tenant_report` templates exist in the UI but do not yet call a backend endpoint — they are placeholders for future implementation.

---

## Adding a New Report — End-to-End Checklist

Follow this order to add a report cleanly through all layers.

### 1. Backend — Model

Add a new struct to `internal/models/report.go` (financial) or `internal/models/reports.go` (hierarchical/analytics):

```go
type MyNewReport struct {
    OrganizationID int       `json:"organization_id"`
    // ... fields
    ReportPeriod   string    `json:"report_period"`
    GeneratedAt    time.Time `json:"generated_at"`
}
```

### 2. Backend — Repository (if new DB query needed)

Add the query method to the appropriate repository file:
- Simple payment aggregation → `payment_repository_raw.go`
- Building-level analytics → `building_repository_analytics.go`
- New entity → create `{entity}_repository_analytics.go`

Add the method signature to `interfaces/interfaces.go` under the relevant repository interface.

### 3. Backend — Service

Add the method to `ReportService` (financial) or `ReportingService` (hierarchical):

```go
func (s *ReportService) MyNewReport(orgID int, startDate, endDate time.Time) (*models.MyNewReport, error) {
    // fetch from repository
    // build and return report struct
}
```

Add the signature to `ReportServiceInterface` in `interfaces/interfaces.go`.

### 4. Backend — Handler

Add a handler method to `report_handler.go`:

```go
func (h *ReportHandler) GetMyNewReport(c *gin.Context) {
    orgID := c.GetInt("org_id")
    startDate, endDate, err := parseDateRange(c)
    // call service, return JSON
}
```

### 5. Backend — Route

Register in `server.go` under the `reports` group:

```go
reports.GET("/my-new-report", middleware.RequireAnyRole(), reportHandler.GetMyNewReport)
```

### 6. Frontend — Service

Add the interface and method to `report.service.ts`:

```typescript
export interface MyNewReport {
  organization_id: number;
  // ...
}

getMyNewReport(startDate?: string, endDate?: string): Observable<MyNewReport> {
  let params = new HttpParams();
  if (startDate) params = params.set('start_date', startDate);
  if (endDate) params = params.set('end_date', endDate);
  return this.http.get<MyNewReport>(`${this.apiUrl}/my-new-report`, { params });
}
```

### 7. Frontend — Report Template

Add an entry to the `reports` array in `report-analysis.ts`:

```typescript
{
  id: 'my_new_report',
  name: 'My New Report',
  description: 'One-line description of what this shows',
  icon: 'insert_chart',              // Material icon name
  category: 'financial',            // financial | operational | tenant | collections
  requires: ['org_id', 'date_range'],
}
```

### 8. Frontend — Wire the generation

In the `generateReport()` method (or equivalent) in `report-analysis.ts`, add a `case` for the new template id that calls `reportService.getMyNewReport()` and binds the result to a component property for display.

---

### Summary of file changes per new report

| Layer | File |
|---|---|
| Model | `internal/models/report.go` or `reports.go` |
| Repository (if needed) | `internal/repositories/{entity}_repository_raw.go` or `_analytics.go` |
| Interface | `internal/interfaces/interfaces.go` |
| Service | `internal/services/report_service.go` |
| Handler | `internal/handlers/report_handler.go` |
| Route | `internal/server/server.go` |
| Frontend service | `src/frontend/src/app/core/services/report.service.ts` |
| Frontend component | `src/frontend/src/app/features/reports/report-analysis.ts` |
