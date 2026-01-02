# Tenantly - Property Rental Management System

A comprehensive tenant management system designed for the Bangladesh market, focusing on property rental management.

## Overview

Tenantly is a robust, multi-service platform built to streamline property management. It handles everything from tenant onboarding and lease management to automated billing and notifications.

- **Backend API**: High-performance Go (Gin) service for core logic.
- **Frontend**: Modern Angular application with a responsive dashboard.
- **Notification Service**: Background worker built with .NET 8 for SMS/Email alerts.

## Quick Start

The fastest way to get started is using Docker Compose.

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Setup

1. **Clone the repository**
2. **Setup environment variables**:
   ```bash
   cp .env.example .env
   ```
3. **Start the platform**:
   ```bash
   docker-compose up -d
   ```

The applications will be available at:
- **Frontend**: `http://localhost:4200`
- **Backend API**: `http://localhost:8080/api/v1`
- **PostgreSQL**: `localhost:5432`

## Project Structure

For detailed documentation, please refer to the specific component directories:

| Component | Description | Documentation |
|-----------|-------------|---------------|
| **Backend API** | Go REST API | [backend/api](./src/backend/api/README.md) |
| **Notification Service** | .NET Background Service | [backend/notification-service](./src/backend/notification-service/README.md) |
| **Frontend** | Angular Application | [frontend](./src/frontend/README.md) |

## Features

- **Shop Management**: Track properties, units, and occupancy.
- **Tenant Management**: Profiles, contact info, and history.
- **Lease Processing**: Terms, deposits, and automated renewals.
- **Payments**: Multi-channel payment tracking and invoicing.
- **Automated Alerts**: SMS and Email reminders for rent and renewals.
- **Localization**: Native support for Bengali and BDT (৳).
- **CI/CD**: Automated linting and formatting via GitHub Actions.

## 🤝 Contributing

Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to get started with development.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.