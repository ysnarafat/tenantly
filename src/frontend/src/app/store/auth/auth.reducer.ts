import { createReducer, on } from '@ngrx/store';
import { User } from '../../core/services/auth.service';
import * as AuthActions from './auth.actions';

export interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  expiresAt: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: unknown;
}

// Initialize state from localStorage if available
const initializeFromStorage = (): AuthState => {
  const token = localStorage.getItem('tenantly_token');
  const refreshToken = localStorage.getItem('tenantly_refresh_token');
  const userStr = localStorage.getItem('tenantly_user');
  const expiresAt = localStorage.getItem('tenantly_expires_at');

  if (token && userStr && expiresAt) {
    const user = JSON.parse(userStr);
    const isExpired = new Date() >= new Date(expiresAt);

    if (!isExpired) {
      return {
        user,
        token,
        refreshToken,
        expiresAt,
        isAuthenticated: true,
        loading: false,
        error: null,
      };
    }
  }

  return {
    user: null,
    token: null,
    refreshToken: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error: null,
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
    refreshToken: response.refresh_token,
    expiresAt:
      typeof response.expires_at === 'string'
        ? response.expires_at
        : new Date(response.expires_at).toISOString(),
    isAuthenticated: true,
    loading: false,
    error: null,
  })),

  on(AuthActions.loginFailure, (state, { error }) => ({
    ...state,
    user: null,
    token: null,
    refreshToken: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error,
  })),

  // Logout
  on(AuthActions.logout, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),

  on(AuthActions.logoutSuccess, (state) => ({
    user: null,
    token: null,
    refreshToken: null,
    expiresAt: null,
    isAuthenticated: false,
    loading: false,
    error: null,
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
    refreshToken: response.refresh_token,
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
    refreshToken: null,
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
    const refreshToken = localStorage.getItem('tenantly_refresh_token');
    const userStr = localStorage.getItem('tenantly_user');
    const expiresAt = localStorage.getItem('tenantly_expires_at');

    if (token && userStr && expiresAt) {
      const user = JSON.parse(userStr);
      const isExpired = new Date() >= new Date(expiresAt);

      if (!isExpired) {
        return {
          ...state,
          user,
          token,
          refreshToken,
          expiresAt,
          isAuthenticated: true,
        };
      }
    }

    return state;
  })
);
