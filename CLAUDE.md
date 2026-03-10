# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 📋 Project Overview

**Tenantly** is a property rental management system built for the Bangladesh market. It's a full-stack platform with three main components:

- **Backend API**: Go (Gin) service at `src/backend/api` - handles core business logic
- **Frontend**: Angular 21 application at `src/frontend` - modern standalone components with Material Design
- **Notification Service**: .NET background service at `src/backend/notification-service` - handles SMS/Email alerts

The system uses PostgreSQL for data storage and Docker Compose for local development.

---

## 🚀 Development Commands

### Quick Start
```bash
# Start entire stack with Docker Compose
docker-compose up -d

# Access points:
# - Frontend: http://localhost:4200
# - Backend API: http://localhost:8080/api/v1
# - PostgreSQL: localhost:5432
```

### Backend API (Go)

**From `src/backend/api/` directory:**

```bash
# Setup
go mod download

# Development (watch mode with air)
air  # Auto-rebuilds on file changes (configured in .air.toml)

# Run manually
go run cmd/server/main.go

# Database migrations
go run cmd/server/main.go migrate

# Testing
go test -v ./internal/...  # All tests
go test -v ./internal/repositories/...  # Specific package

# Building
go build -o ./tmp/main ./cmd/server
```

### Frontend (Angular)

**From `src/frontend/` directory:**

```bash
# Install dependencies
npm ci  # Use npm ci instead of npm install for CI environments

# Development
npm start:local       # localhost:4200
npm start:network     # 0.0.0.0:4200 (accessible on network)

# Building
npm run build

# Quality & Formatting
npm run lint          # Check ESLint violations
npm run lint:fix      # Auto-fix ESLint issues
npm run format        # Format with Prettier
npm run format:check  # Check formatting without changes
npm run quality       # Run format + lint:fix

# Testing
npm test              # Unit tests
npm run e2e           # End-to-end tests
```

### Notification Service (.NET)

**From `src/backend/notification-service/` directory:**

```bash
# Build
dotnet build

# Run
dotnet run

# Testing
dotnet test

# Publish
dotnet publish -c Release
```

---

## 🏗️ Architecture & Structure

### Backend (Go) - Layered Architecture

```
internal/
├── models/          # Domain entities (User, Property, Unit, Lease, etc.)
├── repositories/    # Database access layer (CRUD, queries)
├── services/        # Business logic & use cases (orchestrates repositories)
├── handlers/        # HTTP handlers (request/response mapping)
├── middleware/      # Auth, logging, error handling
├── config/          # Environment & database configuration
├── database/        # DB connection & migration utilities
├── server/          # Gin server setup & routing
└── testutil/        # Test helpers & mocks

cmd/
├── server/          # Main API server entry point
└── migrate/         # Database migration CLI tool

migrations/         # SQL migration files (versioned)
```

**Key Patterns:**
- Services receive interfaces (repositories, dependencies) via dependency injection
- Handlers use services to handle HTTP requests
- All database access goes through repositories
- Tests use table-driven patterns and mocks from `testutil`

**Module Refactoring:** The Building module is split to separate core CRUD (`building_repository.go`) from analytics (`building_repository_analytics.go`) to prevent code bloat.

### Frontend (Angular) - Standalone Components

```
src/app/
├── core/            # Singletons: services, guards, interceptors
│   ├── services/    # API clients, auth, state management
│   ├── guards/      # Route guards
│   ├── interceptors/# HTTP interceptors
│   └── models/      # Shared TypeScript interfaces
├── features/        # Feature modules by domain
│   ├── [feature]/
│   │   ├── [name].ts    # Standalone component (no .component suffix)
│   │   ├── [name].html  # Template
│   │   └── [name].scss  # Styles
│   └── ...
├── shared/          # Shared UI components, pipes
├── store/           # NgRx store, effects, actions (global state)
└── app.routes.ts    # Route definitions
```

**Key Patterns:**
- Uses standalone components (no NgModule)
- NgRx for global state management
- Services use RxJS observables
- Material Design for UI components
- Environment files for configuration

### Notification Service (.NET)

```
src/backend/notification-service/
├── Services/        # Business logic (SMS, Email, Queue processing)
├── Models/          # Data models
├── Configuration/   # Settings & dependency injection
├── Data/            # Database context & models
└── Extensions/      # Extension methods & utilities
```

