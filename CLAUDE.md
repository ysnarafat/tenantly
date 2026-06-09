# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working with code in this repository.

## 📋 Project Overview

**Tenantly** = property rental management system, Bangladesh market. Full-stack, three components:

- **Backend API**: Go (Gin) at `src/backend/api` - core business logic
- **Frontend**: Angular 21 at `src/frontend` - standalone components, Material Design
- **Notification Service**: .NET background service at `src/backend/notification-service` - SMS/Email alerts

PostgreSQL for storage. Docker Compose for local dev.

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

# Database migrations (auto-run on server startup; manual control via dedicated CLI)
go run cmd/migrate/main.go up      # Apply pending migrations
go run cmd/migrate/main.go down    # Rollback last migration
go run cmd/migrate/main.go version # Show current version

# Testing
go test -v ./internal/...                                        # All tests
go test -v ./internal/repositories/...                          # Specific package
go test -v -run TestPaymentHandler_Create ./internal/handlers/  # Single test by name
go test -v -count=1 ./internal/...                              # Force re-run (bypass cache)

# Building
go build -o ./tmp/main ./cmd/server
```

### Frontend (Angular)

**From `src/frontend/` directory:**

```bash
# Install dependencies
npm ci  # Use npm ci instead of npm install for CI environments

# Development
npm run start:local       # localhost:4200
npm run start:network     # 0.0.0.0:4200 (accessible on network)

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
- All cross-layer dependencies are interfaces defined centrally in `internal/interfaces/interfaces.go` (15+ interfaces: repositories, services, audit). Concrete implementations wired in `cmd/server/main.go`.
- Services receive repository interfaces via DI; handlers receive service interfaces via DI
- All database access goes through repositories
- Tests use table-driven patterns; integration tests use `testutil.SetupTestDB()` which spins up a test DB with migrations applied; mocks live in `internal/interfaces/mocks/`
- Building module split pattern: `building_repository.go` (CRUD) + `building_repository_analytics.go` (analytics) + `building_validation_service.go` (validation) — follow this for large modules


### Frontend (Angular) - Standalone Components with Feature-Level State

```
src/app/
├── core/            # Singletons: services, guards, interceptors
│   ├── services/    # API clients, auth, state management
│   ├── guards/      # Route guards
│   ├── interceptors/# HTTP interceptors
│   └── models/      # Shared TypeScript interfaces
├── features/        # Feature modules by domain
│   ├── [feature]/
│   │   ├── [name].ts              # Standalone component (no .component suffix)
│   │   ├── [name].html            # Template
│   │   ├── [name].scss            # Styles with theme support
│   │   ├── store/                 # Optional: Feature-specific NgRx store
│   │   │   ├── [name].actions.ts
│   │   │   ├── [name].reducer.ts
│   │   │   ├── [name].selectors.ts
│   │   │   └── [name].effects.ts
│   │   └── ...
│   └── ...
├── shared/          # Shared UI components, pipes
├── store/           # Global NgRx store, effects, actions
│   ├── auth/        # Authentication state
│   ├── app/         # Global app state
│   └── ...
└── app.routes.ts    # Route definitions
```

**Key Patterns:**
- Standalone components (no NgModule)
- Global state via NgRx in `src/app/store/`
- Features can have own NgRx stores (e.g., `features/properties/store/`) for feature-specific state
- Services use RxJS observables
- Material Design for UI
- Environment files for config

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
- **Migrations**: golang-migrate (versioned SQL in `migrations/`)
- **Migration Commands**:
  ```bash
  cd src/backend/api
  go run cmd/migrate/main.go up   # Apply pending migrations
  ```
  Migrations also run automatically on server startup.
- **Creating migrations**: Add files as `migrations/NNNNNN_description.up.sql` and `migrations/NNNNNN_description.down.sql`, where `NNNNNN` is the next sequential number (zero-padded to 6 digits, e.g. `000010`).

---

## 🔐 Authentication & Authorization

**Type**: JWT (JSON Web Tokens)

**Backend**:
- Tokens generated in `internal/handlers/auth.go`
- Middleware validates in `internal/middleware/`
- Claims include user ID and role
- Reference: `src/backend/api/AUTHENTICATION.md`

**Frontend**:
- Tokens stored in localStorage
- HTTP interceptor attaches token to requests
- Auth guard protects routes

**Current Roles & Multi-Tenancy** (5-level hierarchy):
- `SUPER_ADMIN` - Platform level, manages organizations
- `ORG_ADMIN` - Organization level, manages org users
- `Admin` - Organization data management
- `PropertyManager` - Property-level operations, building/unit management
- `Accountant` - Read-only financial access

Role constants live in `internal/middleware/auth.go`. Middleware helpers: `RequireSuperAdmin()`, `RequireOrgAdmin()`, `RequireAdmin()`, `RequireAnyRole()`.

---

