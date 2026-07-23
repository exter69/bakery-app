import { createContext, useState, useEffect, useCallback, useMemo } from 'react';
import type { ReactNode } from 'react';
import { getToken, setToken as storeToken, clearToken, decodeTokenRole, isAuthenticated } from '../api/client';
import { UserRole } from './roles';

export interface AuthState {
  isLoggedIn: boolean;
  role: UserRole | null;
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
}

export const AuthContext = createContext<AuthState | null>(null);

function resolveRole(token: string | null): UserRole | null {
  if (!token) return null;
  const roleNum = decodeTokenRole(token);
  if (roleNum === null) return null;
  // Validate it maps to a known UserRole value
  if (Object.values(UserRole).includes(roleNum as UserRole)) {
    return roleNum as UserRole;
  }
  return null;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState<string | null>(() => getToken());
  const [loggedIn, setLoggedIn] = useState<boolean>(() => isAuthenticated());
  const [role, setRole] = useState<UserRole | null>(() => resolveRole(getToken()));

  const refresh = useCallback(() => {
    const current = getToken();
    const authenticated = isAuthenticated();
    setTokenState(current);
    setLoggedIn(authenticated);
    setRole(authenticated ? resolveRole(current) : null);
  }, []);

  const login = useCallback((newToken: string) => {
    storeToken(newToken);
    setTokenState(newToken);
    setLoggedIn(true);
    setRole(resolveRole(newToken));
  }, []);

  const logout = useCallback(() => {
    clearToken();
    setTokenState(null);
    setLoggedIn(false);
    setRole(null);
  }, []);

  // Listen for storage events (cross-tab sync) and auth:unauthorized custom event
  useEffect(() => {
    const handleStorage = (e: StorageEvent) => {
      if (e.key === 'auth_token' || e.key === null) {
        refresh();
      }
    };

    const handleUnauthorized = () => {
      refresh();
    };

    window.addEventListener('storage', handleStorage);
    window.addEventListener('auth:unauthorized', handleUnauthorized);
    return () => {
      window.removeEventListener('storage', handleStorage);
      window.removeEventListener('auth:unauthorized', handleUnauthorized);
    };
  }, [refresh]);

  const value: AuthState = useMemo(
    () => ({ isLoggedIn: loggedIn, role, token, login, logout }),
    [loggedIn, role, token, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
