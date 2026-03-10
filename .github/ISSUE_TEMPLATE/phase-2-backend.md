---
name: "[Phase 2] Backend API Implementation"
about: "Phase 2: Create services, handlers, middleware, and API endpoints"
title: "[Phase 2] feat(admin): Backend API Implementation"
labels: ["enhancement", "high-priority", "backend"]
---

## 📋 Overview

**[CHILD ISSUE - Phase 2 of 4]**
**Parent Epic**: `[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy`

Implement backend services, handlers, authorization middleware, and API endpoints for admin hierarchy and multi-organization support.

**Effort**: High | **Duration**: 5-7 days
**Depends on**: Phase 1 ✅
**Blocks**: Phase 3 (Frontend Implementation)

---

## 📦 New Services

### OrganizationService
- [ ] Create `internal/services/organization_service.go`
- [ ] Methods:
  - [ ] `GetOrganizationByID(ctx context.Context, id int) (*Organization, error)`
  - [ ] `ListOrganizations(ctx context.Context) ([]*Organization, error)`
  - [ ] `CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*Organization, error)`
  - [ ] `UpdateOrganization(ctx context.Context, id int, req *UpdateOrganizationRequest) (*Organization, error)`
  - [ ] `DeleteOrganization(ctx context.Context, id int) error`
  - [ ] `GetOrganizationUsers(ctx context.Context, orgID int) ([]*User, error)`
  - [ ] `GetOrganizationStats(ctx context.Context, orgID int) (*OrgStats, error)`

### UserInvitationService
- [ ] Create `internal/services/user_invitation_service.go`
- [ ] Methods:
  - [ ] `SendInvitation(ctx context.Context, req *InviteUserRequest) (*UserInvitation, error)`
  - [ ] `GetInvitation(ctx context.Context, token string) (*UserInvitation, error)`
  - [ ] `AcceptInvitation(ctx context.Context, token, password string) (*User, error)`
  - [ ] `RevokeInvitation(ctx context.Context, invitationID int) error`
  - [ ] `ListPendingInvitations(ctx context.Context, orgID int) ([]*UserInvitation, error)`
  - [ ] `ResendInvitation(ctx context.Context, invitationID int) error`

### AuditLogService
- [ ] Create `internal/services/audit_log_service.go`
- [ ] Methods:
  - [ ] `LogAction(ctx context.Context, log *AuditLog) error`
  - [ ] `GetAuditLogs(ctx context.Context, orgID int, filters *AuditFilter) ([]*AuditLog, error)`
  - [ ] Helper methods for common actions: LogUserCreated, LogUserDeleted, LogUserPromoted, LogUserDemoted

### UserService - UPDATE
- [ ] Update `internal/services/user_service.go`
- [ ] Add org context to existing methods:
  - [ ] `GetUserByID(ctx context.Context, userID, orgID int) (*User, error)`
  - [ ] `ListUsers(ctx context.Context, orgID int) ([]*User, error)`
  - [ ] `CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)`
  - [ ] `UpdateUser(ctx context.Context, userID int, req *UpdateUserRequest) (*User, error)`
  - [ ] `DeleteUser(ctx context.Context, userID, orgID int) error`
- [ ] New methods for admin operations:
  - [ ] `PromoteToOrgAdmin(ctx context.Context, userID, orgID int) error`
  - [ ] `DemoteFromOrgAdmin(ctx context.Context, userID, orgID int) error`
  - [ ] `DeactivateUser(ctx context.Context, userID, orgID int) error`

---

## 🔐 Authorization Middleware

### New Middleware Files

#### `internal/middleware/organization.go` - CREATE
- [ ] `RequireOrganization()` middleware:
  - Verify user belongs to requested organization
  - Extract org_id from JWT and validate
  - Abort with 403 if unauthorized
- [ ] `RequireRole(roles ...string)` middleware:
  - Check if user has required role(s)
  - Support role hierarchy (admin > property_manager > accountant)
  - Abort with 403 if insufficient permissions

#### `internal/middleware/auth.go` - UPDATE
- [ ] Update JWT token generation to include:
  - [ ] `organization_id`
  - [ ] `role`
  - [ ] `first_name` and `last_name`
- [ ] Update JWT validation to extract new claims

### Helper Functions
- [ ] `ExtractOrgIDFromContext(c *gin.Context) (int, error)`
- [ ] `ExtractUserFromContext(c *gin.Context) (*User, error)`
- [ ] `IsUserInOrganization(userID, orgID int) (bool, error)`
- [ ] `HasRole(user *User, requiredRoles ...string) bool`

---

## 🌐 API Handlers & Endpoints

### OrganizationHandler - CREATE
Location: `internal/handlers/organization_handler.go`

**Endpoints** (SUPER_ADMIN only):
- [ ] `POST /api/v1/organizations` - Create organization
- [ ] `GET /api/v1/organizations` - List all organizations
- [ ] `GET /api/v1/organizations/:id` - Get organization details
- [ ] `PUT /api/v1/organizations/:id` - Update organization
- [ ] `DELETE /api/v1/organizations/:id` - Delete organization

Handlers:
- [ ] `CreateOrganization(c *gin.Context)`
- [ ] `ListOrganizations(c *gin.Context)`
- [ ] `GetOrganization(c *gin.Context)`
- [ ] `UpdateOrganization(c *gin.Context)`
- [ ] `DeleteOrganization(c *gin.Context)`

### UserHandler - UPDATE
Location: `internal/handlers/user_handler.go`