---

## 💾 Database

- **System**: PostgreSQL
- **Driver**: `lib/pq` (Go)
- **Migrations**: golang-migrate (versioned SQL files in `migrations/`)
- **Migration Commands**:
  ```bash
  cd src/backend/api
  go run cmd/server/main.go migrate  # Run pending migrations
  ```

---

## 🔐 Authentication & Authorization

**Type**: JWT (JSON Web Tokens)

**Backend**:
- Tokens generated in `internal/handlers/auth.go`
- Middleware validates tokens in `internal/middleware/`
- Claims include user ID and role
- Reference: `src/backend/api/AUTHENTICATION.md`

**Frontend**:
- Tokens stored in localStorage
- HTTP interceptor attaches token to requests
- Auth guard protects routes

**Current Roles** (3-level system):
- `Admin` - Full access
- `PropertyManager` - Property-level operations
- `Accountant` - Read-only financial access

---

## 🧪 Testing Strategy

### Backend (Go)
- Table-driven tests in `*_test.go` files
- Mocks in `testutil/` package
- Run all: `go test -v ./internal/...`
- Run specific: `go test -v ./internal/repositories`

### Frontend (Angular)
- Jasmine/Karma test framework
- Tests alongside components as `.spec.ts` files
- Run: `npm test`

### Integration
- Docker Compose for full stack testing
- API tests against running backend

---

## 📁 Important Files & Conventions

### Configuration
- `.env.example` - Environment template (copy to `.env` for local dev)
- `go.mod` (backend), `package.json` (frontend) - Dependencies
- `angular.json`, `tsconfig.json` - Frontend build config
- `docker-compose.yml`, `docker-compose.dev.yml` - Container orchestration

### Git Conventions (from CONTRIBUTING.md)
- **Branch naming**: `feat/`, `fix/`, `chore/`, `docs/`, `test/` prefixes
- **Commits**: Conventional Commits format
  - Example: `feat(auth): add JWT middleware`
  - Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `style`, `perf`

### Code Style
- **Go**: Follow [Effective Go](https://go.dev/doc/effective_go), use `gofmt`
- **Angular**: Follow [Angular Style Guide](https://angular.io/guide/styleguide), use Prettier + ESLint
- **C#**: Follow [Microsoft C# Conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions)

---

## 🔄 Development Workflow

1. Create feature branch: `git checkout -b feat/feature-name`
2. Make changes following conventions
3. Run quality checks:
   - Backend: `go test ./internal/...`
   - Frontend: `npm run quality && npm test`
4. Commit with conventional message: `git commit -m "feat(scope): description"`
5. Push and create pull request for review
6. CI/CD runs automatically (GitHub Actions workflows in `.github/workflows/`)

---

## 📝 Key Project Files

- `README.md` - Project overview and quick start
- `CONTRIBUTING.md` - Detailed conventions & guidelines (reference for code style)
- `src/backend/api/README.md` - Backend architecture details
- `src/backend/api/AUTHENTICATION.md` - Auth implementation details
- `src/frontend/README.md` - Frontend setup & conventions
- `ADMIN_HIERARCHY_PLAN.md` - Upcoming multi-tenancy & role hierarchy implementation plan

---

## 🛠️ Scripts

Available helper scripts in `scripts/` directory:
- `dev-setup.sh` / `dev-setup.bat` - Initial development environment setup
- `fast-build.sh` / `fast-build.bat` - Quick build for all services
- `benchmark-build.sh` - Performance profiling of builds

---

## ⚠️ Common Issues & Notes

- **Frontend**: Uses Angular 21 with standalone components (no NgModule)
- **Backend**: Environment variables required (see `.env.example`)
- **Database**: PostgreSQL must be running before API starts
- **Node version**: Node 25+ required for frontend
- **Go version**: Go 1.25+ required for backend

---

## 🔗 Related Documentation

- **Angular Docs**: https://angular.io/docs
- **Gin Web Framework**: https://github.com/gin-gonic/gin
- **PostgreSQL**: https://www.postgresql.org/docs/
- **Docker Compose**: https://docs.docker.com/compose/