## 🌐 Internationalization (i18n)

Frontend supports **English** and **Bangla** translations using **ngx-translate**.

### Translation File Structure

```
src/frontend/src/assets/i18n/
├── en.json       # English translations (235+ keys)
└── bn.json       # Bangla translations (235+ keys, synchronized with en.json)
```

### Key Organization Pattern

All keys use 3-level hierarchy: `NAMESPACE.SECTION.KEY`

```json
{
  "COMMON": {
    "BUTTONS": { "CANCEL": "Cancel", "SAVE": "Save", ... },
    "ERRORS": { "REQUIRED": "Required", ... },
    "ACTIONS": { "VIEW_DETAILS": "View Details", ... },
    "PAGINATION": { "RESULT": "result", "RESULTS": "results" },
    "EMPTY_STATE": { "TRY_ADJUST": "Try adjusting...", ... },
    "STATUS": { "ACTIVE": "Active", ... }
  },
  "LEASE_LIST": { "PAGE_TITLE": "Lease Management", ... },
  "CREATE_LEASE_DIALOG": { "TITLE": "Create New Lease", ... }
}
```

### Using Translations in Templates

```typescript
// In component TypeScript with TranslateModule imported:
import { TranslateModule } from '@ngx-translate/core';

@Component({
  imports: [TranslateModule],
  template: `<h1>{{ 'LEASE_LIST.PAGE_TITLE' | translate }}</h1>`
})
```

### Key Patterns

- **COMMON namespace**: Shared UI vocabulary (buttons, errors, pagination) used across multiple features — eliminates ~40% duplication
- **Feature namespaces**: Context-specific labels (e.g., `LEASE_LIST.TABLE.TENANT`, `TENANT_FORM_DIALOG.FIELDS.NAME`)
- **Validation errors**: Form-specific error messages stay in feature scope (e.g., `BUILDING_FORM_DIALOG.ERRORS.PATTERN`)

### Screens with Translations

✓ Complete (all keys translated):
- Dashboard
- Lease List, Create/Edit Lease Dialog
- Tenant List, Add Tenant Dialog
- Property List, Create/Edit Property Dialog
- Unit Form Dialog, Building Form Dialog
- Due List

❌ Incomplete (hardcoded text, needs translation):
- Organization Management (admin feature) - 2 screens
- User List (admin feature)
- Attachment List
- Payment List

### Adding New Translations

1. Add key to **both** `en.json` and `bn.json` (must be synchronized)
2. Use in template: `{{ 'NAMESPACE.SECTION.KEY' | translate }}`
3. Run `npm run build` to verify no missing key errors
4. Test both languages: toggle in UI language switcher

---

## 🎨 Frontend Styling & Theming

### Theme System

Frontend supports light/dark themes via **CSS variables**: runtime switching without reloads, consistent palettes, easy maintenance.

**Theme Structure:**
```scss
// In component SCSS files:
:root,
[data-theme='light'] {
  --bg-primary: #ffffff;
  --text-primary: #1a1a1a;
  --color-primary: #1e88e5;
  // ... more variables
}

[data-theme='dark'] {
  --bg-primary: #121212;
  --text-primary: #ffffff;
  --color-primary: #64b5f6;
  // ... more variables
}
```

**Using Theme Variables:**
```scss
.component {
  background-color: var(--bg-primary);
  color: var(--text-primary);
  transition: background-color 0.3s, color 0.3s; // Smooth theme transitions
}
```

**Theme Switching (TypeScript):**
```typescript
toggleTheme(): void {
  this.isDarkMode = !this.isDarkMode;
  localStorage.setItem('theme', this.isDarkMode ? 'dark' : 'light');
  document.documentElement.setAttribute('data-theme', this.isDarkMode ? 'dark' : 'light');
}
```

**Examples**: See `dashboard.scss` and `property-card.scss` for comprehensive theme implementation.

### Component Styling Guidelines

New Angular components:

1. **Organize SCSS**:
   - Group theme variables at top
   - Use mixins for reusable patterns
   - Order: variables → mixins → base styles → responsive media queries

2. **Use CSS Mixins for DRY Code**:
   ```scss
   @mixin card-style {
     background-color: var(--bg-primary);
     border: 1px solid var(--border-color);
     border-radius: 12px;
     box-shadow: var(--shadow-sm);
   }
   
   .card { @include card-style; }
   ```

3. **Responsive Breakpoints**:
   ```scss
   @media (max-width: 768px) { /* Tablet */ }
   @media (max-width: 480px) { /* Mobile */ }
   ```

4. **Animations**:
   - Prefer CSS-only animations for performance
   - Use `transition` for state changes (hover, focus)
   - Use `@keyframes` for complex animations
   - Example: See dashboard sparkline animations

