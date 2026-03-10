---
name: "[Phase 3] Frontend Components & UI"
about: "Phase 3: Update role models, create org/user management components"
title: "[Phase 3] feat(admin): Frontend Components & UI"
labels: ["enhancement", "high-priority", "frontend"]
---

## 📋 Overview

**[CHILD ISSUE - Phase 3 of 4]**
**Parent Epic**: `[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy`

Implement Angular components and services for organization management, user onboarding, and admin hierarchy UI.

**Effort**: High | **Duration**: 5-7 days
**Depends on**: Phase 2 ✅
**Blocks**: Phase 4 (Testing & Deployment)

---

## 🗂️ Models & Constants - UPDATE

### `src/app/core/models/role.model.ts`
- [ ] Update role types to include SUPER_ADMIN and ORG_ADMIN:
  ```typescript
  export type UserRole = 'SUPER_ADMIN' | 'ORG_ADMIN' | 'Admin' | 'PropertyManager' | 'Accountant';
  ```

- [ ] Create role hierarchy constant:
  ```typescript
  export const ROLE_HIERARCHY: Record<UserRole, number> = {
    SUPER_ADMIN: 0,
    ORG_ADMIN: 1,
    Admin: 2,
    PropertyManager: 3,
    Accountant: 4,
  };
  ```

- [ ] Update permission enum with new permissions:
  - [ ] `MANAGE_ORGANIZATIONS` - Create, update, delete organizations
  - [ ] `MANAGE_ORG_ADMINS` - Promote/demote org admins
  - [ ] `MANAGE_USERS` - Create, update, delete users in org
  - [ ] `VIEW_USERS` - View organization users
  - [ ] `INVITE_USERS` - Send user invitations

- [ ] Update ROLE_PERMISSIONS mapping with new permissions

### `src/app/core/models/organization.model.ts` - CREATE
- [ ] Create `Organization` interface:
  ```typescript
  export interface Organization {
    id: number;
    name: string;
    slug: string;
    subscriptionTier: 'basic' | 'professional' | 'enterprise';
    maxUsers: number;
    active: boolean;
    createdAt: Date;
    updatedAt: Date;
  }
  ```

- [ ] Create DTOs: `CreateOrgRequest`, `UpdateOrgRequest`

### `src/app/core/models/user-invitation.model.ts` - CREATE
- [ ] Create `UserInvitation` interface:
  ```typescript
  export interface UserInvitation {
    id: number;
    organizationId: number;
    email: string;
    role: UserRole;
    token: string;
    expiresAt: Date;
    acceptedAt?: Date;
    createdAt: Date;
  }
  ```

---

## 🔧 Services

### `src/app/core/services/organization.service.ts` - CREATE
- [ ] Methods:
  - [ ] `getOrganizations(): Observable<Organization[]>`
  - [ ] `getOrganization(id: number): Observable<Organization>`
  - [ ] `createOrganization(req: CreateOrgRequest): Observable<Organization>`
  - [ ] `updateOrganization(id: number, req: UpdateOrgRequest): Observable<Organization>`
  - [ ] `deleteOrganization(id: number): Observable<void>`
  - [ ] `getOrganizationUsers(orgId: number): Observable<User[]>`
  - [ ] `getOrganizationStats(orgId: number): Observable<OrgStats>`

- [ ] `currentOrganization$: BehaviorSubject<Organization>` for current org context

- [ ] Methods to manage current organization:
  - [ ] `getCurrentOrganization(): Observable<Organization>`
  - [ ] `setCurrentOrganization(org: Organization): void`

### `src/app/core/services/user-invitation.service.ts` - CREATE
- [ ] Methods:
  - [ ] `sendInvitation(orgId: number, req: InviteUserRequest): Observable<UserInvitation>`
  - [ ] `getPendingInvitations(orgId: number): Observable<UserInvitation[]>`
  - [ ] `revokeInvitation(inviteId: number): Observable<void>`
  - [ ] `acceptInvitation(token: string, password: string): Observable<User>`
  - [ ] `resendInvitation(inviteId: number): Observable<void>`
  - [ ] `validateInvitationToken(token: string): Observable<UserInvitation>`

### `src/app/core/services/permission.service.ts` - UPDATE
- [ ] Update permission checking to include organization context:
  - [ ] `hasPermission(permission: Permission, orgId?: number): boolean`
  - [ ] `canManageUser(targetUser: User, currentOrg: Organization): boolean`
  - [ ] `canPromoteUser(targetUser: User, currentOrg: Organization): boolean`

- [ ] Add role hierarchy validation:
  - [ ] `isRoleHigherThan(role1: UserRole, role2: UserRole): boolean`
  - [ ] `getRolePermissions(role: UserRole): Permission[]`

