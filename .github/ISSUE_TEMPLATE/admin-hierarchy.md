---
name: "feat(admin): Implement 4-Level Admin Hierarchy & Multi-Tenancy [EPIC]"
about: Implement robust admin hierarchy system and multi-organization support (4-phase epic)
title: "[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy"
labels: ["enhancement", "high-priority", "backend", "frontend", "database", "epic"]
---

## 📋 Overview

**[PARENT EPIC]** Implement a robust admin hierarchy system and multi-organization (multi-tenancy) support for Tenantly. This is a foundational feature enabling proper role-based access control (RBAC) at both platform and organization levels.

**Current State**: Single-organization system with 3 roles (Admin, PropertyManager, Accountant)
**Target State**: Multi-organization platform with 5-role hierarchy

**Total Effort**: 15-22 days | **Phases**: 4 sequential | **Risk**: High

---

## 🎯 Vision

Transform Tenantly from a single-org system to a scalable multi-org platform with clear admin hierarchy:

```
SUPER_ADMIN (Platform Level)
    ↓
ORG_ADMIN (Organization Level)
    ↓
ADMIN / PropertyManager / Accountant (Organization Users)
```

---

## 📊 New Role Hierarchy

| Level | Role | Scope | Permissions |
|-------|------|-------|-------------|
| 0 | SUPER_ADMIN | Platform | Manage all organizations, assign org admins |
| 1 | ORG_ADMIN | Organization | Manage users, org settings, team structure |
| 2 | ADMIN | Organization | Core data management (properties, units, leases) |
| 3 | PropertyManager | Organization | Property ops, tenant management, lease handling |
| 4 | Accountant | Organization | Financial data (read-only), payments, reports |

---

## 🔗 Child Issues (Blocking Tasks)

This epic is broken down into 4 child issues. Complete in order:

1. **[Phase 1]** feat(admin): Database & Data Models
   - Create organizations, invitations, audit_logs tables
   - Add org_id to existing tables
   - Create migrations
   - **Blocks**: Phase 2
   - **Effort**: High | **Duration**: 2-3 days

2. **[Phase 2]** feat(admin): Backend API Implementation
   - Create services & handlers
   - Implement authorization middleware
   - Build API endpoints
   - **Blocks**: Phase 3
   - **Effort**: High | **Duration**: 5-7 days
   - **Depends on**: Phase 1 ✅

3. **[Phase 3]** feat(admin): Frontend Components & UI
   - Update role models & services
   - Create org management components
   - User onboarding & invitation UI
   - **Blocks**: Phase 4
   - **Effort**: High | **Duration**: 5-7 days
   - **Depends on**: Phase 2 ✅

4. **[Phase 4]** feat(admin): Testing, Security & Deployment
   - Database migration testing
   - Integration & E2E tests
   - Security audit
   - Staged rollout
   - **Effort**: Medium | **Duration**: 3-5 days
   - **Depends on**: Phase 3 ✅

---

## 📋 Implementation: 4 Phases

### Phase 1: Database & Data Models
**Effort**: High | **Duration**: 2-3 days
**Child Issue**: `[Phase 1] feat(admin): Database & Data Models`

Create new tables:
- [ ] `organizations` table
- [ ] `user_invitations` table
- [ ] `audit_logs` table

Update existing tables:
- [ ] Add `org_id` to users table
- [ ] Add `org_id` to all data tables (properties, units, tenants, leases, etc.)
- [ ] Add `first_name`, `last_name`, `status` to users table

Update Go models:
- [ ] `User` struct with org context
- [ ] `Organization` struct
- [ ] `UserInvitation` struct

Database migration:
- [ ] Create migration script
- [ ] Migrate existing users to default org

---

### Phase 2: Backend Implementation (Go API)
**Effort**: High | **Duration**: 5-7 days
**Child Issue**: `[Phase 2] feat(admin): Backend API Implementation`
**Depends on**: Phase 1 ✅

