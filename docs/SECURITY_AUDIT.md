# Security Audit Report — Tenantly

**Date:** 2026-06-12
**Auditor:** Claude Code (claude-sonnet-4-6)
**Branch audited:** `topic/report-and-analysis`
**Scope:** Full codebase — backend (Go/Gin), frontend (Angular), configuration, docker-compose

---

## Verdict

| | |
|---|---|
| **Publishable now?** | **No** |
| **Blockers** | 2 CRITICAL, 2 HIGH |
| **Estimated fix effort** | 1–2 days for CRITICAL + HIGH items |

The two CRITICALs are genuine data isolation breaks that allow tenants to read each other's data. Fix those before any deployment. The two HIGHs should follow in the same sprint. MEDIUMs are hardening items acceptable for an initial guarded release.

---

## Summary

| Severity | Count | Items |
|---|---|---|
| CRITICAL | 2 | IDOR on property/building fetch; internal error strings in responses |
| HIGH | 2 | JWT placeholder in docker-compose; no dedicated login brute-force protection |
| MEDIUM | 4 | CORS misconfiguration (dev); tokens in localStorage; password length inconsistency; in-memory rate limiter |
| LOW | 2 | Gin debug mode; inoperative password reset email |

---

## CRITICAL

### 1. IDOR — Property and Building accessible cross-organisation

**Files:**
- `src/backend/api/internal/handlers/property_handler.go:130`
- `src/backend/api/internal/handlers/building_handler.go:162`
- `src/backend/api/internal/repositories/property_repository.go:61`
- `src/backend/api/internal/repositories/building_repository.go:58`

**Description:**
`GET /api/v1/properties/:id` and `GET /api/v1/buildings/:id` fetch records by ID with no `organization_id` filter. Any authenticated user — even from a completely different tenant — can enumerate another organisation's properties and buildings by guessing sequential integer IDs.

**Vulnerable code:**
```go
// property_repository.go:61 — no org scope
SELECT id, property_name, ... FROM properties WHERE id = $1

// building_repository.go:58 — no org scope
SELECT ... FROM buildings WHERE id = $1
```

**Fix:**
```go
// Repository — add org filter
func (r *PropertyRepository) GetByID(id, orgID int) (*models.Property, error) {
    query := `SELECT ... FROM properties WHERE id = $1 AND organization_id = $2`
    // ...
}

// Handler — pass org_id from JWT context
func (h *PropertyHandler) GetProperty(c *gin.Context) {
    id, _  := strconv.Atoi(c.Param("id"))
    orgID  := c.GetInt("org_id")
    property, err := h.propertyService.GetProperty(id, orgID)
    // ...
}
```

This fix must be applied to both `GET /properties/:id` and `GET /buildings/:id`, including their `?include_stats=true` variants and the underlying service methods.

---

### 2. Internal error strings exposed in production responses

**Files:**
- `src/backend/api/internal/handlers/building_handler.go:30, 53, 129, 170, 185`
- `src/backend/api/internal/handlers/property_handler.go` (similar pattern)

**Description:**
`err.Error()` is returned directly to API clients in `"details"` fields. This leaks database table names, query structure, and internal paths (e.g., `"pq: relation \"buildings\" does not exist"`, `"sql: no rows in result set"`), enabling attacker reconnaissance.

**Vulnerable code:**
```go
c.JSON(http.StatusInternalServerError, gin.H{
    "error":   "Failed to retrieve building",
    "details": err.Error(),  // raw DB/runtime error sent to client
})
```

**Fix:**
```go
resp := gin.H{"error": "Failed to retrieve building"}
if os.Getenv("ENVIRONMENT") == "development" {
    resp["details"] = err.Error()
}
c.JSON(http.StatusInternalServerError, resp)
```

Or centralise in an error-handler middleware that strips `details` in production. This pattern appears in multiple handlers — a middleware approach avoids fixing each call site individually.

---

## HIGH

### 3. Hardcoded JWT placeholder in docker-compose files

**File:** `docker-compose.dev.yml:6`

