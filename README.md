# Tenantly — Property Rental Management System

A multi-tenant property management platform built for the Bangladesh market. Handles tenant onboarding, lease management, payment tracking, and automated billing across multiple organisations.

**Website:** [tenantly.xyz](https://tenantly.xyz)

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

- **Multi-Organisation** — Users belong to multiple organisations with role-based access
- **Property Management** — Buildings, units, and occupancy tracking
- **Tenant & Lease Management** — Profiles, lease terms, and automated renewals
- **Payment Tracking** — Multi-channel payment recording and reporting
- **User Onboarding** — Token-based invitation workflow for new members
- **Admin Panel** — Organisation management, user invitations, and audit logs
- **Notifications** — SMS and email reminders for rent due and lease renewals
- **Localisation** — Bengali language support, BDT (৳) currency

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend API | Go (Gin) + PostgreSQL |
| Frontend | Angular 21 + Material Design 3 |
| Notifications | .NET background worker |

## Architecture

Tenantly is a three-service architecture:
- **API Server** (Go/Gin) — RESTful backend with JWT authentication
- **Web Frontend** (Angular) — SPA with NgRx state management
- **Notification Worker** (.NET) — Async SMS/email service

All services are containerised and coordinated via Docker Compose for easy local development and deployment.

## Documentation

| Component | Link |
|-----------|------|
| Backend API | [src/backend/api/README.md](./src/backend/api/README.md) |
| Frontend | [src/frontend/README.md](./src/frontend/README.md) |
| Notification Service | [src/backend/notification-service/README.md](./src/backend/notification-service/README.md) |
| Authentication | [src/backend/api/AUTHENTICATION.md](./src/backend/api/AUTHENTICATION.md) |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for branch naming, commit conventions, and the development workflow.

## License

MIT — see [LICENSE](LICENSE).
