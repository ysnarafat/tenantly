# OWASP Top 10 Security Audit — Tenantly

Audit date: 2026-07-17

Scope: Backend API (Go/Gin, `src/backend/api`), Frontend (Angular, `src/frontend`), Notification Service (.NET), and infra/CI config (Dockerfiles, docker-compose, GitHub Actions).

Methodology: parallel code-search sweeps per component against the OWASP Top 10 (2021) categories, followed by direct source verification of the highest-severity findings (CORS middleware, org-validation middleware, `shops` route wiring).

---

## Findings (verified, prioritized)

### CRITICAL

**1. CORS misconfiguration — always wildcard + credentials, env gate is broken**
`src/backend/api/internal/middleware/cors.go:13-22`
The `if environment != "development" { c.Next() }` branch has no `return`. Execution falls through unconditionally, so in every environment (including production) the response gets `Access-Control-Allow-Origin: *` **combined with** `Access-Control-Allow-Credentials: true`. Browsers reject `*` + credentials for actual credentialed requests, but the intended prod-only restriction never takes effect, and `c.Next()` is invoked twice in non-dev environments. **A05: Security Misconfiguration.**

**2. Secrets baked into Docker images**
`src/backend/api/Dockerfile` (~lines 16-19): `RUN echo "JWT_SECRET=your-secret-key-change-in-production" > .env` and a `DATABASE_URL` with an embedded password are written into a build layer, recoverable via `docker history`/`docker save` even after the file is later overwritten at runtime. **A02: Cryptographic Failures.**

**3. Deploy workflow triggers on `pull_request` with no restriction**
`.github/workflows/deploy-dev.yml`: triggers on `pull_request` (not scoped to a branch/path) and uses `secrets.DEV_SSH_PRIVATE_KEY`/`DEV_USER`/`DEV_HOST` to auto-deploy. No environment protection rule visible. **A05: Security Misconfiguration.**

### HIGH

**4. Tautological org-ownership check (dead code, but a landmine)**
`src/backend/api/internal/middleware/organization.go`: `OrganizationValidationMiddleware` sets context key `organization_id` to the **URL path's** org id. `RequireOrganization` later reads the same context key expecting the **user's own** org (from JWT) and compares it against itself — the check can never fail. Verified via grep that `RequireOrganization` is never actually wired into any route today, so it's currently inert, but it's exported, looks correct, and the next person who wires it up (e.g. for the `shops` module) gets an instant cross-tenant IDOR. **A01: Broken Access Control.**

**5. `shops` route group has no org-scoping or role middleware**
`src/backend/api/internal/server/server.go:238-245` — `protected.Group("/shops")` has no `middleware.RequireOrgContext()` and no role check, unlike every other data route group (`properties`, `buildings`, `units`, `tenants`, `leases`, `payments`, `dashboard`, `reports`). Currently only placeholder handlers, so no real data exposure yet, but it needs scoping wired in *before* real handlers land, and should not reuse the still-broken `RequireOrganization` from finding 4 without fixing it first. **A01: Broken Access Control.**

**6. Internal errors leaked to API clients**
`organization_handler.go` (7 occurrences), `middleware/organization.go:40`, `middleware/property_validation.go:40`, plus similar spots in `building_handler.go`, `property_handler.go`, `user_handler.go` — `500` responses include raw `err.Error()` as `"details"`, which can surface DB driver text (table/column names, SQL fragments). **A09 / Information Disclosure.**

**7. JWT + refresh token + full user object stored in plain `localStorage`**
`src/frontend/src/app/core/services/auth.service.ts:57-60,213-224` (read again in `auth.guard.ts`, `guest.guard.ts`, `auth.interceptor.ts:70`). Any XSS anywhere in the SPA (or a malicious browser extension) can read a long-lived refresh token → full account takeover, not just session hijack. **A02: Cryptographic Failures.**

**8. Containers run as root**
No `USER` directive in any of the three Dockerfiles (`api`, `frontend`, `notification-service`). **A05 / A06.**

### MEDIUM

**9. No per-account brute-force protection**
`server.go:121,124` — login is rate-limited only per-IP (10/min); no per-account lockout/backoff, so distributed credential stuffing across IPs is unmitigated. Also verify `SetTrustedProxies` is configured in `main.go` — if not, `c.ClientIP()` (used by the limiter) can be spoofed via forwarded headers. **A07: Authentication Failures.**

**10. Default DB credential fallback with no prod hard-fail**
`internal/config/config.go:94` — `DatabaseURL` defaults to `postgres://postgres:password@localhost:5432/...` if `DATABASE_URL` is unset. Unlike `JWT_SECRET` (which has a documented no-default-in-prod guard), a misconfigured prod deploy could silently run against weak default creds. **A05: Security Misconfiguration.**

**11. No CSP; font assets loaded without SRI**
`src/frontend/src/index.html` — no `Content-Security-Policy`, and `fonts.gstatic.com`/`fonts.googleapis.com` are loaded with no integrity hash. Defense-in-depth gap for XSS/CDN-compromise scenarios. **A05 / A08.**

**12. Token refresh has no request queuing**
`src/frontend/src/app/core/interceptors/auth.interceptor.ts:38-90` — expiry check trusts a client-computed `tenantly_expires_at` value from localStorage (tamperable, clock-skew prone), and concurrent 401s can each trigger a parallel refresh call. **A07: Authentication Failures.**

**13. CI has tests/lint/vet commented out**
`.github/workflows/backend-api-ci.yml` — `go vet`, `golangci-lint`, `go test` steps are commented out; no SAST/dependency scanning in any workflow. Doesn't directly map to one OWASP category but weakens detection of regressions in all of the above. **A09: Logging & Monitoring Failures** (process-level).

