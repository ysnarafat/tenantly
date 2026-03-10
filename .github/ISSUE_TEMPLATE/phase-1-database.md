---
name: "[Phase 1] Database & Data Models"
about: "Phase 1: Create database tables and Go models for multi-tenancy"
title: "[Phase 1] feat(admin): Database & Data Models"
labels: ["enhancement", "high-priority", "database", "backend"]
---

## 📋 Overview

**[CHILD ISSUE - Phase 1 of 4]**
**Parent Epic**: `[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy`

Create foundation for multi-tenancy support by adding new tables and updating existing schemas with organization context.

**Effort**: High | **Duration**: 2-3 days
**Blocks**: Phase 2 (Backend API Implementation)

---

## 📊 Deliverables

### New Database Tables

#### 1. Organizations Table
- [ ] Create `organizations` table with:
  - id, name, slug, subscription_tier, max_users
  - active, created_at, updated_at, deleted_at
  - Unique index on slug

#### 2. User Invitations Table
- [ ] Create `user_invitations` table with:
  - id, organization_id, email, role
  - invitation_token (unique), expires_at
  - accepted_at, accepted_by_user_id
  - Foreign keys to organizations and users tables

#### 3. Audit Logs Table
- [ ] Create `audit_logs` table with:
  - id, organization_id, user_id, action
  - resource_type, resource_id, changes (JSON)
  - ip_address, user_agent, created_at
  - Foreign keys to organizations and users tables

### Alter Existing Tables

#### Users Table
- [ ] Add `organization_id` (nullable - NULL = SUPER_ADMIN)
- [ ] Add `first_name` (VARCHAR 100)
- [ ] Add `last_name` (VARCHAR 100)
- [ ] Add `status` (ENUM: active, inactive, pending_invite)
- [ ] Create index: `idx_users_org_id`
- [ ] Create index: `idx_users_email_org`

#### Data Tables (Properties, Units, Tenants, Leases, Payments, Buildings, etc.)
- [ ] Add `organization_id` to each data table
- [ ] Create indexes for org_id lookups
- [ ] Add foreign key constraints

### Database Migrations

- [ ] Create migration script: `migrations/0XX_add_multi_tenancy.sql`
- [ ] Migration includes:
  - [ ] Create organizations table
  - [ ] Create user_invitations table
  - [ ] Create audit_logs table
  - [ ] Alter users table
  - [ ] Alter data tables
  - [ ] Create all necessary indexes
  - [ ] Create foreign key constraints

#### Migration Rollback Plan
- [ ] Rollback script prepared
- [ ] Tested on staging database

---

## 🗂️ Go Model Updates

### Files to Create/Update

#### `internal/models/user.go` - UPDATE
- [ ] Update User struct with:
  ```go
  OrganizationID *int      // NULL = SUPER_ADMIN
  FirstName      string
  LastName       string
  Role           string    // Extended to include ORG_ADMIN, SUPER_ADMIN
  Status         string    // active, inactive, pending_invite
  ```

#### `internal/models/organization.go` - CREATE
- [ ] Create Organization struct with:
  - id, name, slug, subscription_tier, max_users
  - active, created_at, updated_at

#### `internal/models/audit_log.go` - CREATE
- [ ] Create AuditLog struct with:
  - id, organization_id, user_id, action
  - resource_type, resource_id, changes (JSON)
  - ip_address, user_agent, created_at

#### `internal/models/user_invitation.go` - CREATE
- [ ] Create UserInvitation struct with:
  - id, organization_id, email, role
  - token, expires_at, accepted_at, accepted_by_user_id
  - created_at

### Request/Response DTOs

- [ ] Create DTOs in respective model files:
  - `CreateOrganizationRequest`
  - `UpdateOrganizationRequest`
  - `CreateUserRequest` (updated with org_id)
  - `UpdateUserRequest` (updated with org_id)
  - `InviteUserRequest`
  - `InvitationAcceptRequest`

---

## 🧪 Testing

- [ ] Write tests for migration script:
  - [ ] Test migration runs without errors
  - [ ] Test rollback works correctly
  - [ ] Test all tables created with correct schema
  - [ ] Test indexes created properly
  - [ ] Test foreign key constraints work

- [ ] Data migration tests:
  - [ ] Existing users assigned to default org
  - [ ] All existing data maintains integrity
  - [ ] No data loss on migration

- [ ] Model tests:
  - [ ] User model marshals/unmarshals correctly
  - [ ] Organization model validates inputs
  - [ ] Audit log model handles JSON changes field

---

## 📝 SQL Reference

Key migration components needed:

```sql
-- Organizations
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

-- Alter users
ALTER TABLE users ADD COLUMN (
    organization_id INT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    status ENUM('active', 'inactive', 'pending_invite') DEFAULT 'active',
    FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Indexes
CREATE INDEX idx_users_org_id ON users(organization_id);
CREATE INDEX idx_users_email_org ON users(email, organization_id);
```

---

## ✅ Acceptance Criteria

- [ ] All new tables created with correct schema
- [ ] All alter table statements executed successfully
- [ ] All indexes created and working
- [ ] All foreign key constraints working
- [ ] Migration script runs without errors
- [ ] Rollback script tested and working
- [ ] All Go models updated/created
- [ ] All DTOs created
- [ ] Unit tests passing (100% coverage for models)
- [ ] Migration tested on staging database
- [ ] Documentation updated with new tables

---

## 📋 Files Changed

**New Files**:
```
internal/models/organization.go
internal/models/audit_log.go
internal/models/user_invitation.go
migrations/0XX_add_multi_tenancy.sql
```

**Updated Files**:
```
internal/models/user.go
```

---

## 🔗 Related

**Parent Epic**: `[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy`
**Blocks**: `[Phase 2] feat(admin): Backend API Implementation`

---

## 📊 Progress Checklist

- [ ] Database tables created
- [ ] Existing tables altered
- [ ] Migrations written & tested
- [ ] Go models created/updated
- [ ] DTOs created
- [ ] Unit tests written
- [ ] Staging database tested
- [ ] Code review approved
- [ ] Merged to develop

---

**Effort**: High | **Duration**: 2-3 days
**Status**: Ready to Start
**Last Updated**: March 3, 2026