**Description:**
The compose file ships a known, public JWT secret. If this file is used in any non-local environment (staging, CI, accidental production deploy), all JWT tokens can be forged because the secret is publicly known. The backend config rejects weak secrets in `ENVIRONMENT=production`, but `docker-compose.dev.yml` sets `ENVIRONMENT=development`, bypassing that guard.

**Vulnerable config:**
```yaml
JWT_SECRET: your-secret-key-change-in-production
```

**Fix:**
```yaml
environment:
  JWT_SECRET: ${JWT_SECRET:?JWT_SECRET must be set}  # process exits if not provided
```

---

### 4. Login endpoint has no dedicated brute-force protection

**Files:** `src/backend/api/internal/server/server.go:55-57`

**Description:**
The global rate limiter (100 req/min per IP) is shared across all endpoints. An attacker targeting `POST /auth/login` can submit ~98 password guesses per minute per IP without triggering a lockout. In a multi-container deployment the effective limit is `100 × N instances`. No account lockout mechanism exists.

**Fix:**
1. Add a stricter per-endpoint limiter on `/auth/login` (e.g., 10 req/min per IP):
```go
loginLimiter := middleware.NewRateLimiter(10, time.Minute)
auth.POST("/login", middleware.RateLimitMiddleware(loginLimiter, auditService), userHandler.Login)
```
2. Consider tracking failed attempts per username (not just per IP) to prevent distributed brute-force.

---

## MEDIUM

### 5. CORS: `Allow-Origin: *` + `Allow-Credentials: true`

**File:** `src/backend/api/internal/middleware/cors.go:19-20`

**Description:**
The development CORS middleware sets both `Access-Control-Allow-Origin: *` and `Access-Control-Allow-Credentials: true`. Browsers reject this combination (the spec forbids credentials with wildcard origin), so it has no practical effect now. However it signals an intent that could be broken by a future refactor. Only runs in development (`if environment != "development" { return }`), so production is unaffected today.

**Vulnerable code:**
```go
c.Header("Access-Control-Allow-Origin", "*")
c.Header("Access-Control-Allow-Credentials", "true")
```

**Fix:**
```go
c.Header("Access-Control-Allow-Origin", "http://localhost:4200")
c.Header("Access-Control-Allow-Credentials", "true")
```

---

### 6. JWT and refresh tokens stored in `localStorage`

**File:** `src/frontend/src/app/core/services/auth.service.ts:217-220`

**Description:**
Both access and refresh tokens are stored in `localStorage`, which is readable by any JavaScript on the page. Angular's template engine prevents XSS by default (no `innerHTML` or `bypassSecurityTrust` usage detected in the codebase), so the current attack surface is limited. However, any future XSS — via a third-party dependency, a template mistake, or a library vulnerability — immediately exposes all tokens.

**Vulnerable code:**
```typescript
localStorage.setItem(this.TOKEN_KEY, response.token);
localStorage.setItem(this.REFRESH_TOKEN_KEY, response.refresh_token);
```

**Fix (post-launch hardening):**
Move to `HttpOnly` cookies set by the server. The backend would set the cookie on login; the frontend would send it automatically on each request. Cookies with `HttpOnly` cannot be read by JavaScript regardless of XSS. This is a non-trivial architectural change; acceptable as a hardening item after initial release provided the XSS surface stays controlled.

---

### 7. Password minimum length inconsistency — model accepts 6, service requires 8

**File:** `src/backend/api/internal/models/user.go:23` vs `src/backend/api/internal/services/user_service.go:48`

**Description:**
The model-level binding validation accepts passwords of 6+ characters, but the service rejects anything below 8. The service check is the real gate so 6–7 character passwords are correctly rejected, but the model signals a weaker policy. API clients generating passwords from the model's constraints will produce non-functional inputs.

**Vulnerable code:**
```go
// model — accepts 6 chars
Password string `json:"password" binding:"required,min=6"`

// service — rejects below 8
const MinPasswordLength = 8
```

**Fix:**
```go
// models/user.go:23
Password string `json:"password" binding:"required,min=8"`
```

---

### 8. In-memory rate limiter — ineffective in multi-instance deployments

**File:** `src/backend/api/internal/middleware/security.go:14`

