# AI Integration Roadmap — Reducing Manual Data Entry

## Context

Tenantly's forms already reduce manual entry a little (auto-generated property/building codes, cascading dropdowns, lease-based autofill in the payment-create dialog), but several high-friction, error-prone entry points remain fully manual — most notably: `monthly_rent` on lease creation is never pre-filled from history, tenant NID/name/address are hand-typed even though a photo of the ID card exists, and there's no quick way to create records from a short natural-language description. There is currently **zero AI/LLM/OCR integration anywhere in this codebase** (confirmed by exhaustive grep) — this is greenfield.

This roadmap covers three directions, using the **Anthropic Claude API** as the single provider for all of them (vision-capable, strong structured-JSON output, avoids needing a second specialized OCR service). Phase 1 is scoped concretely enough to implement directly from this doc; Phases 2–3 are a lighter-weight roadmap that reuse Phase 1's foundation.

**Ordering rationale:** Phase 1 (rent suggestions) is deliberately first because it's the lowest-risk direction — no PII, no file uploads — and it builds the reusable AI-calling plumbing (config, HTTP client, route group, error-handling conventions) that Phases 2–3 depend on. Phase 2 (ID-card OCR) is more sensitive — sending a government ID photo to a third-party API is a materially different risk than Phase 1/3's plain text, and this repo has a documented recent focus on protecting tenant NID at rest — so it gets explicit consent/no-persistence requirements. Phase 3 (natural-language quick entry) is last because it's the most open-ended (ambiguous input, needs a review-before-save UX) and benefits from the other two phases proving out the JSON-extraction pattern first.

---

## Phase 1: AI rent suggestion on lease creation

**Goal:** When a user selects a unit in `CreateLeaseDialog`, call a new backend endpoint that suggests a monthly rent based on recent comparable leases (same building, falling back to same property + unit type). Shown as a dismissible hint with an explicit "Use this" button — **never silently auto-filled**, and **fails completely silently** (no error toast) if the feature isn't configured or the call fails, since this is a non-critical enhancement that must never block lease creation.

### Backend

**Config** (`internal/config/config.go`) — mirror the existing `SMSConfig` pattern (confirmed at line 62/field at line 52):
```go
type AnthropicConfig struct {
    APIKey  string
    Model   string
    Enabled bool // true only if APIKey is non-empty
}
```
Add `AnthropicAI AnthropicConfig` to `Config`. In `Load()`, read `ANTHROPIC_API_KEY`/`ANTHROPIC_MODEL` (default `claude-sonnet-5`) via the existing `getEnv` helper. **Unlike `JWT_SECRET`/`DATABASE_URL`, this must NEVER hard-fail startup in production** — if the key is empty, log an informational warning and set `Enabled: false`. Add `ANTHROPIC_API_KEY=` / `ANTHROPIC_MODEL=claude-sonnet-5` to both root `.env.example` (new `# EXTERNAL SERVICES` section — none exists there yet) and `src/backend/api/.env.example` (existing `# EXTERNAL SERVICES (Optional - for future integrations)` section).