Services:
- [ ] `OrganizationService` - Manage organizations
- [ ] `UserInvitationService` - Handle user invitations
- [ ] `AuditLogService` - Log admin actions
- [ ] Update `UserService` for org context

Middleware & Authorization:
- [ ] `RequireOrganization()` middleware
- [ ] `RequireRole()` middleware
- [ ] Update auth middleware to include org_id in JWT

Handlers & Endpoints:
- [ ] `OrganizationHandler` - Org CRUD (SUPER_ADMIN)
- [ ] `UserInvitationHandler` - Invitation flows
- [ ] Update `UserHandler` - org-scoped operations
- [ ] Admin management endpoints (promote/demote)

API Endpoints:
- [ ] Organization endpoints: POST/GET/PUT/DELETE
- [ ] User management: List, Invite, Update, Remove
- [ ] Invitation endpoints: Accept, List pending, Revoke
- [ ] Audit logs: Retrieve audit trail

Testing:
- [ ] Unit tests for services & handlers
- [ ] Authorization middleware tests
- [ ] Integration tests for multi-org data isolation

---

### Phase 3: Frontend Implementation (Angular)
**Effort**: High | **Duration**: 5-7 days
**Child Issue**: `[Phase 3] feat(admin): Frontend Components & UI`
**Depends on**: Phase 2 ✅

Models & Services:
- [ ] Update `role.model.ts` with 5-level hierarchy
- [ ] Update `ROLE_PERMISSIONS` constants
- [ ] Create `OrganizationService`
- [ ] Update `PermissionService` for org context

Components:
- [ ] Organization Management (SUPER_ADMIN)
  - [ ] `organization-list` component
  - [ ] `organization-create` component
  - [ ] `organization-detail` component
- [ ] User Onboarding (ORG_ADMIN)
  - [ ] `invite-user` component
  - [ ] `pending-invitations` component
  - [ ] `bulk-import` component
- [ ] User Management (UPDATE)
  - [ ] `user-list` with org context
  - [ ] `admin-promotion` component
- [ ] Audit Logs (NEW)
  - [ ] `audit-log-viewer` component

Features:
- [ ] Organization selector/switcher
- [ ] User invitation form
- [ ] Bulk CSV import
- [ ] Pending invitations list
- [ ] Organization users table
- [ ] Organization statistics dashboard
- [ ] Audit log viewer

Testing:
- [ ] Unit tests for services
- [ ] Component tests
- [ ] E2E tests for critical flows

---

### Phase 4: Testing & Deployment
**Effort**: Medium | **Duration**: 3-5 days
**Child Issue**: `[Phase 4] feat(admin): Testing, Security & Deployment`
**Depends on**: Phase 3 ✅

Testing:
- [ ] Full system integration tests
- [ ] Multi-org data isolation tests
- [ ] Performance testing (<5% response time impact)
- [ ] Security audit on authorization

Migration Strategy:
- [ ] Database backup procedure
- [ ] Data migration script
- [ ] Backward compatibility testing
- [ ] Default organization creation for existing installations

Deployment:
- [ ] Update deployment documentation
- [ ] Staging environment testing
- [ ] Production rollout plan
- [ ] Monitoring & logging setup

---

## 📁 Critical Files to Create/Update

### Backend (Go)
```
internal/models/
  ├── user.go (UPDATE)
  ├── organization.go (NEW)
  └── audit_log.go (NEW)

internal/services/
  ├── user_service.go (UPDATE)
  ├── organization_service.go (NEW)
  ├── user_invitation_service.go (NEW)
  └── audit_log_service.go (NEW)

internal/handlers/
  ├── user_handler.go (UPDATE)
  ├── organization_handler.go (NEW)
  └── user_invitation_handler.go (NEW)

internal/middleware/
  ├── auth.go (UPDATE)
  └── organization.go (NEW)

migrations/
  └── 0XX_add_multi_tenancy.sql (NEW)
```