**Description:**
The `RateLimiter` stores counters in a Go `sync.Map` in process memory. In a multi-container deployment each instance maintains its own counter — an attacker can send `100 req/min × N instances` before being blocked. On any restart, all counters reset. The config already has a commented-out `REDIS_URL` variable anticipating this.

**Fix:**
Replace the in-memory map with a Redis-backed limiter (e.g., `go-redis` + a sliding window or token bucket pattern). Wire `REDIS_URL` from config:

```go
// Use Redis INCR + EXPIRE for a distributed sliding window
func (rl *RedisRateLimiter) Allow(key string) bool {
    count, _ := rdb.Incr(ctx, key).Result()
    if count == 1 {
        rdb.Expire(ctx, key, rl.window)
    }
    return count <= int64(rl.limit)
}
```

---

## LOW

### 9. Gin runs in debug mode in production

**Description:**
Gin defaults to debug mode, logging every registered route on startup. No `gin.SetMode(gin.ReleaseMode)` call was found. In debug mode, Gin also logs warning details that can reveal internal structure.

**Fix:**
```go
// cmd/server/main.go or server/server.go
if config.Environment == "production" {
    gin.SetMode(gin.ReleaseMode)
}
```

---

### 10. Password reset token generated but email never sent

**File:** `src/backend/api/internal/services/user_service.go:774`

```go
// TODO: Integrate with email service to send reset link
```

**Description:**
A `POST /auth/reset-password` request creates and stores a valid reset token (24-hour TTL) in `password_reset_tokens` but never dispatches it. Users receive a success-looking response but no email. Tokens accumulate in the table for a feature that is non-operational. Not exploitable by external attackers, but represents a data governance gap — tokens with no delivery mechanism are dead weight that could confuse auditors.

**Fix:**
Wire the existing `EmailProvider` config (SMTP credentials already read from env) to send the reset link. This is a functional completion item, not a new security feature.

---

## What Is Well Done

| Area | Assessment |
|---|---|
| **JWT implementation** | Algorithm pinned to HMAC, signature verified on every request, secret enforced ≥32 chars, startup fails in production without it |
| **Password hashing** | bcrypt with cost 12; `PasswordHash` tagged `json:"-"` — never appears in any API response |
| **SQL injection** | All dynamic query builders use parameterised `$N` placeholders; `fmt.Sprintf` only used for static column/table name construction, never user input |
| **Authentication coverage** | All routes except login, registration, and public reset require a valid JWT |
| **org_id sourced from JWT** | `RequireOrgContext()` extracts org from the token — callers cannot spoof their org via URL or query params |
| **Role-based access** | Middleware helpers consistently applied per route; `SUPER_ADMIN`, `ORG_ADMIN`, `Admin`, `PropertyManager`, `Accountant` correctly enforced |
| **Audit logging** | Auth events, CRUD actions, rate-limit violations all logged with IP and user-agent |
| **Security headers** | `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Content-Security-Policy`, HSTS all set |
| **XSS surface (frontend)** | No `innerHTML`, no `bypassSecurityTrust`, no `dangerouslySetInnerHTML` found — Angular's default escaping is intact |
| **Secrets management** | `.env` excluded from git; `JWT_SECRET` blocked at startup if absent in production; no hardcoded credentials in Go source |

---

## Recommended Fix Order

1. **This sprint (blockers):**
   - Fix IDOR on `GET /properties/:id` and `GET /buildings/:id` — add `AND organization_id = $N` to both repository `GetByID` methods
   - Strip `"details": err.Error()` from all production error responses
   - Remove JWT placeholder from docker-compose; require env injection

2. **Next sprint:**
   - Add dedicated login rate limiter (10 req/min per IP)
   - Fix password `min=6` → `min=8` in user model
   - Set `gin.SetMode(gin.ReleaseMode)` for production

3. **Post-launch hardening:**
   - Replace in-memory rate limiter with Redis-backed distributed limiter
   - Implement password reset email dispatch
   - Evaluate moving tokens from `localStorage` to `HttpOnly` cookies
   - Fix dev CORS to use specific origin instead of wildcard
