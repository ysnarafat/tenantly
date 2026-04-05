# Tenantly Frontend

Angular application for the Tenantly platform.

## Tech Stack

- **Framework**: Angular 21 (standalone components, no NgModule)
- **Language**: TypeScript
- **UI**: Angular Material (Material Design 3)
- **State**: NgRx (Store, Effects, Selectors)
- **Styling**: SCSS with CSS custom properties for light/dark theming
- **Build**: esbuild / Vite dev server

## Getting Started

```bash
cd src/frontend
npm ci
npm run start:local       # http://localhost:4200
npm run start:network     # http://0.0.0.0:4200 (LAN access)
```

**Prerequisites**: Node.js 25+, npm 10+

## Project Structure

```
src/app/
├── core/
│   ├── guards/        # AuthGuard (class), orgAdminGuard, superAdminGuard (functional)
│   ├── interceptors/  # HTTP auth interceptor — attaches token, handles 401 with refresh
│   ├── models/        # Shared TypeScript interfaces (re-exported from core/models/index.ts)
│   └── services/      # AuthService, OrganizationService, PermissionService
├── features/          # One directory per domain; components use [name].ts (no .component suffix)
│   ├── auth/          # login, organization-picker
│   ├── admin/         # organization-management, user-onboarding, audit-logs
│   └── [domain]/      # Optional store/ sub-directory for feature-level NgRx
├── shared/            # Reusable UI components (OrganizationSelector)
└── store/
    └── auth/          # Global auth state: actions, reducer, effects, selectors, facade
```

Feature-level NgRx stores (e.g. `features/properties/store/`) are provided lazily via `provideState`/`provideEffects` in route config, not registered globally.

## Auth & Multi-Organisation Flow

After a successful login the API returns `organizations[]` and an optional `default_organization_id`:

- **Multiple orgs, no default** → navigate to `/select-organization` (org picker)
- **Single org or default set** → store org context and navigate to `/dashboard`

Selecting an organisation calls `POST /auth/set-organization`, which returns a new org-scoped JWT. The token and org context are stored in `localStorage` and restored on page refresh by `initializeFromStorage()` (runs at module load) and the `initializeAuth` NgRx action.

### localStorage Keys

| Key | Content |
|-----|---------|
| `tenantly_token` | JWT access token |
| `tenantly_refresh_token` | Refresh token |
| `tenantly_user` | Serialised `User` object |
| `tenantly_expires_at` | ISO expiry string |
| `tenantly_organizations` | Full `UserOrganizationRole[]` array |
| `tenantly_current_org_id` | Active `organization_id` (FK to `organizations` table) |
| `tenantly_current_org` | Active `UserOrganizationRole` object |

Always use `org.organization_id` (not `org.id`) when identifying the current organisation.

## Route Guards

| Guard | Allows |
|-------|--------|
| `AuthGuard` | Any authenticated user |
| `orgAdminGuard` | `SUPER_ADMIN` or `ORG_ADMIN` |
| `superAdminGuard` | `SUPER_ADMIN` only |

`/admin/*` routes require `orgAdminGuard`. `/admin/organizations/*` additionally require `superAdminGuard`.

## Theming

Components use CSS custom properties scoped to `[data-theme]` on `:root`:

```scss
:root, [data-theme='light'] { --bg-primary: #ffffff; }
[data-theme='dark']         { --bg-primary: #121212; }

.card { background: var(--bg-primary); }
```

See `dashboard.scss` for a complete example. The component style budget is 18 kB (configured in `angular.json`).

## Quality & Testing

```bash
npm run quality            # Prettier + ESLint fix
npm run lint               # Lint only
npm run format:check       # CI formatting check

npm test                   # All unit tests (Karma/Jasmine, watch mode)

# Run a specific spec file without watch
npm test -- --include="**/store/auth/*.spec.ts" --watch=false

npm run build
```

### Test Patterns

- **Reducers**: call `reducer(state, action)` directly
- **Effects**: use `provideMockActions` + `HttpTestingController`; spy on `isDemoMode` for demo-mode branches
- **Selectors**: call `selector.projector(mockState)` directly — no store setup needed

## CI

GitHub Actions runs `npm ci`, `npm run lint`, and `npm run format:check` on every push and pull request.
