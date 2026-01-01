# Tenantly Backend API

The backend for the Tenantly property management platform, built with Go (Golang).

## 🏗 Project Architecture

This project follows a modular, layer-based architecture designed for scalability and maintainability.

### Directory Structure

```text
src/backend/api/
├── cmd/                    # Entry points for the application
├── internal/               # Private application code
│   ├── models/             # Domain entities and data structures
│   ├── repositories/       # Database access layer (DAO)
│   ├── services/           # Business logic and use cases
│   ├── handlers/           # HTTP handlers (Controllers)
│   ├── middleware/         # HTTP middleware (Auth, Logging, etc.)
│   └── database/           # Database connection and migration tools
├── migrations/             # SQL migration files
└── go.mod                  # Dependencies
```

## 🚀 Recent Refactoring & Improvements

### Building Module Optimization

To address code bloat and improve maintainability, the Building module has been refactored to separate **Core Domain Logic** from **Analytics & Reporting**.

*   **Models Split**:
    *   `models/building.go`: Contains the core `Building` entity and basic CRUD structures.
    *   `models/building_analytics.go`: Contains complex analytics structures like `BuildingMetrics`, `RevenueAnalytics`, and `OccupancyAnalytics`.

*   **Repositories Split**:
    *   `repositories/building_repository.go`: Handles standard CRUD operations (Create, Read, Update, Delete) and search.
    *   `repositories/building_repository_analytics.go`: Handles heavy aggregation queries and statistical data retrieval.

This separation ensures that core operational logic remains lightweight while complex analytical processing is isolated.

## 🛠 Tech Stack

*   **Language**: Go 1.25+
*   **Web Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
*   **Database**: PostgreSQL (using `lib/pq` driver)
*   **Authentication**: JWT (JSON Web Tokens)
*   **Migrations**: `golang-migrate`
*   **Testing**: `testify`

## 🏃‍♂️ Getting Started

### Prerequisites

*   Go 1.25+
*   PostgreSQL running locally or via Docker

### Manual Setup (Local)

1.  **Configure Environment**:
    ```bash
    cp .env.example .env
    ```
2.  **Install Dependencies**:
    ```bash
    go mod download
    ```
3.  **Run Migrations**:
    ```bash
    # Ensure DB_URL is set in .env
    go run cmd/server/main.go migrate
    ```
4.  **Start Server**:
    ```bash
    go run cmd/server/main.go
    ```

### Running Tests

To run all tests in the internal directory:

```bash
cd src/backend/api
go test -v ./internal/...
```

To run a specific test suite (e.g., Building Repository):

```bash
cd src/backend/api
go test -v ./internal/repositories/...
```
