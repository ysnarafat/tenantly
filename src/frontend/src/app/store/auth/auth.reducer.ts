import { createReducer, on } from '@ngrx/store';
import { User } from '../../core/services/auth.service';
import { UserOrganization } from '../../core/models/organization.model';
import * as AuthActions from './auth.actions';

export interface AuthState {
  user: User | null;
  token: string | null;
  expiresAt: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: unknown;
  userOrganizations: UserOrganization[];
  currentOrganizationId: number | null;
}

// Initialize state from localStorage if available. The refresh token is never
// stored in localStorage — it lives only in an httpOnly cookie the backend
// sets and reads directly, so it's not readable (or storable) here.
const initializeFromStorage = (): AuthState => {
  const token = localStorage.getItem('tenantly_token');
  const userStr = localStorage.getItem('tenantly_user');
  const expiresAt = localStorage.getItem('tenantly_expires_at');

  if (token && userStr && expiresAt) {
    const user = JSON.parse(userStr);
    const isExpired = new Date() >= new Date(expiresAt);

    if (!isExpired) {
      const orgId = localStorage.getItem('tenantly_current_org_id');
      const orgsStr = localStorage.getItem('tenantly_organizations');
      const userOrganizations = orgsStr ? JSON.parse(orgsStr) : [];
      return {
        user,
        token,
        expiresAt,
        isAuthenticated: true,
        loading: false,
        error: null,
        userOrganizations,
        currentOrganizationId: orgId ? parseInt(orgId, 10) : null,
      };
    }
  }

  return {
    user: null,
    token: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error: null,
    userOrganizations: [],
    currentOrganizationId: null,
  };
};

export const initialState: AuthState = initializeFromStorage();

export const authReducer = createReducer(
  initialState,

  // Login
  on(AuthActions.login, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),

  on(AuthActions.loginSuccess, (state, { response }) => ({
    ...state,
    user: response.user,
    token: response.token,
    expiresAt:
      typeof response.expires_at === 'string'
        ? response.expires_at
        : new Date(response.expires_at).toISOString(),
    isAuthenticated: true,
    loading: false,
    error: null,
    userOrganizations: response.organizations || [],
    currentOrganizationId:
      response.default_organization_id ??
      (response.organizations?.length === 1 ? response.organizations[0].organization_id : null),
  })),

  on(AuthActions.loginFailure, (state, { error }) => ({
    ...state,
    user: null,
    token: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error,
  })),

  // Logout
  on(AuthActions.logout, () => ({
    user: null,
    token: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: true,
    error: null,
    userOrganizations: [],
    currentOrganizationId: null,
  })),

  on(AuthActions.logoutSuccess, () => ({
    user: null,
    token: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error: null,
    userOrganizations: [],
    currentOrganizationId: null,
  })),

  on(AuthActions.logoutFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Token Refresh
  on(AuthActions.refreshToken, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),

  on(AuthActions.refreshTokenSuccess, (state, { response }) => ({
    ...state,
    user: response.user,
    token: response.token,
    expiresAt:
      typeof response.expires_at === 'string'
        ? response.expires_at
        : new Date(response.expires_at).toISOString(),
    isAuthenticated: true,
    loading: false,
    error: null,
  })),

  on(AuthActions.refreshTokenFailure, (state, { error }) => ({
    ...state,
    token: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error,
  })),

  // Password Management
  on(AuthActions.changePassword, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),

  on(AuthActions.changePasswordSuccess, (state) => ({
    ...state,
    loading: false,
    error: null,
  })),

  on(AuthActions.changePasswordFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  on(AuthActions.resetPassword, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),

  on(AuthActions.resetPasswordSuccess, (state) => ({
    ...state,
    loading: false,
    error: null,
  })),

  on(AuthActions.resetPasswordFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Organization Actions
  on(AuthActions.setUserOrganizations, (state, { organizations, defaultOrganizationId }) => ({
    ...state,
    userOrganizations: organizations,
    currentOrganizationId: defaultOrganizationId ?? state.currentOrganizationId,
  })),

  on(AuthActions.setCurrentOrganization, (state, { organizationId }) => ({
    ...state,
    currentOrganizationId: organizationId,
  })),

  on(AuthActions.switchOrganization, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),

  on(AuthActions.switchOrganizationSuccess, (state, { response }) => ({
    ...state,
    token: response.token,
    expiresAt:
      typeof response.expires_at === 'string'
        ? response.expires_at
        : new Date(response.expires_at).toISOString(),
    currentOrganizationId: response.organization.organization_id,
    loading: false,
    error: null,
  })),

  on(AuthActions.switchOrganizationFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Utility Actions
  on(AuthActions.clearError, (state) => ({
    ...state,
    error: null,
  })),

  on(AuthActions.setLoading, (state, { loading }) => ({
    ...state,
    loading,
  })),

  // Initialize Auth (load from localStorage)
  on(AuthActions.initializeAuth, (state) => {
    const token = localStorage.getItem('tenantly_token');
    const userStr = localStorage.getItem('tenantly_user');
    const expiresAt = localStorage.getItem('tenantly_expires_at');

    if (token && userStr && expiresAt) {
      const user = JSON.parse(userStr);
      const isExpired = new Date() >= new Date(expiresAt);

      if (!isExpired) {
        const orgId = localStorage.getItem('tenantly_current_org_id');
        const orgsStr = localStorage.getItem('tenantly_organizations');
        const userOrganizations = orgsStr ? JSON.parse(orgsStr) : state.userOrganizations;
        return {
          ...state,
          user,
          token,
          expiresAt,
          isAuthenticated: true,
          userOrganizations,
          currentOrganizationId: orgId ? parseInt(orgId, 10) : state.currentOrganizationId,
        };
      }
    }

    return state;
  })
);