**Updated Endpoints** (with org context):
- [ ] `GET /api/v1/organizations/:orgId/users` - List org users
- [ ] `POST /api/v1/organizations/:orgId/users` - Create user
- [ ] `GET /api/v1/organizations/:orgId/users/:userId` - Get user details
- [ ] `PUT /api/v1/organizations/:orgId/users/:userId` - Update user
- [ ] `DELETE /api/v1/organizations/:orgId/users/:userId` - Remove user

**New Endpoints** (Admin operations):
- [ ] `POST /api/v1/organizations/:orgId/users/:userId/promote` - Make ORG_ADMIN
- [ ] `POST /api/v1/organizations/:orgId/users/:userId/demote` - Remove admin status
- [ ] `POST /api/v1/organizations/:orgId/users/:userId/deactivate` - Deactivate user

Update handlers:
- [ ] All handlers to accept and validate org_id parameter
- [ ] All handlers to include organization middleware checks

### UserInvitationHandler - CREATE
Location: `internal/handlers/user_invitation_handler.go`

**Endpoints**:
- [ ] `POST /api/v1/organizations/:orgId/users/invite` - Send invitation (ORG_ADMIN)
- [ ] `GET /api/v1/organizations/:orgId/invitations` - List pending (ORG_ADMIN)
- [ ] `DELETE /api/v1/organizations/:orgId/invitations/:inviteId` - Revoke (ORG_ADMIN)
- [ ] `POST /api/v1/invitations/:token/accept` - Accept invitation (Public)
- [ ] `POST /api/v1/invitations/:token/resend` - Resend invitation (ORG_ADMIN)

Handlers:
- [ ] `SendInvitation(c *gin.Context)`
- [ ] `ListPendingInvitations(c *gin.Context)`
- [ ] `RevokeInvitation(c *gin.Context)`
- [ ] `AcceptInvitation(c *gin.Context)`
- [ ] `ResendInvitation(c *gin.Context)`

### AuditLogHandler - CREATE
Location: `internal/handlers/audit_log_handler.go`

**Endpoints** (ORG_ADMIN+):
- [ ] `GET /api/v1/organizations/:orgId/audit-logs` - Get audit trail

Handlers:
- [ ] `GetAuditLogs(c *gin.Context)` with filtering by:
  - [ ] User ID
  - [ ] Action type
  - [ ] Resource type
  - [ ] Date range

---

## 🛣️ Route Configuration

Update `internal/server/routes.go`:
- [ ] Register organization routes (protected by SUPER_ADMIN middleware)
- [ ] Register user management routes (protected by ORG_ADMIN middleware)
- [ ] Register invitation routes
- [ ] Register audit log routes (protected by ORG_ADMIN middleware)

---

## 🧪 Testing

### Unit Tests
- [ ] Service tests (table-driven):
  - [ ] OrganizationService CRUD operations
  - [ ] UserInvitationService invitation flow
  - [ ] AuditLogService logging
  - [ ] UserService with org context
  - [ ] Error cases and validation

- [ ] Middleware tests:
  - [ ] RequireOrganization validates org access
  - [ ] RequireRole validates permissions
  - [ ] Auth middleware includes org claims
  - [ ] Invalid tokens rejected

- [ ] Handler tests:
  - [ ] Successful requests return 200
  - [ ] Unauthorized requests return 403
  - [ ] Invalid data returns 400
  - [ ] Not found returns 404
  - [ ] SUPER_ADMIN can access all orgs
  - [ ] ORG_ADMIN only sees their org

### Integration Tests
- [ ] Create org → Create users → Verify isolation
- [ ] Invite user → Accept invitation → Verify role assigned
- [ ] Promote user → Verify audit logged
- [ ] Cross-org access attempt → Verify denied
- [ ] JWT token includes org_id → Verify used in authorization

### Coverage Goals
- [ ] Services: 85%+ coverage
- [ ] Handlers: 80%+ coverage
- [ ] Middleware: 90%+ coverage

---

## ✅ Acceptance Criteria

- [ ] All services created and working
- [ ] All handlers created and working
- [ ] All authorization middleware working correctly
- [ ] All API endpoints tested and working
- [ ] Unit tests passing (80%+ coverage)
- [ ] Integration tests passing
- [ ] JWT tokens include org context
- [ ] Multi-org data isolation verified
- [ ] Error handling consistent across endpoints
- [ ] API documentation updated
- [ ] Code review approved
- [ ] No blocking issues from Phase 1

---

## 📋 Files Created/Updated

**New Files**:
```
internal/services/organization_service.go
internal/services/user_invitation_service.go
internal/services/audit_log_service.go
internal/handlers/organization_handler.go
internal/handlers/user_invitation_handler.go
internal/handlers/audit_log_handler.go
internal/middleware/organization.go
```

**Updated Files**:
```
internal/services/user_service.go
internal/handlers/user_handler.go
internal/middleware/auth.go
internal/server/routes.go
```

---

## 🔗 Related

**Depends on**: `[Phase 1] feat(admin): Database & Data Models` ✅
**Parent Epic**: `[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy`
**Blocks**: `[Phase 3] feat(admin): Frontend Components & UI`

---

## 📊 Progress Checklist

- [ ] All services created
- [ ] All handlers created
- [ ] All middleware created
- [ ] All routes configured
- [ ] Unit tests written & passing
- [ ] Integration tests written & passing
- [ ] API endpoints tested manually
- [ ] Documentation updated
- [ ] Code review approved
- [ ] Merged to develop

---

**Effort**: High | **Duration**: 5-7 days
**Status**: Ready to Start (after Phase 1)
**Last Updated**: March 3, 2026