### `src/app/core/services/audit-log.service.ts` - CREATE
- [ ] Methods:
  - [ ] `getAuditLogs(orgId: number, filters?: AuditFilter): Observable<AuditLog[]>`
  - [ ] `exportAuditLogs(orgId: number): Observable<Blob>`

---

## 🎨 New Components

### Organization Management (SUPER_ADMIN only)

#### 1. `src/app/features/admin/organization-management/organization-list.ts`
- [ ] Display list of all organizations
- [ ] Columns: Name, Slug, Subscription Tier, Users Count, Status, Actions
- [ ] Features:
  - [ ] Pagination
  - [ ] Search/filter by name or slug
  - [ ] Sort by date, users count, status
  - [ ] Action buttons: View, Edit, Delete
  - [ ] Create new org button

#### 2. `src/app/features/admin/organization-management/organization-detail.ts`
- [ ] Display organization details and statistics
- [ ] Show:
  - [ ] Org info (name, slug, tier, max users)
  - [ ] Total users breakdown by role
  - [ ] Pending invitations count
  - [ ] Activity timeline
  - [ ] Last synced date
- [ ] Actions: Edit, Delete, View Users, View Audit Logs

#### 3. `src/app/features/admin/organization-management/organization-create.ts`
- [ ] Form to create new organization
- [ ] Fields:
  - [ ] Name (required)
  - [ ] Slug (auto-generate from name, editable)
  - [ ] Subscription Tier (basic, professional, enterprise)
  - [ ] Max Users (number)
- [ ] Validation and error handling
- [ ] Success notification and redirect

#### 4. `src/app/features/admin/organization-management/organization-edit.ts`
- [ ] Form to edit existing organization
- [ ] Same fields as create form
- [ ] Disable slug edit (for integrity)
- [ ] Confirmation dialog for changes

---

### User Onboarding & Invitation (ORG_ADMIN)

#### 5. `src/app/features/admin/user-onboarding/invite-user.ts`
- [ ] Form to invite single user
- [ ] Fields:
  - [ ] Email (required, email validation)
  - [ ] First Name (required)
  - [ ] Last Name (required)
  - [ ] Role (dropdown: ORG_ADMIN, ADMIN, PropertyManager, Accountant)
  - [ ] Department (optional)
- [ ] Features:
  - [ ] Email validation against existing users
  - [ ] Prevent duplicate invitations
  - [ ] Copy invitation link to clipboard
  - [ ] Resend button if already invited
- [ ] Success notification showing invitation sent

#### 6. `src/app/features/admin/user-onboarding/pending-invitations.ts`
- [ ] Table of pending invitations
- [ ] Columns: Email, Role, Invited Date, Expires Date, Status, Actions
- [ ] Features:
  - [ ] Pagination
  - [ ] Search by email
  - [ ] Copy invitation link
  - [ ] Resend invitation button
  - [ ] Revoke invitation button
  - [ ] Filter by status (pending, expired, sent)
- [ ] Show count of pending invitations

#### 7. `src/app/features/admin/user-onboarding/bulk-import.ts`
- [ ] CSV file upload for bulk user import
- [ ] Fields in CSV: email, first_name, last_name, role
- [ ] Features:
  - [ ] File format validation
  - [ ] Preview uploaded data before import
  - [ ] Show validation errors (duplicate emails, invalid roles)
  - [ ] Import button with progress indicator
  - [ ] Summary of successful/failed imports
  - [ ] Download template CSV
- [ ] Error handling and retry

---

### User Management (ORG_ADMIN)

#### 8. `src/app/features/admin/user-management/user-list.ts` - UPDATE
- [ ] Update to show org-scoped users
- [ ] Columns: Name, Email, Role, Status, Last Login, Actions
- [ ] Features:
  - [ ] Pagination, search, filter by role/status
  - [ ] Sort by name, email, role, login date
  - [ ] User status badges (active, inactive, pending)
  - [ ] Action buttons: Edit, Deactivate, Promote/Demote, Delete
  - [ ] Bulk actions (deactivate multiple, delete multiple)

#### 9. `src/app/features/admin/user-management/admin-promotion.ts`
- [ ] Modal/dialog to promote user to ORG_ADMIN or demote
- [ ] Confirmation dialog with warning:
  - Show what permissions will be granted/revoked
  - [ ] Audit trail implication
- [ ] Success notification with audit log reference

---

### Audit Logs & Monitoring

