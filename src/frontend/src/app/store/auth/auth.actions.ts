import { createAction, props } from '@ngrx/store';
import { LoginRequest, LoginResponse, User, ChangePasswordRequest, ResetPasswordRequest } from '../../core/services/auth.service';

// Login Actions
export const login = createAction(
    '[Auth] Login',
    props<{ credentials: LoginRequest }>()
);

export const loginSuccess = createAction(
    '[Auth] Login Success',
    props<{ response: LoginResponse }>()
);

export const loginFailure = createAction(
    '[Auth] Login Failure',
    props<{ error: any }>()
);

// Logout Actions
export const logout = createAction('[Auth] Logout');

export const logoutSuccess = createAction('[Auth] Logout Success');

export const logoutFailure = createAction(
    '[Auth] Logout Failure',
    props<{ error: any }>()
);

// Token Refresh Actions
export const refreshToken = createAction('[Auth] Refresh Token');

export const refreshTokenSuccess = createAction(
    '[Auth] Refresh Token Success',
    props<{ response: LoginResponse }>()
);

export const refreshTokenFailure = createAction(
    '[Auth] Refresh Token Failure',
    props<{ error: any }>()
);

// Password Management Actions
export const changePassword = createAction(
    '[Auth] Change Password',
    props<{ request: ChangePasswordRequest }>()
);

export const changePasswordSuccess = createAction(
    '[Auth] Change Password Success',
    props<{ message: string }>()
);

export const changePasswordFailure = createAction(
    '[Auth] Change Password Failure',
    props<{ error: any }>()
);

export const resetPassword = createAction(
    '[Auth] Reset Password',
    props<{ request: ResetPasswordRequest }>()
);

export const resetPasswordSuccess = createAction(
    '[Auth] Reset Password Success',
    props<{ message: string }>()
);

export const resetPasswordFailure = createAction(
    '[Auth] Reset Password Failure',
    props<{ error: any }>()
);

// Initialization Actions
export const initializeAuth = createAction('[Auth] Initialize');

export const clearError = createAction('[Auth] Clear Error');

export const setLoading = createAction(
    '[Auth] Set Loading',
    props<{ loading: boolean }>()
);