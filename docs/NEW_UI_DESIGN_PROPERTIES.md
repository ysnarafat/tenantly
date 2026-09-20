# Idea: Properties list + detail page redesign

Status: **idea / not started**. Captured from two mockups the user shared, with an
effort estimate. No implementation yet.

## Mockups being targeted

1. **Properties list** — header with Total/Active stat chips and search bar (mostly
   already built), property cards with color-coded icon avatars and a prominent
   "UNITS 124" stat, plus two new bottom widgets: "Portfolio Insights" (decorative,
   links to Reports) and "Maintenance Requests" (ticket count + "Manage Tickets" link).
2. **Property detail drill-down page** — a page that doesn't exist yet
   (`/properties/:id`): breadcrumb, header card with Edit/Delete, a stats row
   (Buildings / Total Units / Occupancy Rate / Revenue MoM), an expandable Buildings
   list with per-unit rent + OCCUPIED/VACANT status, and bottom widgets for Recent
   Payments, Lease Status, and a "Smart Insights" card.

## What already exists vs. what's new

Verified against the current codebase before estimating:

- Occupancy rate is **already computed** server-side (`property_repository.go:98-137`
  `GetByIDWithStats`, `building_repository_analytics.go:47-60`, and a
  `GetPropertyAggregations` endpoint in `property_handler.go:249-260`).
- Revenue is **only a lifetime cumulative total** today (`property_repository.go:107`)
  — no month-over-month figure exists anywhere.
- Per-unit rent is **not currently joined** into the unit list query
  (`unit_repository.go:353-415` `GetByPropertyWithDetails` returns `TenantName` /
  `LeaseActive` but no rent amount) — needs a query change to pull `leases.monthly_rent`.
- **No maintenance/ticket concept exists anywhere** — not in migrations, models,
  services, or the Angular side. This is a ground-up feature, not styling.
- **No property detail route/page exists** — `app.routes.ts` only has the list route;
  `features/properties/` has list + card + dialogs, no drill-down component.

## Effort breakdown

### A. Properties list page
| Item | Status | Effort |
|---|---|---|
| Header title, Total/Active stat chips, search bar | Already built | Match mockup's pill colors — **~15 min** |
| Color-coded square icon avatar per card | New, purely visual (rotate a palette by index or type) | **~30 min** |
| "UNITS 124" big-number stat replacing "X Building(s)" | Needs total-unit-count per property in the list response (not returned today) — backend query change + frontend bind | **~1.5–2 hrs** |
| "Show Buildings" / Edit / Delete footer | Already built | none |
| "Portfolio Insights" card (decorative gradient, links to Reports) | New component, static copy + link, no new data | **~1–1.5 hrs** |
| "Maintenance Requests" card, real data ("14 Pending", ticket list) | Full new feature: migration, model, repo, service, handler, Angular service + UI | **~2–3 days** |
| "Maintenance Requests" card, decorative placeholder only (static count, no backend) | Just a styled card, no wiring | **~45 min** |

**Subtotal A (cosmetic-only, maintenance as placeholder): ~4 hours**
**Subtotal A (maintenance fully real): ~2.5–3.5 days**

### B. New Property detail page
| Item | Status | Effort |
|---|---|---|
| New route `/properties/:id` + detail component shell, breadcrumb, header card, Edit/Delete | New page, straightforward Angular work | **~2 hrs** |
| Stats row: Buildings / Total Units / Occupancy Rate | Reuses existing `GetPropertyAggregations` endpoint | **~1 hr** |
| Stats row: Revenue (MoM) | New backend aggregation (this-month vs last-month paid amount) | **~2–3 hrs** |
| Buildings list, first expanded by default, unit mini-cards with Rent + OCCUPIED/VACANT | Needs per-unit rent join (repo change) + UI | **~2–3 hrs** |
| "View Analytics" link per building | Link to existing Reports module — reuse | **~30 min** |
| "Maintenance History" link per building | Depends entirely on the maintenance feature above — blocked until A's real version is built | included in maintenance estimate |
| "Recent Payments" widget | Existing payment endpoints already scoped by property/unit — reuse | **~1 hr** |
| "Lease Status" (75% secured / N expiring soon) | New but small: leases-expiring-within-window aggregation | **~1.5–2 hrs** |
| "Smart Insights" card ("12% above market average") | No market/benchmark data exists anywhere — would have to be static copy, or you'd need to define what it's actually computing | **~30 min** (static) / undefined (real) |
| Floating add button | Reuses existing add-property/building/unit dialogs | **~30 min** |

**Subtotal B (excluding Maintenance History, Smart Insights kept static): ~9–11 hours**

## Total

- **Cosmetic/reuse-first pass** (both pages, maintenance as placeholder, insights static): **~1.5–2 working days**
- **Full-fidelity version** (real maintenance/ticketing module + everything wired): **~4–5 working days**, because the maintenance feature alone is a full vertical slice (DB migration → repo → service → handler → Angular service/UI) and is the single biggest line item by far.

## Recommendation

Ship in two passes: (1) list + detail page with real occupancy/revenue/rent/lease-status
data and Maintenance as a static/placeholder card — gets you visually at parity fast;
(2) build the maintenance/ticketing module as its own follow-up feature once you
confirm you actually want that in scope (it's a real product surface, not a UI tweak).
