# Data Security & Compliance Plan

**Last Updated:** 2026-05-19  
**Document Owner:** Security & Compliance Team  
**Status:** 🟡 In Progress (Phase 1)

---

## Executive Summary

This document tracks the implementation of security and legal compliance measures for the Tenantly payment management system, particularly focusing on:
- Protection of Personally Identifiable Information (PII)
- Payment data security
- Regulatory compliance (Bangladesh market, general GDPR-ready, PCI DSS awareness)
- Audit logging and access control

---

## Current Legal & Regulatory Context

### Applicable Regulations
- **Bangladesh Data Protection Act** - Pending (draft stage, but good practice to follow)
- **Bangladesh Telecommunications Regulation (2018)** - Phone numbers are protected
- **Bangladesh National ID Act** - NID numbers are protected government identifiers
- **General Contract Law** - Implied duty to protect tenant data
- **Banking Regulation Act** - Applies if processing bank payments
- **GDPR** - If any EU residents' data is processed (future-proofing)

### Risk Level: 🔴 HIGH (Current State)
**Reason:** Sensitive financial data (payment amounts) directly linked to unencrypted PII (phone numbers, NID) in plaintext storage.

---

## Phase 1: Data Encryption at Rest

### Status: 🟠 IN PROGRESS
**Timeline:** Started (PII_ENCRYPTION_PLAN.md exists), Estimated completion: June 30, 2026

### 1.1 - PII Encryption Infrastructure
**Description:** Implement AES-256-GCM encryption for sensitive tenant fields  
**Status:** ⚪ NOT STARTED  
**Priority:** 🔴 CRITICAL  
**Owner:** Backend Team

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Create `internal/crypto/pii.go` with Encrypt/Decrypt functions | ⚪ NOT STARTED | - | 2026-05-26 | [Reference: PII_ENCRYPTION_PLAN.md](src/backend/api/PII_ENCRYPTION_PLAN.md) Step 1 |
| Create `internal/crypto/pii_test.go` with test coverage | ⚪ NOT STARTED | - | 2026-05-26 | Round-trip, nonce, key validation tests |
| Load encryption keys in `internal/config/config.go` | ⚪ NOT STARTED | - | 2026-05-28 | Add TENANT_PII_KEY and TENANT_LOOKUP_HMAC_KEY environment variables |
| Add `.env.example` entries for encryption keys | ⚪ NOT STARTED | - | 2026-05-28 | Document key generation: `openssl rand -base64 32` |
| Database migration: Add blind index columns | ⚪ NOT STARTED | - | 2026-05-30 | `nid_number_hash`, `phone_number_hash`, `email_hash`, `address_hash` |

**Dependencies:** None (can start immediately)  
**Blocks:** Phase 1.2

---

### 1.2 - Tenant Data Encryption
**Description:** Encrypt PII fields in tenants table and implement in TenantRepository  
**Status:** ⚪ NOT STARTED  
**Priority:** 🔴 CRITICAL  
**Owner:** Backend Team

#### Fields to Encrypt:

| Field | Sensitivity | Current Status | Target Status |
|-------|-------------|-----------------|----------------|
| `nid_number` | HIGH | Plaintext | Encrypted + hashed for lookup |
| `phone_number` | HIGH | Plaintext | Encrypted + hashed for lookup |
| `email` | MEDIUM | Plaintext | Encrypted + hashed for lookup |
| `address` | LOW-MEDIUM | Plaintext | Encrypted + hashed for lookup |
| `name` | LOW | Plaintext | Keep plaintext (needed for display) |

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Extend `TenantRepository.Create()` to encrypt PII on write | ⚪ NOT STARTED | - | 2026-06-06 | Encrypt before INSERT, compute blind indexes |
| Extend `TenantRepository.Update()` to re-encrypt on modify | ⚪ NOT STARTED | - | 2026-06-06 | Encrypt new values, update blind indexes |
| Extend `TenantRepository.GetByID()` to decrypt on read | ⚪ NOT STARTED | - | 2026-06-06 | Decrypt after SELECT from database |
| Update lookup queries to use blind indexes | ⚪ NOT STARTED | - | 2026-06-08 | `GetByPhone()`, `GetByNID()`, `GetByEmail()` use hashes not plaintext |
| Update `TenantRepository` tests to assert ciphertext storage | ⚪ NOT STARTED | - | 2026-06-08 | Verify DB stores ciphertext, not plaintext |

