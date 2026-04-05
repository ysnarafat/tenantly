# Tenantly — Property Rental Management System

A multi-tenant property management platform built for the Bangladesh market. Handles tenant onboarding, lease management, payment tracking, and automated billing across multiple organisations.

## Services

| Service | Stack | Port |
|---------|-------|------|
| Backend API | Go (Gin) + PostgreSQL | 8080 |
| Frontend | Angular + Material Design | 4200 |
| Notification Service | .NET background worker | — |

## Quick Start

**Prerequisites:** Docker Desktop, Docker Compose

```bash
cp .env.example .env
docker-compose up -d
```

- Frontend: http://localhost:4200
- API: http://localhost:8080/api/v1

For local development without Docker, see the component READMEs below.

## Features

- **Multi-Organisation** — Users belong to one or more organisations; login routes to an organisation picker when multiple are available, issuing organisation-scoped JWTs
- **Role-Based Access** — Five roles: `SUPER_ADMIN`, `ORG_ADMIN`, `Admin`, `PropertyManager`, `Accountant`
- **Property Management** — Buildings, units, occupancy tracking
- **Tenant & Lease Management** — Profiles, lease terms, automated renewals
- **Payment Tracking** — Multi-channel payment recording and reporting
- **User Onboarding** — Token-based invitation workflow for new organisation members
- **Admin Panel** — Organisation management, invitation management, and audit logs (SUPER_ADMIN / ORG_ADMIN)
- **Notifications** — SMS and email reminders for rent due and lease renewals
- **Localisation** — Bengali language support, BDT (৳) currency

## Documentation

| Component | README |
|-----------|--------|
| Backend API | [src/backend/api/README.md](./src/backend/api/README.md) |
| Frontend | [src/frontend/README.md](./src/frontend/README.md) |
| Notification Service | [src/backend/notification-service/README.md](./src/backend/notification-service/README.md) |
| Authentication | [src/backend/api/AUTHENTICATION.md](./src/backend/api/AUTHENTICATION.md) |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for branch naming, commit conventions, and the development workflow.

## License

MIT — see [LICENSE](LICENSE).
