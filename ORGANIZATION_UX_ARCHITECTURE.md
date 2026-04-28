# Organization Context & Multi-Organization User Support
## UX and Architectural Analysis

**Date:** 2026-04-05  
**Status:** Design Decision Needed

---

## Current State

### Architecture
```
User Model:
├── id
├── username
├── email
├── organization_id (single, NOT NULL)  ← user belongs to exactly ONE org
├── role (Admin, PropertyManager, Accountant)
└── ...

Login Flow:
├── POST /api/v1/auth/login
│   └── Input: username + password (NO organization parameter)
│   └── Returns: token with user.organization_id embedded in JWT claims
└── User can access only resources in their organization_id
```

### Current Limitation
- **User can belong to exactly ONE organization**
- No organization selection during login
- No way to switch organizations after login
- System assumes 1:1 user-to-organization relationship

---

## Problem Statement

**Three interrelated questions:**

1. **How does the system know which organization the user belongs to?**
   - Currently: Implicit in the database record (`User.organization_id`)
   - No explicit selection needed at login

2. **Should users switch organizations?**
   - Currently: No mechanism exists
   - Question: Should it be possible?

3. **Is the 1:1 user-to-organization design optimal?**
   - Currently: Yes, by design
   - Question: Is this the right long-term choice?

---

## Option 1: Current Design (Keep As-Is)
### User belongs to exactly ONE organization

**How it works:**
```
Login:
  username + password → Find User → Load user.organization_id → Embed in JWT
  
Access Control:
  Every request carries organization_id in JWT
  Middleware validates: "Is this resource in my org?"
  
Organization Switch:
  Not possible - user is locked to their organization
```

### Pros ✅
- **Simplest architecture** — No multi-org logic to manage
- **Clear ownership** — Each user belongs to exactly one org
- **Lower complexity** — Fewer edge cases and bugs
- **Performance** — No need to query which orgs user can access
- **Clear intent** — Eliminates confusion about "which org am I in?"

### Cons ❌
- **Bad for organizations with multiple roles**
  - Property manager at HQ can't manage properties at branches
  - Accountant for the whole company can't see all locations
  - Common real-world need in property management
  
- **Difficult team management**
  - Can't have one person (e.g., consultant) working for multiple clients
  - Forces duplicate user accounts per organization
  
- **Onboarding friction**
  - New users can only belong to one org at creation
  - If they need to move orgs later, account must be recreated

### UX Experience
```
Login Page:
┌─────────────────────┐
│ Username/Email: ___ │
│ Password: _________ │
│ [Login Button]      │
└─────────────────────┘

↓ (automatic org context)

Dashboard:
┌────────────────────────┐
│ Welcome, John!         │
│ Organization: ABC Ltd  │ ← Fixed, can't change
│ (no org selector)      │
└────────────────────────┘
```

---

## Option 2: Organization Selection at Login
### User belongs to ONE org, but explicitly selects at login

**How it works:**
```
Step 1 - Username/Password:
  POST /api/v1/auth/login
  Body: { username, password }
  Response: { organizations: [...available orgs...] }

Step 2 - Select Organization:
  POST /api/v1/auth/select-organization
  Body: { organization_id }
  Response: { token (with selected org_id), user }
  
OR use a 2-step login form on frontend
```

### Pros ✅
- **Prepares for multi-org future** — Framework already in place
- **Better UX if users belong to multiple orgs** (see Option 3)
- **Explicit choice** — User consciously confirms which org they're accessing
- **Security benefit** — User sees available orgs (helps catch account misuse)

### Cons ❌
- **Extra step in login flow** — Slower user experience
- **Unnecessary if user truly has ONE org**
  - Adds friction without benefit
  - Extra API call per login
  
- **Still doesn't solve multi-org user problem**
  - User still belongs to only one org
  - Still need duplicate accounts for multi-org users

