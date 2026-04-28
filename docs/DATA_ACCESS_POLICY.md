# Data Access Policy — Tenantly

## Legal Basis (Bangladesh)

### Applicable Law

| Law | Relevant Provision |
|-----|--------------------|
| **Digital Security Act 2018 (DSA)** | Section 26 — unauthorised access to computer data; Section 35 — data breach liability |
| **ICT Act 2006 (amended 2013)** | Section 56 — hacking / unauthorised system access |
| **Draft Personal Data Protection Act (Bangladesh)** | Data minimisation, purpose limitation, processor obligations (pending enactment; treated as anticipated compliance baseline) |
| **Bangladesh Bank BRPD Circular** | Financial transaction data of tenants must be accessible only to authorised parties within the data controller's organisation |

### Key Legal Distinctions

**Data Controller** — The property management company (an Organisation within Tenantly) that collects and decides the purpose of processing tenant personal data.

**Data Processor** — Tenantly SaaS platform (the operator). Processes data solely on behalf of the data controller. Has no independent legal basis to access, read, or aggregate tenant personal data across organisations.

Under DSA Section 26, access to computer data without authorisation — even by a platform operator — constitutes an offence. The data controller (Organisation) has not authorised the platform operator to read individual tenant records unless a signed Data Processing Agreement (DPA) explicitly grants that right for a specific support purpose.

---

## Business Rules

### BR-001 — SUPER_ADMIN Scope Boundary

**Rule:** `SUPER_ADMIN` is a platform administration role. It governs the Tenantly platform itself, not the business data of any Organisation.

**SUPER_ADMIN MAY:**
- Create, configure, suspend, or delete Organisations
- Create and deactivate `ORG_ADMIN` accounts
- View organisation-level metadata: name, subscription tier, user count, creation date
- View system audit logs (actor, action, timestamp, table name — not record content)
- Initiate a time-limited break-glass impersonation session (see BR-004)

**SUPER_ADMIN MUST NOT:**
- Read tenant personal data (name, NID, phone, address, emergency contact)
- Read lease terms, rent amounts, payment history
- Read property or unit details belonging to any Organisation
- Read documents or attachments uploaded by any Organisation
- Execute cross-organisation queries that aggregate tenant or financial data

**Backend enforcement:** Any repository query for business-domain resources (properties, buildings, units, tenants, leases, payments, documents) must always include `WHERE organization_id = $1`. When the caller is `SUPER_ADMIN` and no `organization_id` context is set in the JWT, the backend MUST return an empty result set or `403 Forbidden` — never a cross-org result.

---

### BR-002 — Organisation Data Isolation

**Rule:** All business data belongs exclusively to the Organisation that created it. No role in Organisation A can access data belonging to Organisation B under any circumstances.

**Backend enforcement:** Every query for organisation-scoped resources must be parameterised by `organization_id` extracted from the authenticated JWT. Hardcoded org IDs in queries, admin bypass flags, or fallback-to-all patterns are prohibited.

---

### BR-003 — Personal Data Minimisation

**Rule:** Endpoints must return only the fields required for the calling role's function.

| Role | Tenant data visible |
|------|---------------------|
| `ORG_ADMIN` | Full profile (for compliance reporting) |
| `Admin` | Full profile |
| `PropertyManager` | Name, unit assignment, lease status |
| `Accountant` | Name, payment records only — no NID or personal contact |

Future API changes that add new fields to tenant/lease/payment responses must be reviewed against this table before merging.

---

### BR-004 — Break-Glass Access (Support)

**Rule:** When `SUPER_ADMIN` requires access to Organisation data for a legitimate support purpose, the following procedure applies:

1. `SUPER_ADMIN` calls `POST /auth/set-organization` with target `organization_id` and a documented `reason` string.
2. System issues a time-limited (max 1 hour) org-scoped JWT.
3. Every action taken under this token is logged with `impersonated_by: <superadmin_user_id>` in the audit log.
4. The Organisation's `ORG_ADMIN` receives an automated notification of the session.
5. Token cannot be refreshed — session ends when it expires.

This creates a lawful basis (legitimate interest, proportionate, logged) under the DSA and anticipated PDPA framework.

---

### BR-005 — Audit Log Access

**Rule:** `SUPER_ADMIN` may read audit logs. Audit log entries record: actor user ID, action type, table name, record ID, timestamp, IP address. They must NOT store the full `old_values` / `new_values` payload for tables containing personal data (`tenants`, `leases`, `payments`) in any endpoint accessible to `SUPER_ADMIN`. Full payloads are accessible only to `ORG_ADMIN` and `Admin` within the owning organisation.

---

## Acceptance Criteria (for any feature touching data access)

```
Given a SUPER_ADMIN token with no organization_id claim
When GET /api/v1/tenants (or properties, leases, payments, documents) is called
Then response is 403 Forbidden or empty list
And no personal data is returned

Given a SUPER_ADMIN token
When GET /api/v1/users is called
Then only platform-level user metadata is returned (username, role, org name)
And no tenant or financial data is included in the response

Given ORG_ADMIN of Organisation A
When any resource endpoint is called
Then only resources with organization_id = A are returned
And resources from Organisation B are never included, even under error conditions

Given SUPER_ADMIN performing break-glass access to Organisation A
When any action is taken
Then audit log records action with impersonated_by = superadmin_user_id
And ORG_ADMIN of Organisation A is notified
And the session token expires within 1 hour and cannot be refreshed
```

---

## Implementation Checklist

- [ ] All repository methods for `tenants`, `leases`, `payments`, `documents` filter by `organization_id` — no exceptions
- [ ] `SUPER_ADMIN` calling org-scoped endpoints without org context returns empty/403
- [ ] Break-glass impersonation endpoint implemented with audit logging and notification
- [ ] Audit log `GET` endpoint strips `old_values`/`new_values` for personal-data tables when caller is `SUPER_ADMIN`
- [ ] No frontend route or nav item exposes tenant/lease/payment data to `SUPER_ADMIN` without explicit org selection

---

> **Note:** Bangladesh's Personal Data Protection Act is pending enactment as of 2025. This policy is written to comply with the DSA 2018 currently in force and to anticipate the PDPA framework. It should be reviewed by a BD-licensed legal counsel before production deployment involving real tenant data.