5. **Accessibility**:
   - All interactive elements need `:hover`, `:focus`, `:active` states
   - Use `tabindex="0"` for custom interactive elements
   - Provide `aria-label` for icon-only buttons
   - Color contrast must meet WCAG AA (4.5:1 for text)

---

## 🧪 Testing Strategy

### Backend (Go)
- Table-driven tests in `*_test.go` files
- Mocks in `testutil/` package
- Run all: `go test -v ./internal/...`
- Run specific: `go test -v ./internal/repositories`

### Frontend (Angular)
- Jasmine/Karma framework
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

**Build Budgets** (Angular):
- Component style budget: 18kB max error (updated from 6.5kB for feature-rich components)
- Adjust in `angular.json` under `projects → tenantly-frontend → architect → build → configurations → production → budgets`

### Git Conventions (from CONTRIBUTING.md)
- **Branch naming**: `feat/`, `fix/`, `chore/`, `docs/`, `test/`, `topic/` prefixes
- **Commits**: Conventional Commits format
  - Example: `feat(auth): add JWT middleware`
  - Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `style`, `perf`

### Code Style
- **Go**: Follow [Effective Go](https://go.dev/doc/effective_go), use `gofmt`
- **Angular**: Follow [Angular Style Guide](https://angular.io/guide/styleguide), use Prettier + ESLint
  - Component naming: No `.component` suffix for standalone components
  - Use `[data-theme]` for theme-aware styling
  - Prefer `signal()` and `computed()` over traditional change detection when possible
- **C#**: Follow [Microsoft C# Conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions)

---

## 🔄 Development Workflow

1. Create branch: `git checkout -b feat/feature-name` or `git checkout -b topic/##/feature-name`
2. Make changes per conventions
3. Run quality checks:
   - Backend: `go test ./internal/...`
   - Frontend: `npm run quality && npm test`
4. Commit: `git commit -m "feat(scope): description"`
5. Push and create PR
6. CI/CD runs automatically (GitHub Actions in `.github/workflows/`)

---

## 🤖 Claude Code Behaviour

- **Commits**: Never add a `Co-Authored-By` trailer to any commit.

## 📝 Key Project Files

- `README.md` - Project overview and quick start
- `CONTRIBUTING.md` - Conventions & guidelines (code style reference)
- `src/backend/api/README.md` - Backend architecture details
- `src/backend/api/AUTHENTICATION.md` - Auth implementation details
- `src/frontend/README.md` - Frontend setup & conventions
- `ADMIN_HIERARCHY_PLAN.md` - Multi-tenancy & role hierarchy implementation status

---

## 🛠️ Scripts

Helper scripts in `scripts/`:
- `dev-setup.sh` / `dev-setup.bat` - Initial dev environment setup
- `fast-build.sh` / `fast-build.bat` - Quick build for all services
- `benchmark-build.sh` - Build performance profiling

---

## 🎯 Recent Enhancements (Reference Examples)

### Dashboard Redesign
- **Location**: `src/frontend/src/app/features/dashboard/`
- **Features**: Financial metrics, collection rate progress ring, 6-month trend sparklines, light/dark theme
- **Reference for**: Theme implementation, SVG charts, progress indicators, staggered animations

### Property Card Redesign
- **Location**: `src/frontend/src/app/features/properties/property-card/`
- **Features**: Visual hero section, quick stats bar, color-coded property types, responsive grid
- **Reference for**: Component styling, color-coding patterns, responsive design, nested hierarchies

---

## ⚠️ Common Issues & Notes

- **Frontend**: Angular 21 standalone components (no NgModule), signals for state
- **Backend**: Environment variables required (see `.env.example`)
- **Database**: PostgreSQL must run before API starts
- **Node version**: Node 25+ required
- **Go version**: Go 1.25+ required
- **SCSS Build Size**: Feature-rich components may exceed default style budgets; update `angular.json` if needed
- **Theme Persistence**: Stored in localStorage under `dashboard-theme` or `theme`
- **Bundle Budget Warnings**: Production build shows 3 warnings (bundle +293.58kB, fonts +3.72kB, styles +1.12kB) — acceptable for current scope, monitor if adding large features
- **Translation Key Sync**: Always keep en.json and bn.json synchronized (same number of keys, same structure). Build will pass without errors but missing Bangla keys fall back to English

---

## 📚 Frontend Documentation Files

- `docs/I18N_ORGANIZATION.md` - Comprehensive i18n patterns, Approach A (COMMON namespace), future migration path to Approach B (feature-split files)

---

## 🔗 Related Documentation

- **Angular Docs**: https://angular.io/docs
- **ngx-translate**: https://github.com/ngx-translate/core
- **Gin Web Framework**: https://github.com/gin-gonic/gin
- **PostgreSQL**: https://www.postgresql.org/docs/
- **Docker Compose**: https://docs.docker.com/compose/
- **Material Design**: https://material.angular.io/
- **NgRx**: https://ngrx.io/docs
