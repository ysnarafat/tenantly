import { createFeatureSelector, createSelector } from '@ngrx/store';
import { AuthState } from './auth.reducer';

export const selectAuthState = createFeatureSelector<AuthState>('auth');

export const selectUser = createSelector(selectAuthState, (state: AuthState) => state.user);

export const selectToken = createSelector(selectAuthState, (state: AuthState) => state.token);

export const selectRefreshToken = createSelector(
  selectAuthState,
  (state: AuthState) => state.refreshToken
);

export const selectIsAuthenticated = createSelector(
  selectAuthState,
  (state: AuthState) => state.isAuthenticated
);

export const selectAuthLoading = createSelector(
  selectAuthState,
  (state: AuthState) => state.loading
);

export const selectAuthError = createSelector(selectAuthState, (state: AuthState) => state.error);

export const selectUserRole = createSelector(selectUser, (user) => user?.role || '');

export const selectUserOrganizationId = createSelector(
  selectUser,
  (user) => user?.organization_id || null
);

export const selectIsSuperAdmin = createSelector(selectUserRole, (role) => role === 'SUPER_ADMIN');

export const selectIsOrgAdmin = createSelector(selectUserRole, (role) => role === 'ORG_ADMIN');

export const selectIsAdmin = createSelector(selectUserRole, (role) => role === 'Admin');

export const selectIsPropertyManager = createSelector(
  selectUserRole,
  (role) => role === 'PropertyManager'
);

export const selectIsAccountant = createSelector(selectUserRole, (role) => role === 'Accountant');

export const selectHasRole = (role: string) =>
  createSelector(selectUserRole, (userRole) => userRole === role);

export const selectHasAnyRole = (roles: string[]) =>
  createSelector(selectUserRole, (userRole) => roles.includes(userRole));

export const selectIsTokenExpired = createSelector(selectAuthState, (state: AuthState) => {
  if (!state.expiresAt) return true;
  return new Date() >= new Date(state.expiresAt);
});

export const selectCanViewLeases = createSelector(
  selectUserRole,
  (role) => role === 'Admin' || role === 'PropertyManager'
);

export const selectCanViewAttachments = createSelector(
  selectUserRole,
  (role) => role === 'Admin' || role === 'PropertyManager'
);

export const selectCanManageUsers = createSelector(
  selectUserRole,
  (role) => role === 'SUPER_ADMIN' || role === 'ORG_ADMIN' || role === 'Admin'
);

export const selectCanManageOrganizations = createSelector(
  selectUserRole,
  (role) => role === 'SUPER_ADMIN'
);

export const selectCanInviteUsers = createSelector(
  selectUserRole,
  (role) => role === 'SUPER_ADMIN' || role === 'ORG_ADMIN' || role === 'Admin'
);

export const selectUserOrganizations = createSelector(
  selectAuthState,
  (state: AuthState) => state.userOrganizations
);

export const selectCurrentOrganizationId = createSelector(
  selectAuthState,
  (state: AuthState) => state.currentOrganizationId
);