### UX Experience
```
Login Page - Step 1:
┌─────────────────────┐
│ Username/Email: ___ │
│ Password: _________ │
│ [Login Button]      │
└─────────────────────┘

        ↓ (call: /auth/login)

Organization Selection - Step 2:
┌──────────────────────────────────┐
│ Select Organization:             │
│ ☐ ABC Ltd                        │ ← User can only access ONE
│ ☐ XYZ Properties                 │ ← So this extra step is
│ ☐ MyRental Group                 │    unnecessary friction
│ [Continue Button]                │
└──────────────────────────────────┘

        ↓ (call: /auth/select-organization)

Dashboard:
┌────────────────────────┐
│ Welcome, John!         │
│ Organization: ABC Ltd  │
│ [Can't switch]         │
└────────────────────────┘
```

---

## Option 3: Users Belong to MULTIPLE Organizations (Recommended)
### User can be member of multiple orgs, switch between them

**How it works:**

```
Database Schema:
CREATE TABLE users (
  id, username, email, password_hash, ...
  -- NO organization_id here!
);

CREATE TABLE user_organization_roles (
  id, user_id, organization_id, role, created_at
  -- User 1 → Org A as PropertyManager
  -- User 1 → Org B as Accountant
  -- User 1 → Org C as Admin
);

Login Flow:
  1. username + password → Find User
  2. Load all (organization_id, role) pairs from user_organization_roles
  3. Return: { token, organizations: [...], user }
     token = base token (not org-specific yet)

After Login - Organization Selection:
  1. Frontend displays: "Select Organization"
  2. User picks: ABC Ltd
  3. Call: /auth/set-organization { organization_id }
  4. Get back: new token WITH organization_id + role embedded
  5. Store in JWT: { user_id, organization_id, role, permissions }

Switching Organizations:
  1. User clicks: "Switch Organization" in UI
  2. Call: /auth/set-organization { organization_id }
  3. Get back: new token for selected org
  4. All subsequent requests use new token
```

### Architecture Changes Required

**Backend:**
1. Create `user_organization_roles` junction table
2. Modify `User` model: remove `organization_id`, add `organizations[]`
3. Update login: return list of accessible organizations
4. New endpoint: `POST /auth/set-organization` (select/switch org)
5. Update all permission checks to use JWT's current `organization_id`

**Frontend:**
1. After login, show org selector (skip if user has only 1 org)
2. Add "Switch Organization" UI element
3. Store current `organization_id` from JWT
4. Support seamless org switching

### Pros ✅
- **Real-world use cases** 
  - Consultant works for 3 clients simultaneously
  - Area manager oversees multiple property locations
  - Finance person consolidated view across orgs
  
- **Better multi-location support**
  - Single account for multi-branch operations ✅
  - Easy to add/remove org access (no account recreation)
  
- **Scalable** — Prepares for enterprise features
  - Cross-org reporting
  - Consolidated financial views (for admins)
  
- **Better onboarding** — User has one account for all organizations
  
- **Eliminates duplicate accounts**

### Cons ❌
- **More complex** — Another junction table, more queries
- **UX complexity** — Users need to understand they can switch orgs
- **Extra login step if user has multiple orgs** — But can be skipped if only 1
- **More API calls** — Extra call to set organization after login
- **Requires migration** — Can't be added later without data migration

### UX Experience (Optimized)

**Login Flow:**
```
┌─────────────────────┐
│ Username: John      │
│ Password: _________ │
│ [Login]             │
└─────────────────────┘
        ↓ (Single step - no org selection screen)
```

**Dashboard (Right After Login):**
```
┌────────────────────────────────────────────┐
│ Tenantly                      ☰ Menu       │
├────────────────────────────────────────────┤
│ Welcome, John!                             │
│                                            │
│ Organization: [ABC Ltd ▼]  ← Org Selector │
│               • XYZ Properties             │
│               • MyRental Group             │
│               • Manage Organizations       │
│                                            │
│ [Dashboard Content for ABC Ltd...]         │
├────────────────────────────────────────────┤
│ Properties | Buildings | Users | Reports   │
└────────────────────────────────────────────┘
```

**Organization Switching:**
```
User clicks [ABC Ltd ▼] dropdown
  ↓
Select "XYZ Properties"
  ↓
Dashboard reloads with XYZ context
  ↓
All subsequent requests use new org token
```

### Why This Approach is Better ✅

