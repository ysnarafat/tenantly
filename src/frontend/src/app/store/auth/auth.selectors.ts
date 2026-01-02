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

export const selectCanManageUsers = createSelector(selectUserRole, (role) => role === 'Admin');
