# Agent Guidance for Tenantly

## Architecture

**Three services, one DB:**
- **API**: Go (Gin) at `src/backend/api` - core business logic
- **Frontend**: Angular 21 at `src/frontend` - standalone components
- **Notification Service**: .NET at `src/backend/notification-service` - SMS/Email background jobs

All use PostgreSQL. Start with `docker-compose up -d`.

---

## Critical Patterns

### Interface Mock Maintenance

Adding a method to any repository interface in `src/backend/api/internal/interfaces/interfaces.go` breaks compilation of mock structs (e.g., `MockPaymentRepo` in `internal/services/payment_service_test.go`).

**Always add stub implementations:**
```go
func (m *MockPaymentRepo) NewMethod(args...) (ReturnType, error) {
    return zero, nil
}
```

### Multi-Service Boundaries

- **API**: Define interfaces → implement in repos/services → wire in `cmd/server/main.go`
- **Frontend**: Global state in `src/app/store/`, feature-level state in `features/[name]/store/`
- **Notification**: Background jobs, separate build/deploy cycle

**Never cross these boundaries directly.** Use defined interfaces and APIs.

### Formatting Before Commit

CI fails on unformatted files:

```bash
# From src/frontend
npm run format        # Prettier: single quotes, trailingComma: es5, printWidth: 100

# From src/backend/api
gofmt -s -w .         # Go formatting

# From src/backend/notification-service
dotnet format TenantlyNotificationService.sln
```

---

## Testing Prerequisites

**Backend tests split by type:**

| Test Type | Location | DB Required | Command |
|-----------|----------|-------------|---------|
| Service unit tests | `internal/services/*_test.go` | No | `go test ./internal/services/...` |
| Repository integration tests | `internal/repositories/*_test.go` | Yes | `go test ./internal/repositories/...` |

**Integration tests need PostgreSQL:** user=`user`, password=`p@ssw0rd`, database=`tenantly_test`.

---

## i18n Sync (Frontend)

Translation files: `src/frontend/src/assets/i18n/en.json` and `bn.json`

**Always keep both in sync.** Missing Bangla keys fall back to English silently. Use 3-level keys: `NAMESPACE.SECTION.KEY` (e.g., `LEASE_LIST.PAGE_TITLE`).

---

## Theme Switching Pattern (Frontend)

Runtime theme switching via CSS variables on `[data-theme]`:

```scss
:root, [data-theme='light'] {
  --bg-primary: #ffffff;
  --color-primary: #1e88e5;
}

[data-theme='dark'] {
  --bg-primary: #121212;
  --color-primary: #64b5f6;
}
```

Toggle by setting `data-theme` on `document.documentElement`; persist in localStorage.

---

## Org-Scoping Middleware

Every org-related route group must include `middleware.RequireOrgContext()`. Reads `organization_id` from JWT claims and injects as `"org_id"` (int) into Gin context.

**Example:**
```go
group := protected.Group("/reports")
group.Use(middleware.RequireOrgContext())

// Handler reads: c.GetInt("org_id")
```

---

## Database Migrations

Migrations live in `src/backend/api/migrations/` with `.up.sql` / `.down.sql` pairs.

**Commands from `src/backend/api`:**
```bash
go run cmd/migrate/main.go up        # Apply pending migrations
go run cmd/migrate/main.go down      # Rollback last migration
go run cmd/migrate/main.go seed      # Seed dev data (idempotent)
```

Migrations also run automatically on server startup.

---

## Build Commands

**API (Go):**
```bash
air                          # Dev watch mode
go run cmd/server/main.go    # Manual run
```

**Frontend (Angular):**
```bash
npm run start:local          # localhost:4200
npm run build                # Production build
npm run quality              # format + lint:fix
```

**Notification Service (.NET):**
```bash
dotnet build
dotnet run
```

---

## Definition of Done

**Frontend**: `npm run build` must succeed (checks TypeScript, SCSS bundles, and respects budgets).

**Backend API**: `go build ./...` must succeed with no compilation errors.

**Notification Service**: `dotnet build` must succeed with no warnings or errors.

Build success is required before considering any changes complete for their respective domain.

---

## Roles & Multi-Tenancy (5-level)

- `SUPER_ADMIN` - Platform level, manages organizations
- `ORG_ADMIN` - Organization level, manages org users
- `Admin` - Organization data management
- `PropertyManager` - Property-level operations
- `Accountant` - Read-only financial access

Middleware helpers: `RequireSuperAdmin()`, `RequireOrgAdmin()`, `RequireAdmin()`, `RequireAnyRole()`.
