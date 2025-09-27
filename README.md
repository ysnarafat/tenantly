# Tenantly - Shop Rental Management System

A comprehensive tenant management system designed for the Bangladesh market, focusing on shop rental management with future expansion to apartment rentals.

## 🏗️ Architecture

- **Backend API**: Go (Gin framework) - High-performance REST APIs
- **Frontend**: Angular 20+ with TypeScript - Modern responsive web application  
- **Notification Service**: C# (.NET 8) - Background processing for notifications and reminders
- **Database**: PostgreSQL - ACID compliant with connection pooling
- **Containerization**: Docker & Docker Compose for local development

## 📁 Project Structure

```
tenantly/
└──src
   ├── backend/                   # Backend services
   │   ├── api/                   # Go REST API (github.com/ysnarafat/tenantly)
   │   │   ├── cmd/server/        # Application entry point
   │   │   ├── internal/          # Private application code
   │   │   │   ├── config/        # Configuration management
   │   │   │   ├── database/      # Database connection & migrations
   │   │   │   ├── handlers/      # HTTP handlers (controllers)
   │   │   │   ├── middleware/    # HTTP middleware
   │   │   │   ├── models/        # Data models & DTOs
   │   │   │   ├── repositories/  # Data access layer
   │   │   │   ├── services/      # Business logic layer
   │   │   │   └── server/        # HTTP server setup
   │   │   ├── migrations/        # Database migration files
   │   │   └── Dockerfile         # Go API container
   │   └── notification-service/  # C# notification service
   │       ├── Services/          # Background workers
   │       ├── Models/            # Data models
   │       ├── Configuration/     # Service configuration
   │       └── Dockerfile         # Notification service container
   ├── frontend/                  # Angular 20+ application
   │   ├── src/app/              # Angular features and core services
   │   │   ├── core/             # Core services, guards, interceptors
   │   │   └── features/         # Feature areas (login, dashboard, etc.)
   │   │       └── [feature]/    # Each feature folder (e.g., auth, dashboard)
   │   │           ├── [name].ts # Angular 20+ component (no .component suffix)
   │   │           ├── [name].html # External template
   │   │           └── [name].scss # External styles
   │   ├── src/environments/     # Environment configurations
   │   └── Dockerfile            # Frontend container
   ├── scripts/                  # Development setup scripts
   └── docker-compose.yml       # Container orchestration
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ and npm (for frontend development)
- .NET 8 SDK (for notification service development)

### Using Docker Compose (Recommended)

1. Clone the repository
2. Copy environment configuration:
   ```bash
   cp .env.example .env
   ```
3. Update `.env` with your configuration values
4. Start all services:
   ```bash
   docker-compose up -d
   ```

### Manual Setup

#### Database Setup
```bash
# Start PostgreSQL
docker run -d --name tenantly-postgres \
  -e POSTGRES_DB=tenantly \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:15-alpine
```

#### Go API
```bash
cd backend/api

# Install dependencies
go mod download

# Run migrations
go run cmd/server/main.go migrate

# Start API server
go run cmd/server/main.go
```

#### Angular Frontend
```bash
cd frontend
npm install
npm start
```

##### Angular 20+ Component Conventions
- Components use external `.html` and `.scss` files for templates and styles.
- Component files are named `[name].ts` (no `.component` suffix), e.g., `login.ts`.
- Update all imports and routes to use the new file names.

#### C# Notification Service
```bash
cd backend/notification-service
dotnet restore
dotnet run
```

## 🔧 Development

### API Endpoints

The Go API provides RESTful endpoints for:
- Authentication (`/api/v1/auth`)
- User Management (`/api/v1/users`)
- Shop Management (`/api/v1/shops`)
- Tenant Management (`/api/v1/tenants`)
- Payment Processing (`/api/v1/payments`)
- Dashboard Analytics (`/api/v1/dashboard`)
- Reports (`/api/v1/reports`)


### Frontend Features

- **Angular 20+** with standalone components and new file naming conventions
- **External templates and styles** for all components
- **Material Design** components and responsive layout
- **Role-based access control** (Admin, Property Manager, Accountant)
- **Real-time dashboard** with financial summaries
- **Multi-language support** (Bengali and English)
- **Lazy loading** for optimal performance

### Background Services

- **Notification Worker**: Processes SMS and email notifications
- **Reminder Worker**: Automated rent payment reminders
- **Analytics Worker**: Generates reports and analytics (planned)

## 🗄️ Database Schema

Key entities:
- **Users**: Authentication and role management
- **Shops**: Property and shop information
- **Tenants**: Tenant profiles and contact information
- **Leases**: Tenant-shop relationships and lease terms
- **Payments**: Rent payment tracking and history
- **Notification Queue**: Asynchronous notification processing

## 🔐 Security Features

- JWT-based authentication with role-based access control
- Password hashing with bcrypt
- SQL injection prevention with parameterized queries
- CORS protection
- Audit logging for all financial transactions

## 🌐 Localization

- Bengali (বাংলা) and English language support
- Bangladesh Taka (৳) currency formatting
- Local date and time formatting
- SMS integration with Bangladesh telecom providers

## 📊 Monitoring & Logging

- Structured logging with Serilog (C# services)
- Request/response logging (Go API)
- Error tracking and monitoring
- Performance metrics collection

## 🚀 Deployment

### Production Deployment

1. Update environment variables for production
2. Build and deploy containers:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Environment Variables

Key configuration variables:
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: JWT signing secret (change in production)
- `SMS_API_KEY`: Bangladesh SMS provider API key
- `SMTP_*`: Email configuration for notifications

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue in the GitHub repository
- Check the documentation in the `/docs` folder
- Review the API documentation at `/api/v1/docs` (when running)

## 🗺️ Roadmap

- [ ] Complete user authentication system
- [ ] Implement payment processing
- [ ] Add SMS integration for Bangladesh providers
- [ ] Build comprehensive reporting system
- [ ] Add mobile application support
- [ ] Implement apartment rental management
- [ ] Add multi-property support
- [ ] Integrate with local payment gateways

## 🔄 Recent Updates

### Project Restructure
- **Backend Organization**: Separated Go API and C# Notification Service into distinct domains
- **Angular 20+ Upgrade**: Updated to latest Angular version with standalone components and new file naming conventions (no `.component` suffix)
- **Externalized Templates/Styles**: All Angular components now use external `.html` and `.scss` files
- **Improved Structure**: Clear separation of concerns with proper naming conventions
- **Docker Optimization**: Updated containers for new structure