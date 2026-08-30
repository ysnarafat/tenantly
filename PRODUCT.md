# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

**Primary — the small independent landlord in Bangladesh.** Owns one to three buildings or a handful of flats and manages them personally: collects the rent, chases the late payer, keeps the records. Their incumbent tool is a paper rent-book. They are not technical; the honest skill benchmark (stated in the product's own FAQ copy) is "if you can use WhatsApp or send a bKash payment, you already have the skills you need." Their day is spent moving between units, not sitting at a desk.

**Supported, but not shape-driving:**

- **PropertyManager / Accountant** — staff of a larger landlord or management org, each with their own scoped login. Design for them, but never at the primary user's expense.
- **ORG_ADMIN** — manages users and invitations inside one organization.
- **SUPER_ADMIN** — platform operator who creates and manages organizations. Internal audience.

## Product Purpose

Tenantly replaces the paper rent-book for Bangladeshi landlords with one system that holds the whole chain: **property → building → unit → tenant → lease → payment → report**. It exists so a landlord always knows who owes what, what was actually collected, and can hand a clean number to their accountant without a spreadsheet detour.

**Current stage: pre-pilot hardening. No real customers yet.** Success in the next stretch is narrow and specific: the core loop is airtight enough — no demo data, no dead buttons, no half-built screens, org isolation proven — to hand to a first real organization with real tenants and real money in it.

## Positioning

Built for how rent actually gets collected in Bangladesh, not a generic Western property-management SaaS with a language file bolted on:

- **One ledger, every local payment method.** Cash, bank transfer, bKash, and Nagad are all first-class ways money arrives, reconciled in one place.
- **Bilingual by design, per person.** The entire system switches between English and বাংলা at runtime, independently for each user on the team — not an admin-level setting applied to everyone.
- **The first reminder isn't a phone call.** Overdue rent triggers an automatic SMS and email nudge, removing the socially awkward step that keeps landlords on paper.

## Operating Context

- **Phone-first, in the field.** Rent is collected door-to-door or on the move. The app screens (dashboard, payments, leases, due list) are entered on a mid-range Android phone, sometimes on patchy data. Desktop is the secondary case, mostly for accountants and reporting sessions.
- **Migration is gradual, not a cutover.** The paper rent-book and Tenantly coexist for a while; the product explicitly promises users can bring records over at their own pace.
- **Multi-organization.** A user can belong to more than one organization; login routes to an organization picker when several are available, and issues organization-scoped JWTs. All org data access is scoped by `organization_id`.
- **Five roles, permission-gated.** `SUPER_ADMIN`, `ORG_ADMIN`, `Admin`, `PropertyManager`, `Accountant`. Frontend routes are gated by explicit permissions (`MANAGE_PROPERTIES`, `RECORD_PAYMENTS`, `VIEW_REPORTS`, …); blocked access lands on a dedicated 403 screen.
- **Onboarding is invitation-based.** New org members join via a token invitation (64-char token, 7-day expiry), not self-signup.
- **Reminders run out-of-band.** A .NET background worker delivers SMS and email on rent-due and lease-renewal events; it is not part of the request cycle.

## Capabilities and Constraints

**Confirmed and working:** authentication with role guards; dashboard metrics on real data; property/building/unit hierarchy; tenant CRUD with search and pagination; leases including per-unit lease defaults, due list, termination, and atomic tenant turnover (terminate + create — there are no placeholder leases per unit); payment recording including bulk and monthly generation; reports with CSV export; user management with invitations; organization list and create; document list, download, and delete; light/dark theming; English/বাংলা switching.

**Technical constraints:** Angular standalone components with signals and NgRx; Go (Gin) API over raw SQL on PostgreSQL — no ORM; .NET background worker for notifications; JWT auth carrying user, role, and organization claims.

**Terminology:** *property* contains *buildings*, which contain *units*; a *lease* binds a *tenant* to a *unit*; *dues* are unpaid amounts owed; *turnover* is terminating one lease and creating the next on the same unit.

**Explicitly undecided or unfinished — do not present these as shipped:**

- Attachment **upload** is not implemented (list, download, delete work). Whether the pilot ships attachments at all is an open decision.
- Password-reset **email delivery** is not wired. Open decision: wire the notification service, or disable the "forgot password" entry point.
- Report export exists for **CSV only**; PDF and XLSX are not implemented.
- Several admin routes are stubs (organization detail/edit, audit-log viewer, bulk invitations, user promotion) and are not pilot-ready.
- The **shops** module was cut: the API returns 501 and `/shops` redirects to `/properties`.
- `payment_method` is a free-form string on the API model; the method vocabulary lives in the UI layer.
- Organizations carry a `subscription_tier` field, but **no pricing, plans, billing, or commerce exists** anywhere in the product.

## Brand Commitments

- **Name and wordmark:** "Tenantly", set as `TENANTLY` beside an apartment glyph. No dedicated logo asset exists beyond `favicon.ico` and a Material icon.
- **Component foundation:** Angular Material stays as the base layer. Theming is CSS variables switched by a `data-theme` attribute on the document root, persisted in localStorage, with light and dark both supported.
- **Language:** English and বাংলা at full parity — every user-facing string exists in both, switchable at runtime per user. Bangla is never a second-class fallback. Translation files must stay key-for-key in sync.
- **Money:** BDT (৳), with cash, bank transfer, bKash, and Nagad as first-class payment methods — never collapsed into an "other" bucket.
- **Voice:** plain-spoken, concrete, and unpatronizing toward non-technical users. Compare against paper and everyday tools, not software categories. Reference line: *"Your rent-book, minus the torn pages."* Questions get answered plainly; no jargon, no enterprise register.
- **Contact:** `hello@tenantly.com`.

## Evidence on Hand

**Real, usable:**

- A working product across the full core loop, with the marketing homepage at `src/frontend/src/app/features/marketing/homepage/`.
- Bilingual copy in `src/frontend/src/assets/i18n/{en,bn}.json`, including hero, feature, and FAQ narratives already written in both languages.
- Development seed users, one per role, documented in `src/backend/api/README.md` (dev only — not proof of usage).
- Product and planning documents in `docs/` (MVP and pilot plan, reports architecture, i18n organization, security and data-access policy).

**Absent — must never be fabricated or implied:** customers, tenants of the platform, testimonials, partner or client logos, usage or collection metrics, case studies, press mentions, awards, pricing or plans, uptime figures, and security certifications or compliance badges. The product is pre-pilot; **no real customer exists yet**, and no design may suggest otherwise.

## Product Principles

1. **Paper is the competitor, not other software.** Every screen is judged against a notebook that never crashes, never asks for a password, and is always in the user's pocket. Beat it on recall and arithmetic, match it on speed.
2. **The phone is the real screen.** A flow that only makes sense at a desk is a flow the primary user will not complete. Field entry — one-handed, mid-range Android, imperfect network — is the design target, not the fallback.
3. **বাংলা is not a translation layer.** Both languages are the product. Layout, line lengths, numerals, and truncation must hold in both, and the switch must be findable by someone who cannot read the English label.
4. **Local money reality is the model, not an adapter.** Cash, bank transfer, bKash, and Nagad are how rent arrives; the ledger reflects that directly rather than normalizing it away.
5. **Never ship a button that does nothing.** Unfinished capability is hidden, not stubbed, and never demoed with fake data. Trust with someone's rent records is lost exactly once.

## Accessibility & Inclusion

**Conformance target: WCAG 2.1 AA.** This is binding on new and reworked surfaces — contrast, keyboard operability, visible focus, labels and names, and text resize hold in both English and বাংলা. Existing screens are not yet audited against it; treat gaps found in older code as debt to record, not as license to add more.

Product-specific needs on top of the standard:

- **Low digital literacy is the norm, not an edge case.** Users self-describe as "not good with computers"; affordances must be obvious without instruction, and destructive actions must be recoverable.
- **The language toggle must be operable by someone who cannot read English**, since it is the first control a Bangla-only user needs.
- **Bangla script must not break layout.** Bengali strings run longer and taller than their English counterparts; components must absorb that without clipping or reflow damage.
- **Performance is an accessibility concern here.** Mid-range Android hardware and patchy mobile data are the baseline device and network, not the degraded case.
