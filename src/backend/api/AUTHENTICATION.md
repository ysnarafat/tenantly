# Authentication and Security Infrastructure

This document describes the authentication and security infrastructure implemented for the Tenantly API.

## Overview

The authentication system provides:
- JWT-based authentication with access and refresh tokens
- Role-based authorization (Admin, PropertyManager, Accountant)
- Password security with strength validation
- Comprehensive audit logging
- Rate limiting and security headers
- Session management with timeout

## Authentication Flow

### 1. Login
```
POST /api/v1/auth/login
{
  "username": "user@example.com",
  "password": "StrongPass123!"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "user@example.com",
    "role": "Admin",
    "active": true
  },
  "expires_at": "2024-01-01T16:00:00Z"
}
```

### 2. Token Refresh
```
POST /api/v1/auth/refresh
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 3. Logout
```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

## Role-Based Authorization

### Roles
- **Admin**: Full system access, user management
- **PropertyManager**: Property and tenant management, payment recording
- **Accountant**: Read-only access to financial data and reports

### Middleware Usage
```go
// Require specific role
router.Use(middleware.RequireAdmin())

// Require one of multiple roles
router.Use(middleware.RequireRole("Admin", "PropertyManager"))

// Convenience middlewares
router.Use(middleware.RequireAdminOrPropertyManager())
router.Use(middleware.RequireAnyRole())
```

## Password Security

### Requirements
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character

### Password Change
```
POST /api/v1/auth/change-password
Authorization: Bearer <access_token>
{
  "current_password": "OldPass123!",
  "new_password": "NewStrongPass456!"
}
```

### Password Reset
```
POST /api/v1/auth/reset-password
{
  "email": "user@example.com"
}
```

## Security Features

### Rate Limiting
- 100 requests per minute per IP address
- Failed attempts are logged for monitoring

### Security Headers
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security
- Content-Security-Policy

### Audit Logging
All authentication events are logged including:
- Login attempts (successful and failed)
- Token refresh
- Password changes
- Logout events
- API access with user context

### Session Management
- Access tokens expire after 8 hours (configurable)
- Refresh tokens expire after 7 days
- Automatic session timeout tracking
- User activity logging

## API Endpoints

### Public Endpoints
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh
- `POST /api/v1/auth/reset-password` - Password reset request

### Protected Endpoints
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/change-password` - Change password

### User Management (Admin only)
- `GET /api/v1/users` - List users (Admin/PropertyManager)
- `POST /api/v1/users` - Create user (Admin only)
- `GET /api/v1/users/:id` - Get user (Any role)
- `PUT /api/v1/users/:id` - Update user (Admin only)
- `DELETE /api/v1/users/:id` - Delete user (Admin only)

## Error Responses

All authentication errors include structured error responses:

```json
{
  "error": "Invalid credentials",
  "code": "LOGIN_FAILED"
}
```

Common error codes:
- `AUTH_HEADER_MISSING`
- `INVALID_TOKEN_FORMAT`
- `TOKEN_INVALID`
- `INSUFFICIENT_PERMISSIONS`
- `LOGIN_FAILED`
- `PASSWORD_CHANGE_FAILED`

## Configuration

Environment variables:
- `JWT_SECRET` - Secret key for JWT signing
- `JWT_EXPIRATION` - Token expiration duration (default: 8h)
- `DATABASE_URL` - Database connection string

## Testing

Run authentication tests:
```bash
go test ./internal/services -v
go test ./internal/middleware -v
```

## Security Considerations

1. **JWT Secret**: Use a strong, randomly generated secret in production
2. **HTTPS**: Always use HTTPS in production
3. **Token Storage**: Store tokens securely on the client side
4. **Rate Limiting**: Monitor and adjust rate limits based on usage patterns
5. **Audit Logs**: Regularly review audit logs for suspicious activity
6. **Password Policy**: Enforce strong password requirements
7. **Session Management**: Implement proper token invalidation on logout