import { AuthState } from '../../store/auth/auth.reducer';

export interface AppState {
  auth: AuthState;
  // Feature states will be added here as they are lazy loaded
}
