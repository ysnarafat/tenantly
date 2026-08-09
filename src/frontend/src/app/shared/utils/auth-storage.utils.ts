/**
 * Backs the "Remember me" checkbox on login: when unchecked, session data
 * goes to sessionStorage instead of localStorage, so it disappears when the
 * browser closes rather than surviving a restart. The remember-me flag
 * itself always lives in localStorage — it has to outlive the session it's
 * describing so a fresh page load knows which backend to read from.
 */
const REMEMBER_ME_KEY = 'tenantly_remember_me';

function activeStorage(): Storage {
  return localStorage.getItem(REMEMBER_ME_KEY) === 'false' ? sessionStorage : localStorage;
}

export function setRememberMe(remember: boolean): void {
  localStorage.setItem(REMEMBER_ME_KEY, String(remember));
}

export const authStorage = {
  getItem(key: string): string | null {
    return activeStorage().getItem(key);
  },
  setItem(key: string, value: string): void {
    activeStorage().setItem(key, value);
  },
  removeItem(key: string): void {
    // Clear from both backends: if remember-me was toggled between logins,
    // a stale copy could otherwise be left sitting in the one not currently
    // active.
    localStorage.removeItem(key);
    sessionStorage.removeItem(key);
  },
};