**Dependencies:** Phase 1.1  
**Blocks:** Phase 1.3

---

### 1.3 - Data Migration (One-Time)
**Description:** Encrypt all existing tenant records in production database  
**Status:** ⚪ NOT STARTED  
**Priority:** 🔴 CRITICAL (Only after 1.1 & 1.2 deployed)  
**Owner:** DevOps + Backend Team

#### Process:

| Step | Status | Timeline | Notes |
|------|--------|----------|-------|
| Create data migration tool: `cmd/migrate-pii/main.go` | ⚪ NOT STARTED | After 1.2 | Reads plaintext, encrypts, writes ciphertext + hashes |
| Test migration on staging database | ⚪ NOT STARTED | Before production | Verify round-trip, no data loss |
| Schedule maintenance window | ⚪ NOT STARTED | June 2026 | Zero-downtime if possible, else notify users |
| Run migration in production | ⚪ NOT STARTED | 2026-06-22 | With transaction rollback on error |
| Verify all data encrypted post-migration | ⚪ NOT STARTED | After migration | Sample spot-checks of encrypted fields |

**Dependencies:** Phase 1.1, 1.2, Backend deployment  
**Blocks:** Phase 1.4

---

### 1.4 - Finalization
**Description:** Enforce encryption constraints and cleanup  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 HIGH  
**Owner:** Backend Team

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Add NOT NULL constraints to blind index columns | ⚪ NOT STARTED | After migration | Post-migration SQL |
| Remove plaintext detection shim from code (if used) | ⚪ NOT STARTED | After migration | Assume all data is encrypted |
| Update documentation: PII_ENCRYPTION_PLAN → COMPLETED | ⚪ NOT STARTED | 2026-06-23 | Update status in doc |

**Dependencies:** Phase 1.3  
**Blocks:** Phase 2

---

## Phase 2: Payment Data Security

### Status: ⚪ NOT STARTED
**Timeline:** Estimated: July 1 - July 31, 2026  
**Dependencies:** Phase 1 (PII encryption)

### 2.1 - Payment Audit Logging Enhancement
**Description:** Enhance audit logs to redact/protect sensitive payment data  
**Status:** 🟡 IN PROGRESS — PII redaction complete; admin audit view & access control pending  
**Priority:** 🟡 HIGH  
**Owner:** Backend Team

#### Current State:
- ✅ Payment access IS logged (CREATE, GET, UPDATE, DELETE)
- ✅ Audit values redacted before write — app-level (`maskAuditValue`) + DB trigger (`mask_audit_json`, migration 000004)
- ❌ No distinction between "admin can view full audit" vs "user cannot view their own audit logs"

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Redact PII/secrets from logged values | ✅ COMPLETE | - | - | Done centrally in `internal/database/audit.go` (`maskAuditValue`) — covers all audited actions, not only payments (03cc332) |
| Redact sensitive fields on payment audit writes | ✅ COMPLETE | - | - | Covered by the central `AuditService.logAudit` redaction; every caller incl. payment access is masked (03cc332) |
| Add admin audit log view with access control | ⚪ NOT STARTED | - | 2026-07-14 | Only SUPER_ADMIN/ORG_ADMIN can view full audit logs |
| Extend `audit_log` table with `redacted` flag | ⚪ NOT STARTED | - | 2026-07-10 | Track which logs have PII removed |
| Update audit tests to verify redaction | ✅ COMPLETE | - | - | `internal/database/audit_test.go` asserts sensitive values are absent from the log JSON (03cc332) |

**Dependencies:** Phase 1  
**Blocks:** Phase 2.2

---

### 2.2 - Data Retention & Deletion Policy
**Description:** Define and implement data retention rules for payment records  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 HIGH  
**Owner:** Product + Legal Team

#### Policy Definition:

