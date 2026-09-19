# Forgot Password — Implementation Plan

**Status:** 📋 PLANNED (approved, not yet implemented)
**Date:** 2026-08-04
**Approach:** Email reset link, delivered by the Go API over SMTP
**Scope:** Authentication / account recovery

---

## Context

Users who log in (SUPER_ADMIN / ORG_ADMIN / Admin / PropertyManager / Accountant) have an **email but no phone** on the `users` table — only *tenants* have phone numbers, and tenants don't log in. So **email is the only viable reset channel**; SMS/OTP-by-SMS would first require adding and verifying phone numbers on users.

The backend reset flow is already ~90% built:

- `UserService.ResetPassword(email)` (`internal/services/user_service.go:730`) mints a CSPRNG 43-char token, stores it in `password_reset_tokens` (**1-hour, single-use**), and is **anti-enumeration** (always returns a neutral message).
- `UserService.ConfirmPasswordReset(token, newPassword)` (`user_service.go:798`) fully works (validates token, bcrypt, min-8 password).
- Routes exist and are public: `POST /api/v1/auth/reset-password`, `POST /api/v1/auth/confirm-reset-password` (`internal/server/server.go:131-132`).

**What's missing:**

1. **Delivery** — `user_service.go:790` is literally `// TODO: Integrate with email service to send reset link`. The token is generated but never sent.
2. **Frontend UI** — no pages/routes; the login "Forgot password?" link is a dead `href="#"` (`features/auth/login/login.html:38`); only the step-1 *request* is wired in NgRx, there is **no step-2 confirm** wiring.

The **.NET notification service** is deliberately **not** used here: both its email and SMS senders are stubs, and its `notification_queue` requires `tenant_id`/`unit_id NOT NULL` (a user reset email has neither).

---

## Backend (`src/backend/api`)

1. **New mailer** `internal/email/mailer.go`:
   - `Mailer` interface `Send(to, subject, htmlBody, textBody string) error`.
   - `SMTPMailer` using `net/smtp` + STARTTLS, built from `config.EmailConfig` (host/port/user/pass/from at `internal/config/config.go:64-70,155-161`).
   - `LogMailer` fallback that logs the reset link when SMTP creds are empty (i.e. dev), so the flow is testable without a real SMTP server. **Never** log the token in `production`.
   - Pure helper `BuildPasswordResetEmail(baseURL, token) (subject, html, text string)` — unit-testable without SMTP.
2. **Config** `internal/config/config.go`: add `AppBaseURL` from env `APP_BASE_URL` (dev default `http://localhost:4200`; else fall back to `AllowedOrigins[0]`). Used to build `${AppBaseURL}/reset-password?token=<token>`. Document it in `.env.example` alongside the existing `SMTP_*`.
3. **Wire delivery** in `UserService.ResetPassword` (`user_service.go:763-792`): after `CreateResetToken`, build the link and call the mailer at the current TODO. Keep anti-enumeration and non-blocking behavior — on send error, log + still `return nil` (never reveal existence). Guard for a nil mailer.
4. **DI** in `internal/server/server.go`: construct the mailer from `cfg` and inject into the user service (extend `NewUserServiceWithOrganization` at `server.go:95`, or add a `SetMailer`/`SetAppBaseURL`). Update the `UserService` struct (`user_service.go:17-24`) + constructor callers.
5. **Interface/mocks invariant** (CLAUDE.md): the mailer is a new dependency, not a repo-interface method, so `MockPaymentRepo` etc. are unaffected — but update `internal/services/user_service_test.go` UserService construction to pass a nil/stub mailer.
6. **Security hardening** (`server.go:126-132`): apply a dedicated strict limiter (e.g. 5/min) to the two reset endpoints (today only the global 100/min applies). Reuse the `middleware.NewRateLimiter` pattern already used for `loginRateLimiter`.

---

## Frontend (`src/frontend/src/app`)

Reuse the **login page as the visual/interaction template** (standalone, Material, signals for loading/error, `MatSnackBar`, theme, i18n) and the **existing step-1 NgRx wiring** (`AuthFacade.resetPassword` → `resetPassword$` effect → `POST /auth/reset-password`; `store/auth/auth.actions.ts:56-69`, `store/auth/auth.effects.ts:237-250`).