### LOW

**14. PII logged to browser console**
`src/frontend/src/app/features/tenants/tenant-list/tenant-list.ts:240` logs full tenant data (NID, phone) via `console.log`; a couple of `console.error(err)` calls elsewhere could echo response bodies containing tokens/PII. **A09.**

**15. Weak default secrets in docker-compose fallbacks**
`docker-compose.yml:11`, `docker-compose.dev.yml:5-6`, `docker-compose.prod.yml:8` fall back to `change-this-in-production` / `your-secret-key-change-in-production` if env vars aren't set. **A02.**

### Confirmed OK (no action needed)
- No SQL injection: raw-query repositories parameterize values (`$N`) and only `fmt.Sprintf` whitelisted column/sort names against a validated allow-list.
- Passwords hashed with bcrypt cost 12; reset tokens use `crypto/rand`.
- JWT: HMAC alg pinned, ≥32-byte secret enforced, no silent prod default.
- No `.env` files with real secrets are git-tracked (`.gitignore` correct); only `.env.example` files tracked.
- No `[innerHTML]`/`bypassSecurityTrust*`/`eval()` usage in the Angular frontend.
- Dependencies (Gin, jwt/v5, lib/pq, golang.org/x/crypto, Angular 21, RxJS 7.8) are current, not outdated majors.

---

## Remediation Plan

Work in the order below; each item is independently shippable.

1. **Fix CORS middleware** (`internal/middleware/cors.go`): add `return` after the `c.Next()` in the non-dev branch, and restrict `Access-Control-Allow-Origin` to a configured allow-list (env var) instead of `*` when credentials are allowed. Drop `Access-Control-Allow-Credentials` entirely if wildcard origin is kept for any environment.
2. **Stop baking secrets into Docker images**: remove the `echo ... > .env` layer from `src/backend/api/Dockerfile`; inject `JWT_SECRET`/`DATABASE_URL` at container runtime via env vars/secrets store only. Add a non-root `USER` directive to all three Dockerfiles (api, frontend, notification-service).
3. **Restrict `deploy-dev.yml` trigger**: scope to `workflow_dispatch` and/or `push` to a specific branch with environment-protection rules, not bare `pull_request`.
4. **Fix or remove `RequireOrganization`** (`internal/middleware/organization.go`): make `OrganizationValidationMiddleware` store the URL org id under a distinct context key (e.g. `"requested_organization_id"`) and have `RequireOrganization` compare it against the caller's own org id (from JWT claims via `RequireOrgContext`'s `"org_id"`), not the same key. Add a unit test proving cross-org access is rejected for a non-SUPER_ADMIN user.
5. **Add `middleware.RequireOrgContext()` (and appropriate role middleware) to the `shops` route group** in `server.go`, matching the pattern used by `properties`/`buildings`/etc., before any real handlers replace the placeholders.
6. **Stop leaking `err.Error()` to clients**: log the detailed error server-side (structured logger) and return a generic message + request ID to the client, across `organization_handler.go`, `middleware/organization.go`, `middleware/property_validation.go`, `building_handler.go`, `property_handler.go`, `user_handler.go`.
7. **Harden token storage on the frontend**: move the refresh token out of `localStorage` into an httpOnly cookie set by the backend (requires a backend endpoint change to set/read it), or at minimum shorten refresh-token lifetime and add rotation. Access token can remain in memory/localStorage short-lived. This is the largest item — treat as a separate scoping discussion before implementing.
8. **Add per-account login throttling**: track failed attempts per username/email (e.g. in Redis or a DB column) in addition to the existing per-IP limiter; verify `SetTrustedProxies` is set in `main.go` so `ClientIP()` can't be spoofed.
9. **Remove the default DB credential fallback** in `config.go`, or make it hard-fail outside `development` the same way `JWT_SECRET` already does.
10. **Add a CSP meta tag / header** to the frontend (`index.html` or via a response header from the serving layer) and add `integrity`/`crossorigin` attributes to the Google Fonts `<link>` tags, or self-host the fonts.
11. **Queue concurrent refresh calls** in `auth.interceptor.ts` (standard `switchMap` + `shareReplay(1)` pattern) instead of allowing parallel refresh requests.
12. **Re-enable `go vet`/`golangci-lint`/`go test` in `backend-api-ci.yml`**; consider adding `govulncheck` or similar dependency scanning.
13. **Remove PII from `console.log`/`console.error`** in `tenant-list.ts` and any other feature components found on a follow-up grep for `console.log` in `src/frontend/src/app/features/`.
14. **Replace hardcoded fallback secrets in docker-compose files** with a `required` env var (fail-fast) rather than a plaintext default, at least for `docker-compose.prod.yml`.

---

## Verification

- Backend: `go build ./...` then `go test ./internal/...` after each middleware/handler change. Add a new test for the org-scoping fix (#4) exercising cross-org access denial.
- CORS fix: manually curl a non-OPTIONS request with `Origin: https://evil.example` against a locally running server in a non-dev `ENVIRONMENT` value and confirm the response no longer echoes `*` with credentials allowed.
- Frontend: `npm run quality && npm test`; manually verify login/refresh flow still works after interceptor changes (`npm run start:local`), watching DevTools Application tab to confirm token storage location changed as intended.
- CI: push a throwaway commit to confirm `backend-api-ci.yml` now actually runs vet/lint/test and fails on an intentionally broken test.
- Docker: `docker build` each image and run `docker history <image>` to confirm no secret strings appear in any layer; confirm containers no longer run as root via `docker inspect --format '{{.Config.User}}'`.