| Data Type | Retention Period | Reason | Deletion Method |
|-----------|------------------|--------|-----------------|
| Payment Records | 7 years | Bangladesh tax law | Soft delete (mark `deleted_at`), archive |
| Audit Logs | 3 years | Compliance | Hard delete after 3 years |
| Backups | 1 year | Disaster recovery | Encrypted, kept in secure storage |
| Temporary Payment Processing Data | 30 days | PCI best practice | Hard delete after processing |

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Draft `DATA_RETENTION_POLICY.md` | ⚪ NOT STARTED | - | 2026-07-10 | Define retention periods per data type |
| Implement soft-delete for payments (add `deleted_at`) | ⚪ NOT STARTED | - | 2026-07-14 | Payment archival, not hard delete |
| Create audit log cleanup job | ⚪ NOT STARTED | - | 2026-07-21 | Cron job to delete logs older than 3 years |
| Add backup encryption validation | ⚪ NOT STARTED | - | 2026-07-21 | Verify backups are encrypted at rest |
| Document backup retention in runbook | ⚪ NOT STARTED | - | 2026-07-21 | Where backups stored, encryption method, retention |

**Dependencies:** Phase 2.1  
**Blocks:** Phase 3

---

### 2.3 - Payment Access Control Audit
**Description:** Verify role-based access control is correctly implemented  
**Status:** ✅ PARTIALLY COMPLETE (Code review done, testing pending)  
**Priority:** 🟡 MEDIUM  
**Owner:** Backend Team (QA)

#### Current Implementation:
- ✅ Handlers check `CanUserAccessPayment()` before GET/UPDATE
- ✅ Accountants restricted to read-only (UpdatePayment returns 403)
- ✅ Organization isolation enforced (`org_id` filter in queries)
- ❌ No explicit test for "User A cannot access Org B payments"
- ❌ No test for "Accountant cannot update payments"

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Write integration tests for cross-org access denial | 🟡 IN PROGRESS | 2026-05-15 | 2026-05-26 | Verify User A cannot access Org B payments |
| Write integration tests for accountant read-only | 🟡 IN PROGRESS | 2026-05-15 | 2026-05-26 | Verify accountants cannot UPDATE payments |
| Test PropertyManager vs Admin payment access differences | ⚪ NOT STARTED | - | 2026-05-26 | Verify role hierarchy is enforced |
| Document access control matrix in code comments | ⚪ NOT STARTED | - | 2026-05-29 | Clear who can do what (CREATE, GET, UPDATE, DELETE) |

**Dependencies:** None (ongoing)  
**Blocks:** None (parallel to Phase 1)

---

## Phase 3: Compliance Documentation & Policy

### Status: ⚪ NOT STARTED
**Timeline:** Estimated: August 1 - August 15, 2026

### 3.1 - Privacy Policy (Tenant-Facing)
**Description:** Create customer-facing privacy & security notice  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 HIGH  
**Owner:** Product/Legal Team

#### Content:

```markdown
# Data Privacy & Security Policy

## What Data We Collect
- Tenant name, phone, email, national ID, address
- Lease details (unit, start/end date, rent amount)
- Payment history (amounts, dates, status)

## How We Protect It
- Data encrypted with AES-256-GCM encryption
- Access restricted to authorized staff by role
- All access logged and audited
- Regular backups stored securely

## Your Rights
- Request your data
- Correct inaccurate data
- Request deletion (subject to legal retention periods)
- Contact: privacy@tenantly.com
```

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Draft privacy policy | ⚪ NOT STARTED | - | 2026-08-01 | Include encryption, access control, retention |
| Legal review of privacy policy | ⚪ NOT STARTED | - | 2026-08-08 | Bangladesh legal counsel review |
| Add privacy policy to frontend (Tenant Account section) | ⚪ NOT STARTED | - | 2026-08-12 | Link from settings, add i18n (English + Bangla) |
| Create data request handling process | ⚪ NOT STARTED | - | 2026-08-12 | How tenants request their data |

**Dependencies:** Phase 2  
**Blocks:** None (communication)

---

### 3.2 - Internal Security Guidelines
**Description:** Document data handling procedures for staff  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 MEDIUM  
**Owner:** Admin/HR Team

#### Content:

```markdown
# Internal Data Security Guidelines

## Who Can Access What
- SUPER_ADMIN: All data across all organizations
- ORG_ADMIN: All data within their organization
- Admin: Property/building/unit data, no payment editing
- PropertyManager: Limited to their property's data
- Accountant: Read-only access to payments

## What NOT to Do
- Never export payment data to spreadsheets without encryption
- Never share passwords or auth tokens
- Never log into the system on public WiFi without VPN
- Never discuss tenant financial info in public areas

## If You Suspect a Breach
1. Do not attempt to fix it yourself
2. Notify IT/Security immediately: security@tenantly.com
3. Document what you saw (date, time, affected data)
4. Cooperate with incident investigation
```

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Draft internal security guidelines | ⚪ NOT STARTED | - | 2026-08-05 | Access control, data handling, breach response |
| Add to employee onboarding checklist | ⚪ NOT STARTED | - | 2026-08-12 | All new staff must acknowledge |
| Create security training module | ⚪ NOT STARTED | - | 2026-08-15 | Online training with quiz |

**Dependencies:** Phase 2  
**Blocks:** None

---

### 3.3 - Incident Response Plan
**Description:** Document how to respond to security breaches  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 HIGH  
**Owner:** Security Team

#### Key Elements:

| Element | Status | Target |
|---------|--------|--------|
| Detection procedures | ⚪ NOT STARTED | 2026-08-08 |
| Notification timeline (tenants, authorities) | ⚪ NOT STARTED | 2026-08-08 |
| Investigation procedures | ⚪ NOT STARTED | 2026-08-10 |
| Remediation steps | ⚪ NOT STARTED | 2026-08-10 |
| Communication templates | ⚪ NOT STARTED | 2026-08-12 |

#### Tasks:

| Task | Status | Started | Target | Notes |
|------|--------|---------|--------|-------|
| Draft incident response playbook | ⚪ NOT STARTED | - | 2026-08-08 | Steps for breach discovery to resolution |
| Identify notification contacts (Bangladesh authorities) | ⚪ NOT STARTED | - | 2026-08-08 | Legal, IT security, tenants, regulators |
| Create breach notification template | ⚪ NOT STARTED | - | 2026-08-10 | Email/SMS template for affected tenants |
| Test incident response (tabletop drill) | ⚪ NOT STARTED | - | 2026-08-15 | Simulate breach, verify response procedures |

**Dependencies:** Phase 2  
**Blocks:** None

---

## Phase 4: Ongoing Monitoring & Maintenance

### Status: ⚪ NOT STARTED
**Timeline:** Ongoing (starts September 2026)

### 4.1 - Regular Security Audits
**Description:** Quarterly code reviews and penetration testing  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 MEDIUM  
**Owner:** Security Team

#### Tasks:

| Task | Frequency | Status | Next Review | Notes |
|------|-----------|--------|-------------|-------|
| Code security review (payment endpoints) | Quarterly | ⚪ NOT STARTED | 2026-09-15 | OWASP Top 10 check |
| Dependency security scanning (Go + npm packages) | Monthly | ⚪ NOT STARTED | 2026-06-15 | `npm audit`, `go vuln` |
| Access control audit (who has what permissions) | Semi-annually | ⚪ NOT STARTED | 2026-09-15 | Verify no orphaned access |
| Database backup integrity test | Quarterly | ⚪ NOT STARTED | 2026-06-15 | Restore backup, verify data |

---

### 4.2 - Key Rotation
**Description:** Periodically rotate encryption keys  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 MEDIUM  
**Owner:** DevOps/Security

#### Schedule:

| Key | Rotation Interval | Status | Next Rotation | Method |
|-----|------------------|--------|----------------|--------|
| TENANT_PII_KEY | Annually | ⚪ NOT STARTED | 2027-05-19 | Re-encrypt all data with new key |
| TENANT_LOOKUP_HMAC_KEY | Annually | ⚪ NOT STARTED | 2027-05-19 | Recompute all blind indexes |
| JWT Signing Key | 6 months | ⚪ NOT STARTED | 2026-11-19 | Rotate in auth service |

---

### 4.3 - Compliance Reporting
**Description:** Track compliance status and prepare audit reports  
**Status:** ⚪ NOT STARTED  
**Priority:** 🟡 MEDIUM  
**Owner:** Compliance Officer

#### Tasks:

| Task | Frequency | Status | Notes |
|------|-----------|--------|-------|
| Update compliance checklist | Monthly | ⚪ NOT STARTED | Track all items in this plan |
| Generate audit log summary | Quarterly | ⚪ NOT STARTED | Stats: access attempts, denials, changes |
| Prepare compliance report for stakeholders | Annually | ⚪ NOT STARTED | Summary of security posture, improvements |

---

## Risk Assessment Matrix (Current vs. Target)

### Current State (May 2026):

| Risk | Severity | Impact | Mitigation |
|------|----------|--------|-----------|
| Plaintext PII storage | 🔴 CRITICAL | Full tenant identity exposure + financial data | Phase 1 in progress |
| Audit logs leaking PII | 🟢 MITIGATED | PII/secrets redacted at write, both layers (03cc332); admin-view access control still pending | Phase 2.1 partial |
| No data retention policy | 🟠 HIGH | Unclear legal compliance, unlimited data storage | Phase 2.2 planned |
| Access control untested | 🟡 MEDIUM | Potential unauthorized access not caught | Phase 2.3 in progress |
| No incident response plan | 🟡 MEDIUM | Delayed response to breaches | Phase 3.3 planned |

### Target State (September 2026):

| Item | Target Status | Comments |
|------|---------------|----------|
| PII encryption | ✅ COMPLETE | All sensitive fields encrypted at rest |
| Audit logging | ✅ COMPLETE | PII redacted from audit logs |
| Access control | ✅ TESTED | Role-based access verified by tests |
| Data retention | ✅ DOCUMENTED | 7-year payment retention, 3-year audit logs |
| Incident response | ✅ READY | Playbook documented and tested |
| Privacy policy | ✅ PUBLISHED | Tenant-facing and staff guidelines available |

---

## Dependencies & Critical Path

```
Phase 1.1 (Crypto) 
  └─→ Phase 1.2 (Tenant Encryption) 
      └─→ Phase 1.3 (Data Migration) 
          └─→ Phase 1.4 (Finalization)
              └─→ Phase 2.1 (Audit Logging)
                  └─→ Phase 2.2 (Data Retention)
                      └─→ Phase 3 (Documentation)
                          └─→ Phase 4 (Monitoring)

Phase 2.3 (Access Control Testing) - PARALLEL to Phase 1
```

**Critical Path:** Phase 1 → Phase 2.1/2.2 → Phase 3 → Phase 4  
**Estimated Total Duration:** 16 weeks (May 26 - September 15, 2026)

---

## How to Update This Plan

### When Starting a Task:
```markdown
| Task Name | Status | Started | Target | Notes |
|-----------|--------|---------|--------|-------|
| Task | ⚪ NOT STARTED | - | 2026-05-26 | Notes |
↓ CHANGE TO:
| Task | 🟡 IN PROGRESS | 2026-05-20 | 2026-05-26 | Notes |
```

### When Completing a Task:
```markdown
| Task | 🟡 IN PROGRESS | 2026-05-20 | 2026-05-26 | Notes |
↓ CHANGE TO:
| Task | ✅ COMPLETE | 2026-05-20 | 2026-05-25 | PR#123: Commit hash abc1234 |
```

### Phase Status Legend:
- 🔴 BLOCKED - Waiting on dependencies
- ⚪ NOT STARTED - Ready to start
- 🟡 IN PROGRESS - Currently being worked on
- 🟠 PARTIALLY COMPLETE - Some tasks done, some pending
- ✅ COMPLETE - All tasks done, verified

---

## Approval & Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Security Lead | [TBD] | [ ] | |
| Engineering Lead | [TBD] | [ ] | |
| Product Lead | [TBD] | [ ] | |
| Legal Counsel | [TBD] | [ ] | |

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-19 | Claude Code | Initial plan created based on security audit |

---

## References

- [PII_ENCRYPTION_PLAN.md](src/backend/api/PII_ENCRYPTION_PLAN.md) - Detailed encryption implementation guide
- [AUTHENTICATION.md](src/backend/api/AUTHENTICATION.md) - Auth/JWT details
- [CLAUDE.md](CLAUDE.md) - Project overview and architecture
- Recent commits: Payment CRUD, i18n, security enhancements

---

**Last Updated:** 2026-05-19  
**Next Review:** 2026-05-26 (weekly during Phase 1)  
**Document Status:** 🟡 ACTIVE - Updated as progress is made