1. **`features/auth/forgot-password/forgot-password.{ts,html,scss}`** — email `FormControl`; submit calls `authFacade.resetPassword({email})`; on success show a **neutral** "If an account exists, we've emailed a reset link" state (no enumeration); back-to-login link.
2. **`features/auth/reset-password/reset-password.{ts,html,scss}`** — read `token` from `ActivatedRoute` query params; new + confirm password fields with a **strength meter**, match validation, and show/hide toggle (mirror login's password eye); submit calls the new service method; success → navigate `/login` + success snackbar; missing/invalid/expired token → error state with a "Request a new link" CTA (backend returns 400).
3. **`core/services/auth.service.ts`** — add `confirmPasswordReset(token, newPassword): Observable<{ message: string }>` (direct `POST /auth/confirm-reset-password`; the interceptor already whitelists it at `core/interceptors/auth.interceptor.ts:54`). Component manages its own loading/error via signals (login-style).
4. **Routing** `app.routes.ts` — add `forgot-password` and `reset-password` routes: `canActivate: [GuestGuard]`, lazy `loadComponent`, plain-string `title` (matches the SEO `TitleStrategy` convention).
5. **Shell** `app.ts` — add `/forgot-password` and `/reset-password` to `AUTH_ROUTE_PREFIXES` (currently `['/login','/select-organization']`) so the toolbar/sidenav stay hidden.
6. **Fix the link** `features/auth/login/login.html:38` — replace `<a href="#">Forgot password?</a>` with `routerLink="/forgot-password"` + a translated label; ensure `RouterModule` is in `login.ts` imports.
7. **i18n** `assets/i18n/en.json` + `bn.json` (kept in sync, CRLF preserved): new `FORGOT_PASSWORD` and `RESET_PASSWORD` namespaces + a `LOGIN.LINKS.FORGOT_PASSWORD` key. (ngx-translate lazy-loads on first pipe — these pages use the pipe, so translations load on them.)

---

## Out of scope (follow-ups)

- Real SMTP credentials (`SMTP_USERNAME` / `SMTP_PASSWORD`) must be set for prod delivery.
- Scheduling `CleanupExpiredTokens` (exists in `user_repository.go`, never called).
- Invalidating active sessions on password reset.
- Free-text/SMS channel; wiring the .NET notification service.

---

## Verification

- **Backend:** `go build ./...`; unit-test the pure `BuildPasswordResetEmail` (link + subject/body) and mailer selection (LogMailer when creds empty). On Windows, run test binaries from the gitignored `tmp/` to avoid the Application-Control block: `go test -c -o ./tmp/x.test.exe ./internal/email && ./tmp/x.test.exe -test.v`.
- **Frontend:** component specs for both pages (email validation, neutral-message state, token-from-query handling, password match/strength, invalid-token error). Run `ng test --watch=false --browsers=ChromeHeadless`; confirm the new specs pass by name. (The suite currently has ~286 pre-existing failures — do not conflate.)
- **End-to-end** (needs a running Postgres + the dev `LogMailer`): start the API + `npm run start:local`; on `/login` click "Forgot password?"; submit a seeded user (`admin@test.com`); copy the reset link from the API log; open it; set a new password (min 8) and confirm; expect redirect to `/login` with a success toast; log in with the new password. Confirm an unknown email still shows the neutral message and no email is logged.

---

## Key file references

**Backend (existing, reused):**
- `internal/services/user_service.go` — `ResetPassword` (730), `ConfirmPasswordReset` (798), `generateSecureToken` (872)
- `internal/handlers/user_handler.go` — `ResetPassword` (160), `ConfirmPasswordReset` (206)
- `internal/repositories/user_repository.go` — `CreateResetToken`, `GetResetToken`, `MarkResetTokenUsed`, `CleanupExpiredTokens`
- `internal/models/user.go` — `ResetPasswordRequest`, `ResetPasswordToken`, `ConfirmPasswordResetRequest`
- `internal/config/config.go` — `EmailConfig`
- `migrations/000001_initial_schema.up.sql` — `password_reset_tokens` table

**Backend (new):**
- `internal/email/mailer.go`

**Frontend (existing, reused):**
- `features/auth/login/` — visual/interaction template
- `store/auth/*` + `core/services/auth.service.ts` — step-1 reset wiring
- `core/guards/guest.guard.ts`, `app.routes.ts`, `app.ts`

**Frontend (new):**
- `features/auth/forgot-password/`, `features/auth/reset-password/`
