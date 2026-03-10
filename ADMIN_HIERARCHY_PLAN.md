# Admin Hierarchy & Multi-Tenancy Implementation Plan

## Overview

This document outlines the implementation plan for introducing a robust admin hierarchy system and multi-organization support to Tenantly. This will enable proper role-based access control (RBAC) at both platform and organization levels.

**Status**: Planning Phase
**Priority**: High
**Estimated Phases**: 3-4 phases

---

## Vision

Transform Tenantly from a single-organization system to a multi-organization platform with clear admin hierarchy:

```
SUPER_ADMIN (Platform) → ORG_ADMIN (Organization) → ADMIN/PM/ACCOUNTANT (Users)
```

---

## Architecture Overview

### Admin Hierarchy (4 Levels)

| Level | Role | Scope | Responsibilities |
|-------|------|-------|------------------|
| 0 | **SUPER_ADMIN** | Platform | Platform management, org creation, org admin assignment |
| 1 | **ORG_ADMIN** | Organization | User management within org, org settings, team management |
| 2 | **ADMIN** | Organization | Core data management (properties, units, tenants, leases) |
| 3 | **PropertyManager** | Organization | Property operations, tenant management, lease handling |
| 4 | **Accountant** | Organization | Financial data (read-only mostly), payment records, reports |

### Multi-Tenancy Model

- **Organizations**: Top-level container for all org data
- **Users**: Belong to organizations (except SUPER_ADMIN who is platform-level)
- **Data Isolation**: Properties, units, tenants, leases scoped to organizations
- **Permissions**: Role + Organization context determines access

---

## Phase 1: Database & Data Models

### 1.1 New Database Tables