**New file `internal/services/anthropic_client.go`** — a minimal, reusable wrapper around Anthropic's Messages API (`POST https://api.anthropic.com/v1/messages`, headers `x-api-key`/`anthropic-version`/`content-type`), built generically so Phases 2–3 reuse it without changes:
- `NewAnthropicClient(apiKey, model string) *AnthropicClient`, `httpClient: &http.Client{Timeout: 20 * time.Second}`.
- `CreateMessage(ctx, system, userText string, maxTokens int) (string, error)` — low-level single-turn call, returns the first text content block.
- `CreateJSONMessage(ctx, system, userText string, maxTokens int, out interface{}) error` — instructs the model (via the caller's system prompt) to return ONLY JSON, defensively strips markdown fences/stray prose (`extractJSONObject`), unmarshals into `out`. **Phases 2/3 will call this same method** (Phase 2 passing an image content block instead of plain text — the `contentBlock`/`imgSource` types already support `type: "image"` with base64 `source`, so no interface change is needed later).
- `AnthropicAPIError{StatusCode int, ErrType, Message string}` with `IsRetryable() bool` (429/5xx) — a distinct, loggable error type for non-2xx responses instead of a generic wrapped error.
- No SDK dependency (this is the first outbound-HTTP-to-external-API call anywhere in this backend — confirmed by grep — so raw `net/http` keeps `go.mod` unchanged and stays trivially testable via `httptest.Server`).

**New repository method** (`internal/repositories/lease_repository.go`) — `GetComparableLeases(unitID, buildingID, propertyID int, unitType string, orgID int) ([]*models.LeaseWithDetails, error)`, reusing `LeaseWithDetails` (already has `UnitType`, `BuildingName`, embeds `Lease` with `MonthlyRent`/`StartDate` — no new model needed). Two-tier query (Go-level fallback, not a SQL UNION):
1. Same building, same unit_type, `start_date >= NOW() - INTERVAL '12 months'`, excluding the target unit itself, ordered `start_date DESC`, `LIMIT 10`.
2. If tier 1 returns fewer than 3 rows, fall back to same property + same unit_type (still building-agnostic), same recency/order/limit, and use this superset instead of merging.

Do **not** filter on `active = true` — a recently-ended lease's rent is still a valid market signal. Scope every query to `orgID`.

Add `GetComparableLeases(...)` to `LeaseRepositoryInterface` in `internal/interfaces/interfaces.go` — confirmed via `grep -rln "LeaseRepositoryInterface" --include=*_test.go .` that **no test fake currently implements this interface**, so this is a zero-risk addition (re-run that grep immediately before implementing, in case that's changed).

**New file `internal/services/ai_service.go`**:
```go
var ErrAINotConfigured = errors.New("AI features are not configured")

type AIService struct {
    client    anthropicJSONCaller // interface, not *AnthropicClient — see testing note below
    leaseRepo interfaces.LeaseRepositoryInterface
    unitRepo  interfaces.UnitRepositoryInterface
}

func (s *AIService) SuggestRent(unitID, orgID int) (*models.RentSuggestion, error)
```
- First line: `if s.client == nil { return nil, ErrAINotConfigured }` — no DB queries, no HTTP call.
- IDOR check: fetch unit via `unitRepo.GetByID`, compare `unit.OrganizationID != orgID` → return a generic "unit not found" (units denormalize `organization_id` directly — confirmed in migration 000002 — so no building lookup is needed here, simpler than `BulkCreateUnits`'s check).
- Fetch comparables via the new repo method; if empty, return `nil, nil` (not an error — handler treats this identically to "not configured": no hint shown).
- Build a **tabular, non-narrative** prompt (unit type/building/rent/date rows) — keeps token count low and JSON-out reliable — and call `client.CreateJSONMessage` requesting `{"suggested_rent": number, "confidence": "low"|"medium"|"high", "reasoning": "<one sentence>"}`.
- Wrap `*AnthropicAPIError` results with `%w` so `errors.As` still works upstream.

New `models.RentSuggestion{SuggestedRent float64, Confidence, Reasoning string, ComparableCount int}` (add to `internal/models/lease.go`).

New `AIServiceInterface{ SuggestRent(unitID, orgID int) (*models.RentSuggestion, error) }` in `interfaces.go` — no existing fake implements it (it's new), so nothing else to stub.

**New file `internal/handlers/ai_handler.go`** — `SuggestRent(c *gin.Context)`: parse `:id`, get `org_id` from context, call the service. On `errors.Is(err, ErrAINotConfigured)` or a nil suggestion → `200 {"available": false}` (not an error status — this keeps the frontend's error handling trivial). On success → `200 {"available": true, "suggested_rent", "confidence", "reasoning", "comparable_count"}`. Real failures (bad unit ID, IDOR, Anthropic error) → normal `respondError` 4xx/5xx (frontend will still swallow these silently per the UX requirement, but they're logged server-side like every other error).

**Route** (`internal/server/server.go`) — new group placed after `leases` (confirmed `POST /leases` uses `RequireAdminOrPropertyManager()` at line 282 — mirror it exactly since this only supports that flow):
```go
ai := protected.Group("/ai")
ai.Use(middleware.RequireOrgContext())
{
    ai.GET("/units/:id/suggest-rent", middleware.RequireAdminOrPropertyManager(), aiHandler.SuggestRent)
}
```
Wire `anthropicClient`/`aiService`/`aiHandler` in `setupRoutes()` alongside the existing repo/service/handler declarations, in the same conventional order.

### Frontend

**New `src/app/core/services/ai.service.ts`**:
```typescript
export interface RentSuggestion {
  available: boolean;
  suggested_rent?: number;
  confidence?: 'low' | 'medium' | 'high';
  reasoning?: string;
  comparable_count?: number;
}

suggestRent(unitId: number): Observable<RentSuggestion> {
  return this.http.get<RentSuggestion>(`${this.apiUrl}/ai/units/${unitId}/suggest-rent`)
    .pipe(catchError(() => of({ available: false })));
}
```
`catchError` here is the entire graceful-degradation contract — any HTTP failure resolves to the same shape as "not configured," so the component never has to special-case errors.

**Wiring in `create-lease-dialog.ts`** (confirmed: reactive form via `FormBuilder`, `unit_id` control exists — starts `disabled: true`, is enabled/disabled dynamically as building/property change; `monthly_rent` is `[null, [Validators.required, Validators.min(0)]]` with no default; no `ngOnDestroy` exists yet):
```typescript
this.leaseForm.get('unit_id')?.valueChanges.pipe(
  debounceTime(300),
  distinctUntilChanged(),
  switchMap((unitId) => {
    this.rentSuggestion = null; // clear stale suggestion immediately
    return unitId ? this.aiService.suggestRent(unitId) : of(null);
  })
).subscribe((s) => { this.rentSuggestion = s?.available ? s : null; });
```
Subscribing to `valueChanges` (rather than intercepting the select's own change handler) catches every path that sets `unit_id` uniformly, including the programmatic reset-to-null in `onBuildingChange`/`onPropertyChange` — `valueChanges` fires on programmatic changes to disabled controls too, so this works regardless of the control's enabled state.

UI: a dismissible `.ai-rent-hint` row directly under the `monthly_rent` field (icon + suggested amount + "Use this" + a close button), rendered only `*ngIf="rentSuggestion?.available"`. `applyRentSuggestion()` does `this.leaseForm.patchValue({ monthly_rent: ... })` and clears the hint — this is the **only** code path that ever touches `monthly_rent` from this feature. `dismissRentSuggestion()` just clears it. No `notifyError` anywhere in this subscription.

i18n: add `CREATE_LEASE_DIALOG.AI_SUGGESTION.{TEXT, USE_THIS, DISMISS}` to both `en.json` and `bn.json`, sibling to the existing `FIELDS`/`BUTTONS` keys, keeping the files structurally identical.

### Testing

- **`internal/services/ai_service_test.go`** (new): give `AIService` an `anthropicJSONCaller` interface field (`CreateJSONMessage(...)`) instead of holding `*AnthropicClient` directly, so tests can inject a fake without an HTTP server — matches this repo's existing "fake in-memory repos for service tests" convention. Cases: nil client → `ErrAINotConfigured` with zero repo calls; IDOR (`unit.OrganizationID != orgID`) → error, no `GetComparableLeases` call; empty comparables → `nil, nil`; fake caller returns `*AnthropicAPIError` → wrapped error survives `errors.As`; happy path → correctly populated `*models.RentSuggestion`.
- **`internal/repositories/lease_repository_test.go`** (add functions to the existing file, using `testutil.SetupTestDB`): same-building match; property+unit_type fallback when in-building count < 3; excludes leases older than 12 months; caps at 10; scopes to `organization_id`.
- **`internal/services/anthropic_client_test.go`** (new, `httptest.NewServer`): happy-path text extraction; JSON extraction through markdown fences; 429 → `*AnthropicAPIError` with `IsRetryable() == true`; 400 → `IsRetryable() == false`.
- **`src/app/core/services/ai.service.spec.ts`** (new, `HttpClientTestingModule` pattern like `property.service.spec.ts`): correct URL; passes through an `available:true` response; **an HTTP error still resolves to `{available:false}`** (the graceful-degradation contract, the most important test here).
- Manual verification: set `ANTHROPIC_API_KEY` in `.env`, create a lease for a unit with existing comparable leases in the dev DB, confirm the hint appears and "Use this" patches the field; then unset the key and confirm lease creation still works with no console errors and no hint shown.

---

## Phase 2 (roadmap): Tenant onboarding via ID-card photo

**Goal:** user uploads/photographs a tenant's NID card when creating a tenant; Claude's vision API extracts `{name, nid_number, address}` to pre-fill the (still fully editable) tenant form — never auto-submitted.

**Key facts:** there is currently **no attachment upload backend at all** (no `attachment_handler.go`/`attachment_service.go` exist despite the frontend having a stubbed "Upload attachment" button that just shows a snackbar) — this phase needs a brand-new multipart endpoint, not a wire-up of existing plumbing. It reuses Phase 1's `AnthropicClient.CreateJSONMessage` (passing an image content block instead of text) and the same config/route-group conventions.

**Security requirements (non-negotiable, given this repo's recent NID-hardening work — "protect tenant NID at rest and in transit"):**
- Explicit user consent UI before the upload/extract action fires ("this photo is sent to Anthropic's API for extraction and is not stored").
- The image is processed in-memory for the single request only — **never written to disk or the attachment store** as a side effect of extraction (a user can separately, deliberately attach the photo via the real attachment feature later if they want it retained).
- The extracted NID number must flow through whatever encryption-at-rest path `TenantService.CreateTenant` already uses for `nid_number` — never bypass it.
- Audit-log the AI call (who, when, which tenant record) without logging the image bytes or extracted NID in plaintext.
- New endpoint gets its own rate limit (reuse the existing `middleware.RateLimiter` pattern) given external vision calls are costly.

New pieces: `POST /api/v1/ai/extract-nid-card` (multipart, size/content-type validated), `AIService.ExtractNIDCard(imageBytes []byte, mimeType string) (*models.NIDExtraction, error)`, a form-dialog addition to `tenant-form-dialog` (create mode only) with a "Scan ID card" affordance above the existing fields.

---

## Phase 3 (roadmap): Natural-language quick entry

**Goal:** a single free-text box ("New tenant John Doe, 01712345678, unit 101, 15000/month from Aug 1") parsed into a structured draft that **pre-fills the existing** `TenantFormDialog`/`CreateLeaseDialog` for the user to review field-by-field — the AI never creates records directly; it only pre-fills forms that still require the user's own explicit Create click.

New pieces: `POST /api/v1/ai/parse-quick-entry` (`{text}` → structured draft + `missing_fields` + intent classification, using the same `CreateJSONMessage` foundation), and a `QuickAddDialog` entry point that renders the *existing* creation dialogs pre-filled rather than building new submission logic.

---

## Critical files (Phase 1)
- `src/backend/api/internal/config/config.go`
- `src/backend/api/internal/services/anthropic_client.go` (new)
- `src/backend/api/internal/services/ai_service.go` (new)
- `src/backend/api/internal/repositories/lease_repository.go`
- `src/backend/api/internal/interfaces/interfaces.go`
- `src/backend/api/internal/server/server.go`
- `src/backend/api/internal/models/lease.go`
- `.env.example` and `src/backend/api/.env.example`
- `src/frontend/src/app/core/services/ai.service.ts` (new)
- `src/frontend/src/app/features/leases/create-lease-dialog/create-lease-dialog.ts` + `.html`
- `src/frontend/src/assets/i18n/en.json` + `bn.json`