### Frontend (Angular)
```
src/app/core/models/
  └── role.model.ts (UPDATE)

src/app/core/services/
  ├── organization.service.ts (NEW)
  └── permission.service.ts (UPDATE)

src/app/features/admin/
  ├── organization-management/ (NEW)
  │   ├── organization-list.ts
  │   ├── organization-detail.ts
  │   ├── organization-create.ts
  │   └── organization-edit.ts
  ├── user-onboarding/ (NEW)
  │   ├── invite-user.ts
  │   ├── bulk-import.ts
  │   └── pending-invitations.ts
  ├── user-management/ (UPDATE)
  │   └── ...
  └── audit-logs/ (NEW)
      └── audit-log-viewer.ts
```

---

## ✅ Acceptance Criteria

- [ ] All 4 phases completed and merged
- [ ] Database migration tested with existing data
- [ ] Backward compatibility maintained for single-org installations
- [ ] All RBAC authorization working as defined
- [ ] Audit trail capturing all admin actions (create, update, delete, promote/demote)
- [ ] User invitation workflow end-to-end tested
- [ ] E2E tests for critical user journeys pass
- [ ] Performance impact <5% on API response times
- [ ] Security audit passed (authorization, data isolation, audit logs)
- [ ] Project documentation updated with new roles/endpoints
- [ ] All unit tests passing (backend & frontend)
- [ ] All integration tests passing

---

## 📚 Reference Documentation

- **Full Implementation Plan**: See `ADMIN_HIERARCHY_PLAN.md` in repository
- **Database Schema**: Detailed SQL in plan Phase 1
- **API Endpoints**: Complete list in plan Phase 2
- **Component Structure**: Frontend layout in plan Phase 3
- **Test Strategy**: Guidelines in plan Phase 4

---

## 🔐 Security Considerations

- [ ] All endpoints verify user org membership
- [ ] Super admin has unrestricted access
- [ ] Org admins can only manage their org
- [ ] Data queries filtered by organization_id
- [ ] JWT token includes org_id
- [ ] Audit logs immutable and comprehensive
- [ ] Invitation tokens expire after 7 days
- [ ] Single-use invitation tokens

---

## 🤔 Questions Before Starting

1. **Data Isolation**: Complete isolation or shared read-only access for super admin?
2. **User Mobility**: Can users belong to multiple organizations?
3. **Super Admin Access**: Can super admin see all org data or just metadata?
4. **Subscription Model**: Enforce user limits per tier?
5. **Custom Roles**: Support custom roles per organization?
6. **Invitation Expiry**: How long should invitations last (7/14/30 days)?

---

## 📊 Metrics

- **Total Estimated Effort**: 15-22 days
- **Phases**: 4 sequential phases
- **Critical Path**: Phase 1 → 2 → 3 (Phase 4 can be parallel)
- **Risk Level**: High (authorization, data integrity)
- **Breaking Changes**: Yes (database schema)

---

## 🔗 Related PRs/Issues

- Link to dependency issues/PRs here once created

---

## 🎯 Parent-Child Issue Structure

This is a **PARENT EPIC**. Create child issues in this order:

```
[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy
├── [Phase 1] feat(admin): Database & Data Models
│   └── Status: [ ] Pending
│   └── Blocks: Phase 2
├── [Phase 2] feat(admin): Backend API Implementation
│   └── Status: [ ] Pending
│   └── Depends on: Phase 1
│   └── Blocks: Phase 3
├── [Phase 3] feat(admin): Frontend Components & UI
│   └── Status: [ ] Pending
│   └── Depends on: Phase 2
│   └── Blocks: Phase 4
└── [Phase 4] feat(admin): Testing, Security & Deployment
    └── Status: [ ] Pending
    └── Depends on: Phase 3
```

**Workflow**:
1. Create parent (this issue)
2. Create Phase 1 child issue
3. Once Phase 1 merged → Create Phase 2 child issue
4. Once Phase 2 merged → Create Phase 3 child issue
5. Once Phase 3 merged → Create Phase 4 child issue
6. Once Phase 4 merged → Close parent epic ✅

---

**Status**: Ready for Planning
**Last Updated**: March 3, 2026
**Version**: 1.0 - Planning Phase (with child issues)