#### 10. `src/app/features/admin/audit-logs/audit-log-viewer.ts`
- [ ] Display audit log table
- [ ] Columns: Timestamp, User, Action, Resource Type, Resource, IP Address, Status
- [ ] Features:
  - [ ] Pagination (50 rows per page)
  - [ ] Filters:
    - [ ] By action type (Create, Update, Delete, Promote, Demote)
    - [ ] By resource type (User, Organization, Property)
    - [ ] By date range
    - [ ] By user
  - [ ] Search by resource name or ID
  - [ ] Export to CSV
  - [ ] Detail view showing full change JSON
  - [ ] Sort by timestamp, action, user

---

## 🔄 Update Existing Components

### Navigation & Sidebar
- [ ] Add org selector dropdown in main header
- [ ] Show current organization name
- [ ] Add "Switch Organization" option
- [ ] Add admin menu items based on role:
  - [ ] SUPER_ADMIN: Organizations, Users
  - [ ] ORG_ADMIN: User Management, Invitations, Audit Logs

### Dashboard (HOME)
- [ ] Update to show org-specific data
- [ ] Add organization selector if SUPER_ADMIN
- [ ] Show organization statistics in dashboard

### User Profile
- [ ] Show organization affiliation
- [ ] Show user role within organization
- [ ] Display first name, last name

---

## 🧪 Testing

### Unit Tests
- [ ] Service tests (table-driven):
  - [ ] OrganizationService API calls
  - [ ] UserInvitationService validation
  - [ ] PermissionService role hierarchy
  - [ ] AuditLogService filtering

- [ ] Component tests:
  - [ ] Organization list loads and displays
  - [ ] User invitation form validates input
  - [ ] Pending invitations table renders correctly
  - [ ] Audit log viewer filters work
  - [ ] Permission checks prevent unauthorized actions

### Integration Tests
- [ ] Flow: Create org → Invite user → Accept invitation
- [ ] Flow: Promote user → Verify role change → Check audit log
- [ ] Flow: Organization selector changes org context
- [ ] Permission service correctly evaluates role hierarchy

### E2E Tests (Cypress)
- [ ] SUPER_ADMIN can create organization
- [ ] ORG_ADMIN can invite users
- [ ] User receives invitation and accepts it
- [ ] Permission-based UI elements show/hide correctly
- [ ] Audit logs record admin actions

### Coverage Goals
- [ ] Services: 85%+
- [ ] Components: 75%+
- [ ] Overall: 80%+

---

## ✅ Acceptance Criteria

- [ ] All components created and working
- [ ] All services created and working
- [ ] Role hierarchy displayed and enforced in UI
- [ ] Organization context management working
- [ ] User invitation flow end-to-end tested
- [ ] Org-scoped data display verified
- [ ] Permission-based component rendering working
- [ ] Unit tests passing (80%+ coverage)
- [ ] E2E tests passing for critical flows
- [ ] No console errors in browser DevTools
- [ ] Responsive design on mobile/tablet
- [ ] Accessibility standards met (WCAG 2.1 AA)
- [ ] Code review approved

---

## 📋 Files Created/Updated

**New Files**:
```
src/app/core/models/organization.model.ts
src/app/core/models/user-invitation.model.ts
src/app/core/services/organization.service.ts
src/app/core/services/user-invitation.service.ts
src/app/core/services/audit-log.service.ts
src/app/features/admin/organization-management/organization-list.ts
src/app/features/admin/organization-management/organization-detail.ts
src/app/features/admin/organization-management/organization-create.ts
src/app/features/admin/organization-management/organization-edit.ts
src/app/features/admin/user-onboarding/invite-user.ts
src/app/features/admin/user-onboarding/pending-invitations.ts
src/app/features/admin/user-onboarding/bulk-import.ts
src/app/features/admin/audit-logs/audit-log-viewer.ts
```

**Updated Files**:
```
src/app/core/models/role.model.ts
src/app/core/services/permission.service.ts
src/app/features/admin/user-management/user-list.ts
src/app/app.routes.ts
src/app/app.component.ts (navigation)
```

---

## 🔗 Related

**Depends on**: `[Phase 2] feat(admin): Backend API Implementation` ✅
**Parent Epic**: `[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy`
**Blocks**: `[Phase 4] feat(admin): Testing, Security & Deployment`

---

## 📊 Progress Checklist

- [ ] Models & interfaces created
- [ ] Services created
- [ ] Components created (org management)
- [ ] Components created (user onboarding)
- [ ] Components created (audit logs)
- [ ] Existing components updated
- [ ] Routes configured
- [ ] Unit tests written & passing
- [ ] E2E tests written & passing
- [ ] Styling & responsiveness complete
- [ ] Accessibility tested
- [ ] Code review approved
- [ ] Merged to develop

---

**Effort**: High | **Duration**: 5-7 days
**Status**: Ready to Start (after Phase 2)
**Last Updated**: March 3, 2026
