# Tenantly Backend API

Go REST API for the Tenantly platform, built with Gin.

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL (`lib/pq`, raw SQL)
- **Auth**: JWT access tokens + refresh tokens (bcrypt password hashing, cost 12)
- **Migrations**: `golang-migrate` — versioned SQL files in `migrations/`
- **Dev server**: `air` (hot reload via `.air.toml`)

## Getting Started

```bash
cp .env.example .env          # configure DB_URL and JWT_SECRET
go mod download
go run cmd/server/main.go migrate
go run cmd/server/main.go     # or: air
```

API is available at `http://localhost:8080/api/v1`.

## Project Structure

```
internal/
├── models/        # Domain structs + request/response types with JSON tags
├── repositories/  # DB access layer (raw SQL, lib/pq)
├── services/      # Business logic; depends only on repository interfaces
├── handlers/      # HTTP handlers; depends only on service interfaces
├── interfaces/    # All interface definitions in one file: interfaces.go
├── middleware/    # JWT auth, logging, CORS
├── config/        # Environment loading
├── database/      # Connection pool + migration runner
└── testutil/      # Shared mocks for table-driven tests

cmd/server/        # Entry point
migrations/        # 000NNN_description.up.sql + .down.sql pairs
```

### Dependency Injection

All cross-layer dependencies are defined as interfaces in `internal/interfaces/interfaces.go`. Handlers receive service interfaces; services receive repository interfaces. Never inject concrete types across layers — wire them in `cmd/server/main.go`.

### Building Module Split

The Building module separates concerns across two repository files to prevent bloat:
- `building_repository.go` — CRUD and search
- `building_repository_analytics.go` — heavy aggregation queries, metrics, forecasting

Apply the same pattern to any module that develops significant analytical queries.

## Multi-Tenancy

Migration `000007_admin_hierarchy_phase1` added `organization_id` to every data table (`users`, `properties`, `buildings`, `units`, `tenants`, `leases`, `payments`). All repository queries must be scoped by `organization_id`.

Migration `000008_user_organization_roles` introduced the `user_organization_roles` table — a user can belong to multiple organisations with different roles.

### Role System

| Role | Description |
|------|-------------|
| `SUPER_ADMIN` | Cross-org platform administration |
| `ORG_ADMIN` | Organisation-wide administration |
| `Admin` | Operational admin within an org |
| `PropertyManager` | Property/building/unit management |
| `Accountant` | Read-only financial access |

JWT claims include `user_id`, `username`, `role`, and `organization_id`. The `POST /auth/set-organization` endpoint issues a new token scoped to the requested organisation (validates membership via `user_organization_roles`).

### Password Policy

Enforced in `UserService.ValidatePassword`: 8–128 chars, at least one uppercase, lowercase, digit, and special character (`!@#$%^&*()_+-=[]{}';:"\\|,.<>/?`).

## Testing

```bash
# All tests
go test -v ./internal/...

# Specific package
go test -v ./internal/services/...

# Single test by name
go test -v ./internal/services/ -run TestLogin
```

Tests are table-driven. Mocks for all interfaces live in `internal/testutil/` — use those rather than creating per-test fakes.
