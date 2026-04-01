# Audit Logging Enhancement Plan

**Status**: Future Enhancement (Post-Phase 3)
**Priority**: Medium
**Scope**: Observability & Compliance

---

## Overview

Audit logging is currently stubbed in Phase 3 (Angular UI placeholder) with backend models and repository layer ready (Phase 1). This document outlines the strategy to enhance audit logging with a robust third-party service for production use.

---

## Current State (Phase 3)

### Backend ✅
- `AuditLog` model in `internal/models/audit_log.go`
- Database schema with `audit_logs` table
- Events captured: organization CRUD, user invitations, role changes
- Local storage in PostgreSQL for legal/compliance requirements

### Frontend 🚧
- `audit-log.model.ts` - Interface definition
- `audit-log.service.ts` - API client (stub)
- `audit-log-viewer.ts` - Component (stub, no UI)
- No audit log visualization in current app

**Gap**: No UI to query/view audit logs; events only stored in DB

---

## Enhancement Strategy

### Phase 1: Third-Party Integration (Future)

**Recommended Service: Axiom.co**

- **Why**: Generous free tier (25GB/month), excellent UI, easy REST API, perfect for audit trails
- **Cost**: Free tier + pay-as-you-go ($0.34/GB after 25GB)
- **Backup Options**: ELK Stack (self-hosted), Grafana Loki (lightweight)

#### Integration Steps

1. **Backend (Go)**
   - Add Axiom client dependency
   - Create `internal/services/audit_logger_service.go` to forward events
   - Update audit event handlers to send to Axiom (async, non-blocking)
   - Fallback: Always log to PostgreSQL first, Axiom as secondary

2. **Configuration**
   - Add to `.env`: `AXIOM_API_TOKEN`, `AXIOM_DATASET`, `AXIOM_API_URL`
   - Feature flag: `ENABLE_AXIOM_AUDIT_LOGGING` (default: false)

3. **Event Payload** (sent to Axiom)
   ```json
   {
     "timestamp": "2026-04-01T10:30:00Z",
     "organization_id": 1,
     "user_id": 42,
     "user_email": "admin@example.com",
     "action": "ORGANIZATION_CREATED",
     "resource_type": "organization",
     "resource_id": 5,
     "resource_name": "Acme Corp",
     "ip_address": "203.0.113.45",
     "status": "success",
     "changes": { "name": "Acme Corp", "subscription_tier": "professional" }
   }
   ```

---

### Phase 2: UI Enhancement (Post-Phase 3)

**Before Phase 4 (acceptance invitations)**

1. **Update `audit-log-viewer.ts` Component**
   - Replace stub with real MatTable implementation
   - Columns: Timestamp, User Email, Action, Resource Type, Resource Name, Status
   - Filters: Action type dropdown, date range picker, user email search
   - Pagination: 50 events per page
   - Expandable rows for full `changes` JSON payload

2. **Two Data Sources** (Configurable)
   - **Local (PostgreSQL)**: Always available, faster queries
   - **Remote (Axiom)**: When enabled, more advanced analytics/retention

3. **Service Updates**
   ```typescript
   // audit-log.service.ts
   getAuditLogs(orgId, filters?)
     // Queries PostgreSQL by default
     // Falls back to Axiom if `useRemote` flag set

   exportAuditLogs(orgId)
     // CSV export from local DB
     // Option to export from Axiom (30-day window)
   ```

---

## Implementation Timeline

| Phase | Task | Timeline | Owner |
|-------|------|----------|-------|
| Phase 3 (Current) | Models & stubs | Complete ✅ | Done |
| Phase 4 | Axiom integration (Go backend) | Month 2 | Backend |
| Phase 4+ | UI implementation (Angular) | Month 2-3 | Frontend |
| Phase 5+ | Advanced analytics dashboard | Month 4+ | Frontend |

---

## Service Comparison

| Service | Type | Free Tier | Retention | UI | Setup |
|---------|------|-----------|-----------|----|----|
| **Axiom.co** | Cloud | 25GB/month | 14 days | ⭐⭐⭐⭐⭐ | Easy |
| **ELK Stack** | Self-Hosted | Unlimited | Configurable | ⭐⭐⭐⭐ | Medium |
| **Grafana Loki** | Self-Hosted | Unlimited | Configurable | ⭐⭐⭐ | Hard |
| **Datadog** | Cloud | 5GB/day | 3 days | ⭐⭐⭐⭐⭐ | Easy |

**Recommendation**: Start with Axiom (Phase 4), migrate to ELK if self-hosted becomes requirement.

---

## Database Retention Strategy

### Current (PostgreSQL)
- Keep all audit logs indefinitely (compliance requirement)
- Monthly archive to cold storage (S3/GCS) for logs older than 90 days
- Indexed on `organization_id`, `created_at` for fast queries

### With Axiom
- Forward to Axiom (14-day retention)
- PostgreSQL remains source of truth (permanent record)
- Axiom UI for recent audits, DB queries for historical

---

## Security & Compliance

- ✅ No PII in audit logs (except email for user context)
- ✅ Logs immutable (append-only in DB)
- ✅ Access control: Only ORG_ADMIN+ can view org's audit logs
- ✅ Encryption: HTTPS to Axiom, PostgreSQL SSL connections
- ✅ GDPR: Implement data retention policy (e.g., auto-delete after 7 years)

---

## File References

**Backend (Phase 1, Complete)**
- `internal/models/audit_log.go` - Model definition
- `internal/repositories/audit_log_repository.go` - Database access
- `internal/models/audit_log.go` - Structs
- `migrations/000007_admin_hierarchy_phase1.up.sql` - Schema

**Frontend (Phase 3, Stubs)**
- `src/frontend/src/app/core/models/audit-log.model.ts` - TS interface
- `src/frontend/src/app/core/services/audit-log.service.ts` - API client
- `src/frontend/src/app/features/admin/audit-logs/audit-log-viewer.ts` - Component stub

**Future Work**
- `internal/services/audit_logger_service.go` - Axiom integration
- `src/frontend/src/app/features/admin/audit-logs/audit-log-viewer.html` - Template
- `src/frontend/src/app/features/admin/audit-logs/audit-log-viewer.spec.ts` - Tests

---

## Success Criteria

- [ ] Axiom integration compiles without errors
- [ ] Audit events forward to Axiom (configurable, off by default)
- [ ] PostgreSQL audit log table remains primary record
- [ ] UI component displays 50+ events with filtering & sorting
- [ ] Expandable rows show full change payloads
- [ ] Export to CSV includes timestamp, user, action, resource, status
- [ ] Performance: Audit log queries < 500ms for 90-day window
- [ ] All unit tests pass (new + existing)

---

## Notes

- Axiom API token must be secrets-managed (not in repo)
- Async logging: Don't block requests on Axiom failures
- Consider alerting if audit logging fails silently
- Plan for GDPR right-to-be-forgotten: implement audit log deletion workflow
- Monitor Axiom costs as event volume grows (scale to ELK if needed)

