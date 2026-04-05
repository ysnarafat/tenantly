# Organization Selector Implementation

## Overview

This document describes the implementation of the **Organization Selector** UI component, which enables users to switch between multiple organizations they have access to.

This implementation follows **Option 3** from `ORGANIZATION_UX_ARCHITECTURE.md`: **Multi-Organization User Support with Dashboard-Based Selection**.

---

## Architecture

### Components & Services

#### 1. **OrganizationSelector Component** 
**File:** `src/app/shared/organization-selector/organization-selector.ts`

A Material Design dropdown component placed in the app toolbar that displays:
- Current organization name with domain icon
- Dropdown menu showing all accessible organizations
- Current organization marked with a checkmark
- User's role for each organization
- "Manage Organizations" link at the bottom

**Features:**
- Smooth Material Design animations
- Loading state during organization switch
- Responsive design (hides org name on mobile, shows icon only)
- Accessible markup with proper ARIA labels

#### 2. **OrganizationService**
**File:** `src/app/core/services/organization.service.ts`

Handles all organization-related API calls and local state:

```typescript
// Key methods:
setOrganization(organizationId: number)  // Switch to different org
getCurrentOrganizationId()                 // Get current org from storage
storeOrganizationContext(orgId, org)       // Persist org context
restoreOrganizationContext()               // Load from localStorage
clearOrganizationContext()                 // Clear on logout
```

#### 3. **Organization Models**
**File:** `src/app/core/models/organization.model.ts`

Data structures for organization support:

```typescript
interface Organization {
  id: number;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

interface UserOrganization {
  id: number;
  organization_id: number;
  organization: Organization;
  role: string;  // Admin, PropertyManager, Accountant
  created_at: string;
}

interface LoginResponseWithOrganizations {
  token: string;
  refresh_token: string;
  user: any;
  organizations: UserOrganization[];  // All orgs user can access
  default_organization_id: number;    // Org to load by default
  expires_at: string | Date;
}

interface SetOrganizationResponse {
  token: string;       // New token with selected org in claims
  refresh_token: string;
  organization: UserOrganization;
  expires_at: string | Date;
}
```

---

## State Management (NgRx)

### Auth Reducer Updates
**File:** `src/app/store/auth/auth.reducer.ts`

Extended AuthState with organization fields:

```typescript
export interface AuthState {
  // ... existing fields
  userOrganizations: UserOrganization[];  // List of accessible orgs
  currentOrganizationId: number | null;   // Currently selected org
}
```

### Auth Selectors Updates
**File:** `src/app/store/auth/auth.selectors.ts`

New selectors for organization data:

```typescript
export const selectUserOrganizations = createSelector(...)
export const selectCurrentOrganizationId = createSelector(...)
```

---

## UI Integration

### Toolbar Placement
**Files:** 
- `src/app/app.ts` - Import OrganizationSelector
- `src/app/app.html` - Add `<app-organization-selector>` between branding and user menu
- `src/app/app.scss` - Style the selector in toolbar

**Toolbar Layout:**
```
[Menu] [Tenantly Logo] [Organization Selector ▼] [Spacer] [User Menu]
```

**Mobile Responsive:**
- On mobile (< 768px): Shows only domain icon
- Text label hidden to save space
- Menu still fully functional

---

## User Flow

### Login
```
1. User enters credentials (username + password)
2. Backend returns login response with:
   - JWT token (contains default org_id)
   - List of all accessible organizations
   - User info
3. Frontend stores:
   - Token in localStorage
   - Organizations in NgRx state
   - Current org_id in localStorage
4. Dashboard loads with default org context
```

### Organization Switch
```
1. User clicks org selector dropdown
2. Sees list of accessible organizations
3. Clicks on different organization
4. Call: POST /api/v1/auth/set-organization { organization_id }
5. Backend returns new token (with new org_id in claims)
6. Frontend updates:
   - localStorage with new token
   - NgRx state with new current org_id
   - UI reflects new org selection
7. Dashboard data reloads automatically (due to org context in middleware)
```

---

## API Endpoints Required

### Backend Implementation Checklist

The following endpoints need to be implemented in the backend:

#### 1. Login Endpoint (Updated)
**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "john@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com"
  },
  "organizations": [
    {
      "id": 1,
      "organization_id": 1,
      "organization": {
        "id": 1,
        "name": "ABC Ltd",
        "description": "Main office"
      },
      "role": "PropertyManager"
    },
    {
      "id": 2,
      "organization_id": 2,
      "organization": {
        "id": 2,
        "name": "XYZ Properties",
        "description": "Branch office"
      },
      "role": "Admin"
    }
  ],
  "default_organization_id": 1,
  "expires_at": "2026-04-12T10:30:00Z"
}
```

#### 2. Set Organization Endpoint (New)
**Endpoint:** `POST /api/v1/auth/set-organization`

**Request:**
```json
{
  "organization_id": 2
}
```

**Response:**
```json
{
  "token": "eyJhbGc...",  // New token with org_id: 2
  "refresh_token": "eyJhbGc...",
  "organization": {
    "id": 2,
    "organization_id": 2,
    "organization": {
      "id": 2,
      "name": "XYZ Properties"
    },
    "role": "Admin"
  },
  "expires_at": "2026-04-12T10:30:00Z"
}
```

**JWT Claims (both endpoints):**
```json
{
  "user_id": 1,
  "username": "john",
  "organization_id": 2,      // Current org context
  "role": "Admin",           // Role in this org
  "permissions": [...],
  "iat": 1234567890,
  "exp": 1234571490
}
```

---

## Styling Details

### Material Design Theming
- Primary color: Material Blue (#1976D2)
- Hover states: Subtle background elevation
- Active state: Checkmark icon + blue highlight
- Smooth transitions on all interactive elements

### Component Styling

**Organization Selector Button:**
- Pill-shaped button with semi-transparent white background
- Icon + text + dropdown arrow
- Responsive: hides text on mobile
- Loading spinner on switching

**Organization Menu:**
- Header with "Organizations" label
- One menu item per accessible organization
- Shows org name + user's role in that org
- Visual indicator for current organization
- Divider before "Manage Organizations" option
- Max-height with scroll for many organizations

---

## Local Storage Keys

```typescript
'tenantly_token'           // Current JWT token
'tenantly_current_org_id'  // Current organization ID
'tenantly_current_org'     // Full organization object (cached)
'tenantly_refresh_token'   // Refresh token
'tenantly_user'            // User object
'tenantly_expires_at'      // Token expiry time
```

---

## Error Handling

### Cases Handled:

1. **Organization Access Revoked**
   - If current org access is removed while logged in
   - Falls back to first available org
   - Shows warning notification

2. **Switch Fails**
   - Network error during org switch
   - Shows error snackbar
   - Keeps previous organization selected

3. **Missing Default Org**
   - Backend returns organizations list but no default_organization_id
   - Frontend defaults to first org in list

---

## Future Enhancements

### Phase 3 (From Architecture Doc)

Once the core multi-org support is stable:

1. **Organization Management Page** (`/settings/organizations`)
   - Add new organizations
   - Leave organizations
   - See organization members
   - Change role requests

2. **Cross-Organization Features**
   - Consolidated reporting for admins
   - Bulk operations across orgs
   - Organization-level analytics

3. **User Preferences**
   - Set preferred/default organization
   - Remember last used org per device

---

## Testing Checklist

### Unit Tests
- [ ] OrganizationService methods
- [ ] Organization selector component signals
- [ ] Auth reducer with org fields
- [ ] Auth selectors for organizations

### Integration Tests
- [ ] Login flow with organizations
- [ ] Organization switching
- [ ] Token refresh with organization context
- [ ] Logout clears organization context

### E2E Tests
- [ ] User can see organization dropdown
- [ ] User can switch organizations
- [ ] Dashboard data updates on org switch
- [ ] Menu items reflect user roles per org

---

## Migration Notes

For existing systems with single-organization per user:

1. **Database Migration:**
   - Create `user_organization_roles` table
   - Migrate existing `User.organization_id` → `user_organization_roles` entries
   - Backfill data with default role assignments

2. **Code Migration:**
   - Update all org context extraction to use JWT
   - Remove hardcoded `User.organization_id` checks
   - Update API response models

3. **Rollout Strategy:**
   - Deploy backend changes first
   - Keep login backward-compatible during transition
   - Update frontend after backend is stable

---

## Files Created/Modified

### New Files
- `src/app/core/models/organization.model.ts`
- `src/app/core/services/organization.service.ts`
- `src/app/shared/organization-selector/organization-selector.ts`
- `src/frontend/ORGANIZATION_SELECTOR_IMPLEMENTATION.md` (this file)

### Modified Files
- `src/app/store/auth/auth.reducer.ts` - Added org state
- `src/app/store/auth/auth.selectors.ts` - Added org selectors
- `src/app/app.ts` - Import OrganizationSelector
- `src/app/app.html` - Add selector to toolbar
- `src/app/app.scss` - Style organization selector
- `src/app/core/models/index.ts` - Export organization model

---

## References

- [ORGANIZATION_UX_ARCHITECTURE.md](../ORGANIZATION_UX_ARCHITECTURE.md) - Complete architecture analysis
- [Angular Material Dropdown](https://material.angular.io/components/menu/)
- [NgRx Documentation](https://ngrx.io/)
- [Auth0 Multi-Organization Pattern](https://auth0.com/docs/manage-users/user-accounts/user-account-linking/multi-tenancy-considerations)