**vs. Organization Selection at Login:**
- **Faster login** — Single step (username/password), no intermediate screen
- **Familiar UX pattern** — Matches Slack, GitHub, Jira, Linear (industry standard)
- **Reduced friction** — Users see dashboard immediately, org switching is natural navigation
- **Better for common case** — Most users work in one org primarily; no extra step
- **Cleaner separation** — Authentication focuses on credentials; organization context is a dashboard concern
- **Power user friendly** — Multi-org users can switch with one click from dashboard

This approach treats **organization selection as navigation, not authentication** — which is the correct mental model.

---

## Comparison Matrix

| Aspect | Option 1 (Current) | Option 2 (Selection) | Option 3 (Multi-Org) |
|--------|-------------------|----------------------|----------------------|
| **Setup Complexity** | Lowest | Low | Moderate |
| **Login Steps** | 1 | 2 | 1-2 (smart) |
| **Organization Switch** | ❌ Not possible | ❌ Not possible | ✅ Yes |
| **Multi-org Users** | ❌ Duplicate accounts | ❌ Duplicate accounts | ✅ Single account |
| **Real Estate Fit** | ⚠️ Limited | ⚠️ Limited | ✅ Excellent |
| **Scalability** | Good | Good | Better |
| **Code Complexity** | Simplest | Medium | Higher |
| **Data Migration** | None | None | Required |

---

## Recommendation for Tenantly (Property Management)

### ✅ **Go with Option 3: Multi-Organization Support**

**Why:**

1. **Property management inherently multi-location:**
   - Companies manage multiple properties/buildings in different locations
   - Same manager, accountant, or admin may work across locations
   - Area managers oversee multiple branches

2. **Real-world user needs:**
   - Property Manager A should manage Properties in Location 1, 2, and 3
   - Finance person needs consolidated view of all properties
   - Franchisees may own multiple locations under one account

3. **Future-proofs the system:**
   - Allows for cross-org reporting later
   - Supports partner/franchise models
   - Enterprise features (consolidated views, bulk operations)

4. **Better UX:** 
   - Single account = simpler password management
   - Users don't need multiple logins
   - "Switch Organization" is familiar pattern (think Slack, GitHub, Jira)

5. **Not much harder than current:**
   - One junction table
   - Two extra API endpoints
   - Modest backend changes

---

## Implementation Roadmap

### Phase 1: Add Multi-Org Support (Next Sprint)
- [ ] Create `user_organization_roles` table with migration
- [ ] Update User model and repositories
- [ ] Modify login endpoint to return `organizations: []` + default org token
- [ ] Create `POST /auth/set-organization` endpoint for switching orgs
- [ ] Update JWT claims to handle org-specific tokens
- [ ] Backfill: assign existing users to their current orgs
- [ ] Add logic to select default organization (most recent, first alphabetical, etc.)

### Phase 2: Frontend Organization Switching (Sprint After) ⭐ Recommended Approach
- [ ] **Add organization dropdown selector in dashboard menu** (NOT in login)
- [ ] Dropdown shows: current org (✓), list of other accessible orgs, "Manage Organizations"
- [ ] Clicking org in dropdown: calls `/auth/set-organization` → new token → dashboard refreshes
- [ ] Update JWT handling: extract and store current `organization_id` from token
- [ ] Update dashboard to show current organization context in header
- [ ] Handle edge case: if default org access revoked, fallback to next available org
- [ ] Smooth transitions: org switching doesn't lose user context (stays on same page if applicable)

### Phase 3: Enhanced Features (Future)
- [ ] Cross-organization reporting (for SUPER_ADMIN)
- [ ] Bulk operations across orgs
- [ ] Organization-level analytics
- [ ] User preference: set default organization

---

## Recommended Implementation Details

### Backend Flow (Option 3 - Dashboard-Based Selection)

```
Step 1: Login
────────────
POST /api/v1/auth/login
Request:  { username, password }
Response: {
  token: "jwt_with_default_org",
  refresh_token: "...",
  user: { id, name, email, ... },
  organizations: [
    { id: 1, name: "ABC Ltd", role: "PropertyManager" },
    { id: 2, name: "XYZ Properties", role: "Admin" },
    { id: 3, name: "MyRental Group", role: "Accountant" }
  ],
  default_organization_id: 1,
  expires_at: "..."
}

JWT Contains: { user_id, organization_id: 1, role: "PropertyManager", ... }
```