```sql
-- Organizations table
CREATE TABLE organizations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    subscription_tier ENUM('basic', 'professional', 'enterprise') DEFAULT 'basic',
    max_users INT DEFAULT 50,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- User invitations table
CREATE TABLE user_invitations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    organization_id INT NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    invitation_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP NULL,
    accepted_by_user_id INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (accepted_by_user_id) REFERENCES users(id)
);

-- Audit log for admin actions
CREATE TABLE audit_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    organization_id INT,
    user_id INT,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id INT,
    changes JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 1.2 Alter Existing Users Table

```sql
ALTER TABLE users ADD COLUMN (
    organization_id INT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    status ENUM('active', 'inactive', 'pending_invite') DEFAULT 'active',
    FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Add index for faster lookups
CREATE INDEX idx_users_org_id ON users(organization_id);
CREATE INDEX idx_users_email_org ON users(email, organization_id);
```

### 1.3 Update Existing Tables with org_id

```sql
-- Properties, Units, Tenants, Leases, Payments need organization_id
ALTER TABLE properties ADD COLUMN organization_id INT NOT NULL;
ALTER TABLE buildings ADD COLUMN organization_id INT NOT NULL;
-- ... repeat for all data tables
```

### 1.4 Go Model Updates

**File**: `internal/models/user.go`

```go
package models

import "time"

type User struct {
    ID             int       `json:"id" db:"id"`
    Username       string    `json:"username" db:"username"`
    Email          string    `json:"email" db:"email"`
    PasswordHash   string    `json:"-" db:"password_hash"`
    FirstName      string    `json:"first_name" db:"first_name"`
    LastName       string    `json:"last_name" db:"last_name"`

    // Organization context
    OrganizationID *int      `json:"organization_id" db:"organization_id"`  // NULL = SUPER_ADMIN

    // Role & Status
    Role           string    `json:"role" db:"role"`      // SUPER_ADMIN, ORG_ADMIN, ADMIN, PropertyManager, Accountant
    Status         string    `json:"status" db:"status"`  // active, inactive, pending_invite
    Active         bool      `json:"active" db:"active"`

    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Request/Response DTOs
type CreateUserRequest struct {
    Username       string `json:"username" binding:"required,min=3,max=50"`
    Email          string `json:"email" binding:"required,email"`
    Password       string `json:"password" binding:"required,min=6"`
    FirstName      string `json:"first_name" binding:"required"`
    LastName       string `json:"last_name" binding:"required"`
    Role           string `json:"role" binding:"required,oneof=ORG_ADMIN ADMIN PropertyManager Accountant"`
    OrganizationID int    `json:"organization_id" binding:"required"`
}

type InviteUserRequest struct {
    Email          string `json:"email" binding:"required,email"`
    Role           string `json:"role" binding:"required,oneof=ORG_ADMIN ADMIN PropertyManager Accountant"`
    OrganizationID int    `json:"organization_id" binding:"required"`
}

type UpdateUserRequest struct {
    FirstName  string `json:"first_name" binding:"omitempty"`
    LastName   string `json:"last_name" binding:"omitempty"`
    Email      string `json:"email" binding:"omitempty,email"`
    Role       string `json:"role" binding:"omitempty,oneof=ORG_ADMIN ADMIN PropertyManager Accountant"`
    Status     string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type Organization struct {
    ID               int       `json:"id" db:"id"`
    Name             string    `json:"name" db:"name"`
    Slug             string    `json:"slug" db:"slug"`
    SubscriptionTier string    `json:"subscription_tier" db:"subscription_tier"`
    MaxUsers         int       `json:"max_users" db:"max_users"`
    Active           bool      `json:"active" db:"active"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type UserInvitation struct {
    ID             int       `json:"id" db:"id"`
    OrganizationID int       `json:"organization_id" db:"organization_id"`
    Email          string    `json:"email" db:"email"`
    Role           string    `json:"role" db:"role"`
    Token          string    `json:"token" db:"token"`
    ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
    AcceptedAt     *time.Time `json:"accepted_at" db:"accepted_at"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
```

---

## Phase 2: Backend Implementation

### 2.1 Services to Create/Update

**New Services**:
- `OrganizationService` - Manage organizations
- `UserInvitationService` - Handle user invitations
- `AuditLogService` - Log admin actions

**Updated Services**:
- `UserService` - Add org context
- Existing services - Add org_id filtering

### 2.2 Middleware & Authorization

**File**: `internal/middleware/organization.go` (NEW)

```go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// RequireOrganization ensures user belongs to requested org
func RequireOrganization() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        orgID, _ := c.Get("organization_id")
        requestedOrgID := c.Param("organization_id")

        // Validate org access
        if !isUserInOrganization(userID.(int), requestedOrgID) {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// RequireRole ensures user has required role
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, _ := c.Get("user_role")
        roleStr := role.(string)

        hasRole := false
        for _, r := range requiredRoles {
            if roleStr == r {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 2.3 Handlers to Create/Update

**New Handlers**:
- `OrganizationHandler` - CRUD for organizations (SUPER_ADMIN only)
- `UserInvitationHandler` - Send/accept invitations
- `AdminManagementHandler` - Promote/demote admins

**Updated Handlers**:
- `UserHandler` - Add org context to all endpoints

### 2.4 API Endpoints

**Organization Management** (SUPER_ADMIN only):
```
POST   /api/organizations                    # Create org
GET    /api/organizations                    # List all orgs
GET    /api/organizations/:id                # Get org details
PUT    /api/organizations/:id                # Update org
DELETE /api/organizations/:id                # Delete org
```

**User Management** (ORG_ADMIN within org):
```
GET    /api/organizations/:orgId/users                      # List org users
POST   /api/organizations/:orgId/users/invite               # Invite user
GET    /api/organizations/:orgId/users/:userId              # Get user details
PUT    /api/organizations/:orgId/users/:userId              # Update user
DELETE /api/organizations/:orgId/users/:userId              # Remove user
POST   /api/organizations/:orgId/users/:userId/promote      # Make ORG_ADMIN
POST   /api/organizations/:orgId/users/:userId/demote       # Remove admin status
```

**Invitations**:
```
POST   /api/invitations/:token/accept        # Accept invitation (public endpoint)
GET    /api/organizations/:orgId/invitations # List pending invitations
DELETE /api/organizations/:orgId/invitations/:inviteId # Revoke invitation
```

**Audit Logs** (ORG_ADMIN+):
```
GET    /api/organizations/:orgId/audit-logs  # Get audit trail
```

---

## Phase 3: Frontend Implementation

### 3.1 Update Role Models (Angular)

**File**: `src/app/core/models/role.model.ts`

```typescript
// Updated role types
export type UserRole = 'SUPER_ADMIN' | 'ORG_ADMIN' | 'Admin' | 'PropertyManager' | 'Accountant';

// Role hierarchy
export const ROLE_HIERARCHY: Record<UserRole, number> = {
  SUPER_ADMIN: 0,
  ORG_ADMIN: 1,
  Admin: 2,
  PropertyManager: 3,
  Accountant: 4,
};

// Permissions
export enum Permission {
  // Platform management
  MANAGE_ORGANIZATIONS = 'manage_organizations',
  MANAGE_ORG_ADMINS = 'manage_org_admins',

  // User Management
  MANAGE_USERS = 'manage_users',
  VIEW_USERS = 'view_users',
  INVITE_USERS = 'invite_users',

  // ... existing permissions
}

// Updated role permissions
export const ROLE_PERMISSIONS: Record<UserRole, Permission[]> = {
  SUPER_ADMIN: [
    Permission.MANAGE_ORGANIZATIONS,
    Permission.MANAGE_ORG_ADMINS,
    Permission.MANAGE_USERS,
    // ... ALL permissions
  ],
  ORG_ADMIN: [
    Permission.MANAGE_USERS,
    Permission.INVITE_USERS,
    Permission.MANAGE_PROPERTIES,
    // ... org-level permissions
  ],
  // ... rest of roles
};
```

### 3.2 New Components

```
src/app/features/admin/
├── organization-management/        (NEW - SUPER_ADMIN)
│   ├── organization-list/
│   ├── organization-detail/
│   ├── organization-create/
│   └── organization-edit/
├── user-onboarding/               (NEW - ORG_ADMIN)
│   ├── invite-user/
│   ├── bulk-import/
│   └── pending-invitations/
├── user-management/               (UPDATE)
│   ├── user-list/
│   ├── user-detail/
│   └── admin-promotion/
└── audit-logs/                    (NEW)
    └── audit-log-viewer/
```

### 3.3 Super Admin Onboarding View

**Component**: `SuperAdminOnboardingComponent`

Features:
- Organization selector dropdown
- User invitation form with fields:
  - Email
  - First Name
  - Last Name
  - Role (ORG_ADMIN, ADMIN, PropertyManager, Accountant)
  - Department (optional)
- Bulk CSV import
- Pending invitations list
- Organization users table with:
  - Username, Email, Role
  - Status badge (Active/Inactive/Pending)
  - Last login
  - Action buttons (Edit, Deactivate, Delete, Resend Invite)
- Organization statistics:
  - Total users
  - User breakdown by role
  - Pending invitations count

### 3.4 Organization Context Service (NEW)

**File**: `src/app/core/services/organization.service.ts`

```typescript
@Injectable({ providedIn: 'root' })
export class OrganizationService {
  private currentOrganization$ = new BehaviorSubject<Organization | null>(null);

  constructor(private http: HttpClient) {}

  // Get current org
  getCurrentOrganization(): Observable<Organization> {
    return this.currentOrganization$.asObservable();
  }

  // Set current org
  setCurrentOrganization(org: Organization) {
    this.currentOrganization$.next(org);
  }

  // API calls
  getOrganizations(): Observable<Organization[]> { ... }
  createOrganization(org: CreateOrgRequest): Observable<Organization> { ... }
  updateOrganization(id: number, org: UpdateOrgRequest): Observable<Organization> { ... }
}
```

### 3.5 Update PermissionService

**File**: `src/app/core/services/permission.service.ts`

- Add organization context checking
- Add role hierarchy checking
- Update permission validation to consider organization scope

---

## Phase 4: Testing & Deployment

### 4.1 Testing Strategy

**Unit Tests**:
- Role hierarchy validation
- Permission checking with org context
- User invitation flow
- Audit logging

**Integration Tests**:
- Multi-org data isolation
- Authorization middleware
- API endpoint access control

**E2E Tests** (Cypress/Angular Testing):
- Super admin org creation
- User invitation workflow
- Organization switching
- Permission-based UI rendering

### 4.2 Database Migration Strategy

1. **Pre-migration**: Backup existing database
2. **Add columns**: org_id to users and data tables
3. **Data migration**: Assign all existing users to default org
4. **Create default org**: For existing installations
5. **Enable constraints**: Add foreign keys and indexes
6. **Verify**: Run integration tests

### 4.3 Backward Compatibility

- Support single-org installations initially
- Super admin can create multiple orgs
- Existing users map to default organization
- Gradual rollout to multi-org

---

## Security Considerations

### Authorization

- [ ] All endpoints verify user org membership
- [ ] Super admin has unrestricted access
- [ ] Org admins can only manage their org
- [ ] Data queries filtered by organization_id
- [ ] JWT token includes org_id

### Audit Trail

- [ ] Log all admin actions (create, update, delete, promote/demote)
- [ ] Include user, timestamp, IP, user-agent
- [ ] Audit logs immutable
- [ ] Accessible only to org admins and super admin

### Invitation Security

- [ ] Tokens expire after 7 days
- [ ] Single-use tokens
- [ ] Email verification recommended
- [ ] Bulk import file validation

---

## Implementation Checklist

### Phase 1: Database
- [ ] Create organizations table
- [ ] Create user_invitations table
- [ ] Create audit_logs table
- [ ] Add org_id to users table
- [ ] Add org_id to all data tables
- [ ] Create migration script

### Phase 2: Backend
- [ ] Create OrganizationService
- [ ] Create UserInvitationService
- [ ] Create AuditLogService
- [ ] Update UserService for org context
- [ ] Create OrganizationHandler
- [ ] Create UserInvitationHandler
- [ ] Update UserHandler endpoints
- [ ] Create authorization middleware
- [ ] Write unit tests
- [ ] Write integration tests

### Phase 3: Frontend
- [ ] Update role.model.ts
- [ ] Create OrganizationService
- [ ] Update PermissionService
- [ ] Create organization management components
- [ ] Create user onboarding components
- [ ] Create super admin dashboard
- [ ] Update existing components for org context
- [ ] Write unit tests
- [ ] Write e2e tests

### Phase 4: Testing & Deployment
- [ ] Full system testing
- [ ] Performance testing
- [ ] Security audit
- [ ] Migration testing
- [ ] Staged rollout plan
- [ ] Monitoring & logging

---

## Future Enhancements

1. **Teams/Departments**: Sub-organization structure within orgs
2. **Custom Roles**: Allow org admins to create custom roles
3. **SSO Integration**: Enterprise authentication (OAuth, SAML)
4. **Two-Factor Authentication**: Enhanced security for admins
5. **Role-Based Dashboards**: Customized views per role
6. **Notification System**: Email/SMS for onboarding status
7. **User Activity Tracking**: Comprehensive user analytics
8. **Compliance Reporting**: GDPR, HIPAA compliance tools

---

## References

- [Tenantly GitHub](https://github.com/ysnarafat/tenantly)
- [Current User Model](src/backend/api/internal/models/user.go)
- [Current Role Model](src/frontend/src/app/core/models/role.model.ts)
- [RBAC Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html)

---

## Questions & Decisions Needed

Before implementation, clarify:

1. **Data Isolation Level**: Complete isolation or shared read-only access?
2. **User Mobility**: Can users belong to multiple organizations?
3. **Super Admin Access**: Can super admin see all org data or just metadata?
4. **Subscription Model**: Enforce user limits per tier?
5. **Custom Roles**: Support custom roles per organization?
6. **Invitation Expiry**: How long should invitations last (7, 14, 30 days)?

---

## Contact & Questions

For questions or clarifications about this plan, please refer to the team or create a GitHub issue.

**Last Updated**: March 2, 2026
**Version**: 1.0 - Planning Phase
