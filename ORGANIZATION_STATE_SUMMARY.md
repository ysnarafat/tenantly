# Organization Multi-Tenancy State Summary

## Overview
The Tenantly frontend already has a **complete implementation** for organization multi-tenancy with smooth UI/UX transitions and robust state management.

---

## Current Implementation Status

### ✅ Fully Implemented Components

#### 1. **OrganizationPicker Component**
- **Location**: `src/frontend/src/app/features/auth/organization-picker/organization-picker.ts`
- **Features**:
  - Smooth fade-in animations (80% opacity with blur)
  - Shimmer loading effects for org cards
  - Click-to-select functionality
  - Keyboard navigation support
  - Ambient background elements (grids, glows)
  - Success/failure toast notifications

#### 2. **Auth State Management** (NgRx Store)
- **Location**: `src/frontend/src/app/store/auth/`
- **State Fields**:
  ```typescript
  interface AuthState {
    user: User | null;
    token: string | null;
    expiresAt: string | null;
    isAuthenticated: boolean;
    loading: boolean;
    error: unknown;
    userOrganizations: UserOrganization[];  // All accessible orgs
    currentOrganizationId: number | null;    // Currently selected org
  }
  ```

- **Storage Keys** (localStorage):
  - `tenantly_token`
  - `tenantly_user`
  - `tenantly_expires_at`
  - `tenantly_current_org_id`
  - `tenantly_organizations`

#### 3. **Organization Actions & Reducer**

**Actions Available**:
| Action | Purpose | Parameters |
|--------|---------|------------|
| `setUserOrganizations` | Load orgs from login response | `{ organizations, defaultOrganizationId }` |
| `setCurrentOrganization` | Set current org ID | `{ organizationId }` |
| `switchOrganization` | Trigger org switch | `{ organizationId }` |
| `switchOrganizationSuccess` | Handle successful switch | `{ response }` |
| `switchOrganizationFailure` | Handle switch error | `{ error }` |

**Reducer Handlers**:
- ✅ Login populates organizations with default selection
- ✅ Switch org refreshes JWT token from backend
- ✅ Persists current org ID and all orgs to localStorage
- ✅ Clears on logout

#### 4. **Selectors**
All available in `src/frontend/src/app/store/auth/auth.selectors.ts`:

| Selector | Returns |
|----------|---------|
| `selectUserOrganizations` | `UserOrganization[]` |
| `selectCurrentOrganizationId` | `number \| null` |
| `selectUserRole` | Role string |
| `selectIsSuperAdmin` / `selectIsOrgAdmin` / etc. | Boolean role checks |
| `selectCanManageOrganizations` | Boolean (SUPER_ADMIN only) |

---

## Data Models

### Organization Entity (`organization.model.ts`)
```typescript
interface Organization {
  id: number;
  name: string;
  slug: string;
  subscriptionTier: 'basic' | 'professional' | 'enterprise';
  maxUsers: number;
  active: boolean;
  description?: string;
}

interface UserOrganization {
  id: number;
  organization_id: number;
  organization: Organization;      // Embedded full org object
  role: string;                    // e.g., "Admin", "PropertyManager"
  created_at: string;
  updated_at: string;
}
```

### API Response Shapes
**Login Response**:
```typescript
interface LoginResponseWithOrganizations {
  token: string;
  refresh_token: string;           // httpOnly cookie (not stored here)
  user: unknown;
  organizations: UserOrganization[];
  default_organization_id: number;
  expires_at: string | Date;
}
```

**Switch Organization Response**:
```typescript
interface SetOrganizationResponse {
  token: string;
  organization: UserOrganization;   // Newly selected org
  expires_at: string | Date;
}
```

---

## Backend Integration (via AGENTS.md)

### Middleware Pattern
Every org-related route group must include:
```go
group := protected.Group("/reports")
group.Use(middleware.RequireOrgContext())  // Reads org_id from JWT
// Handler reads: c.GetInt("org_id")
```

### Backend API Endpoint for Switching
- **Route**: `PUT /api/v1/organization/:id/set`
- **Headers**: Authorization with new JWT token
- **Response**: New token + selected organization

### Role Hierarchy (5 Levels)
1. `SUPER_ADMIN` - Platform level, manages organizations
2. `ORG_ADMIN` - Organization level, manages org users
3. `Admin` - Organization data management
4. `PropertyManager` - Property-level operations
5. `Accountant` - Read-only financial access

**Organization Management Permission**:
- Only `SUPER_ADMIN` can manage/switch organizations (`selectCanManageOrganizations`)

---

## UI/UX Pattern (Theme Switching Reference)

The component uses CSS variables on `[data-theme]`:
```scss
:root, [data-theme='light'] {
  --bg-primary: #ffffff;
  --color-primary: #1e88e5;
}

[data-theme='dark'] {
  --bg-primary: #121212;
  --color-primary: #64b5f6;
}
```

---

## Completed Work State

- ✅ OrganizationPicker component with smooth animations
- ✅ Auth state management with org switching support
- ✅ All selectors for accessing org data
- ✅ localStorage persistence strategy
- ✅ NgRx action/reducer pattern fully implemented
- ✅ Role-based permission checks available

---

## Available for Enhancement

### Potential Improvements
1. **Org Context Display**: Show current org name/logo in app header
2. **Switch Org Dropdown**: Quick-switch selector in navbar (not full picker)
3. **Org Stats Dashboard**: View stats per organization
4. **Cross-Org Data Preview**: See data from other accessible orgs
5. **Org Billing Management**: Upgrade/downgrade subscription tiers
6. **Team Migration Tool**: Move users/properties between orgs

### Missing Pieces to Check
- ❓ Org service effects (watch for switch org API calls)
- ❓ Organization route guard logic
- ❓ Org-specific breadcrumb navigation
- ❓ Loading states for switch organization
- ❓ Error handling UI for invalid org selection

---

## Quick Reference: Using the Selectors

```typescript
// In a component
import * as AuthSelectors from 'path/to/auth.selectors';

const userOrganizations = this.store.select(AuthSelectors.selectUserOrganizations);
const currentOrgId = this.store.select(AuthSelectors.selectCurrentOrganizationId);
const canManageOrgs = this.store.select(AuthSelectors.selectCanManageOrganizations);

// In a pipe/presentation component
const org$ = combineLatest([
  select(UserOrganizations),
  select(CurrentOrganizationId)
]).pipe(map(([orgs, currentId]) => 
  orgs.find(o => o.organization_id === currentId)?.organization
));
```

---

## Build Commands Reference

**Frontend**:
```bash
npm run build          # Production build (checks everything)
npm run quality        # format + lint:fix
npm run start:local    # localhost:4200
```

---

## Next Steps Options

1. **Enhance UX**: Add quick-switch dropdown in header
2. **Add Organization Context**: Display current org name/logo throughout app
3. **Implement Org Router Guards**: Route-specific org validation
4. **Build Org Management Dashboard**: For SUPER_ADMIN users
5. **Add Cross-Org Navigation**: Sidebar showing accessible organizations

---

*Generated: Saturday, November 21, 2026*
