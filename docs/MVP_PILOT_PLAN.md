# Tenantly — MVP & Pilot Release Plan

> Status snapshot taken 2026-06-30. Based on code scan of backend (`src/backend/api`) and frontend (`src/frontend`).

## Where things stand

Further along than typical pre-pilot. Backend ~90-95% on core domains. Frontend core flows production-ready. Gaps concentrated in **admin tooling** and **export/upload polish** — not core rent management.

---

## MVP definition (the pilot product)

Pilot = one or few real orgs managing actual properties, tenants, leases, payments. Core loop must be airtight:

**Property → Building → Unit → Tenant → Lease → Payment → Reports**

Everything else secondary.

### ✅ Ship as-is (ready)

| Area | State |
|------|-------|
| Auth (login, JWT, refresh, guards, RBAC) | Complete |
| Dashboard (metrics, real data) | Complete |
| Properties (CRUD, hierarchy, NgRx) | Complete |
| Tenants (CRUD, search, pagination) | Complete |
| Leases (CRUD, due list, termination) | Complete |
| Payments (CRUD, bulk, monthly generation, tree view) | Complete |
| Reports — view + CSV export | Complete |
| Users (CRUD, invite, roles) | Complete |
| Org management — list + create | Complete |

### ⚠️ Must fix before pilot (small, blocking)

1. **`lease.service.ts` demo data** — `DEMO_MODE` flag + hardcoded `'Building A'` (line ~124-144). Confirm flag off everywhere, remove demo block. Real tenant must not see fake data.
2. **Hide 5 stub admin routes** — render `<div>...(TODO)</div>`. Remove from router/nav so pilot users never hit dead screens:
   - `/admin/organizations/:id` (detail)
   - `/admin/organizations/:id/edit`
   - `/admin/audit-logs`
   - `/admin/invitations/bulk`
   - `/admin/users/promote`
3. **Login demo credentials** — `<!-- TODO: Remove demo credentials -->` in login template. Strip before real users.
4. **Password reset email** — backend flow done, email delivery is TODO (`user_service.go`). Either wire .NET notification-service to send, OR disable "forgot password" link for pilot and reset manually. Pick one — don't ship a button that silently does nothing.

### 🔵 Defer post-pilot (cut cleanly, no half-states)

- **Shops module** — backend returns `501`, frontend redirects `/shops → /properties`. Already cut. Leave it.
- **Reports PDF/XLSX** — CSV covers pilot. Better: hide PDF/XLSX buttons entirely so no dead clicks (currently "coming soon" toast).
- **Attachments upload / image viewer / edit** — list+download+delete work; upload is TODO. See Decision below.
- **Admin stubs** (org detail/edit, audit viewer, bulk import, promotion) — post-pilot.
- **Property manager assignment lookup** — TODO in `payment_service.go`. Only matters if pilot uses PropertyManager role with scoped access. Single-admin pilot → irrelevant.

---

## The one real decision: Attachments

Pilot landlords likely want lease PDFs / tenant docs. Upload is TODO.

- **A.** Hide attachments entirely for pilot — fastest, loses a selling point.
- **B.** Finish upload dialog only (skip image viewer + edit) — ~few days, gives core value.

Recommend **B** if pilot users are document-heavy, **A** if pilot is purely rent tracking.

---

## Suggested sequence

**Phase 0 — Pilot hardening (days, not weeks)**
1. Remove demo/mock data (lease service, login creds)
2. Hide stub routes + dead export/upload buttons
3. Resolve password-reset email (wire or disable)
4. Smoke-test full core loop end-to-end with real data
5. Confirm org isolation holds (one org cannot see another's data) — critical for multi-tenant pilot

**Phase 1 — Pilot launch**
- Onboard 1-3 real orgs
- Core loop only
- Collect feedback on payments/leases UX

**Phase 2 — Post-pilot (feedback-driven)**
- Attachments full
- Reports PDF/XLSX
- Admin tooling (audit viewer first — compliance want it)
- Shops (if market needs commercial units)

---

## OPEN QUESTIONS — answer below

### Q1. Pilot user profile
Single admin per org, or multi-role (PropertyManager / Accountant)?
*(Decides if property-manager-lookup TODO blocks.)*

**Answer:**
> _______________________________________________

### Q2. Attachments
Cut entirely (A), or finish upload only (B)?

**Answer:**
> _______________________________________________

### Q3. Password reset
Wire email now, or manual reset for pilot?

**Answer:**
> _______________________________________________

### Q4. Pilot size
How many orgs in pilot? *(Decides how hard org-isolation testing must be.)*

**Answer:**
> _______________________________________________

---

*Once answered: Phase 0 becomes concrete tickets.*