```
Step 2: User Selects Different Organization
──────────────────────────────────────────────
POST /api/v1/auth/set-organization
Request:  { organization_id: 2 }
Headers:  Authorization: Bearer jwt_with_org_1

Response: {
  token: "jwt_with_org_2",        ← New token for org 2
  refresh_token: "...",
  organization: { id: 2, name: "XYZ Properties", role: "Admin" },
  expires_at: "..."
}

JWT Now Contains: { user_id, organization_id: 2, role: "Admin", ... }
```

### Frontend Flow

```typescript
// Step 1: User logs in
login(username, password) {
  POST /auth/login
  ├─ Store token (org 1)
  ├─ Store organizations list
  ├─ Navigate to dashboard
  └─ Dashboard loads with org 1 selected

// Step 2: User switches organization from menu
switchOrganization(organizationId: number) {
  POST /auth/set-organization { organization_id }
  ├─ Receive new token (org N)
  ├─ Update stored token
  ├─ Update NgRx state with new org_id
  ├─ Refresh dashboard data for new org
  └─ User sees same dashboard, different org's data
}

// Step 3: On page refresh / new session
// JWT still contains organization_id, so current org context is preserved
```

### Key Implementation Points

1. **Default Organization Selection**
   ```go
   // Option: User's most recently accessed org
   // Option: First alphabetical
   // Option: Primary organization (set during org creation)
   // Option: Org with highest role/access
   
   // Recommendation: Most recent + fallback to alphabetical
   ```

2. **Token Generation**
   ```go
   // Always include current org_id in token
   token.Claims["organization_id"] = organizationId
   token.Claims["role"] = userRole  // for this org
   
   // Keep user_id and username org-agnostic
   token.Claims["user_id"] = userId
   token.Claims["username"] = username
   ```

3. **Middleware Enforcement**
   ```go
   // Validate org_id from token
   orgID := c.GetInt64("organization_id")
   
   // All repository queries filtered by this org
   properties := propertyRepo.ListByOrganization(ctx, orgID)
   ```

4. **Seamless Switching** (Frontend)
   ```typescript
   // When org changes, you can:
   // a) Keep user on same page, reload data for new org
   // b) Redirect to org-specific dashboard if structure differs
   // Recommendation: Option (a) for better UX
   ```

---

## Immediate Question for Your Team

**Before we implement, decide:**

1. **Do you expect users to belong to multiple organizations?**
   - If YES → Option 3 is right
   - If NO (user always belongs to exactly one org) → Option 1 is simplest

2. **Will you support franchises/partners with multiple locations?**
   - If YES → Option 3 needed
   - If NO → Option 1 sufficient

3. **Do you have enterprise clients with multiple branches?**
   - If YES → Option 3 essential
   - If NO → Option 1 can work

---

## Current Issue in Code

Your C1 security vulnerability (cross-tenant data leakage) is partly a result of the **implicit** org context:

```go
// Current code has no explicit org context flow
// Just assumes organization_id is in JWT
c.Get("organization_id")  // ← Where does this come from?
```

With Option 3, this becomes much clearer:
```go
// Token explicitly created for this organization
orgID := c.GetInt64("organization_id")  // ← Explicit, enforced at every step
```

---

## Summary

| Current Design (Option 1) | Best Design (Option 3) |
|--------------------------|------------------------|
| User ↔ Organization (1:1) | User ↔ Organizations (1:N) |
| Fixed at creation | Flexible, changeable |
| No switching | Seamless switching |
| Limits real-world use | Enables multi-location ops |
| Simpler code | More scalable code |

**Recommendation: Implement Option 3 while fixing security issues. The junction table pattern is industry-standard and will save you from a painful refactor later.**

---

## Follow-Up Resources

- [Auth0 - Multi-Organization Pattern](https://auth0.com/docs/manage-users/user-accounts/user-account-linking/multi-tenancy-considerations)
- [AWS Multi-Tenancy Best Practices](https://docs.aws.amazon.com/whitepapers/latest/multi-tenant-saas/multi-tenancy-and-isolation.html)
- [Slack's Organization Switching Pattern](https://slack.com/help/articles/212475109-Join-and-leave-workspaces) (good UX reference)
