import { useContext } from 'react';
import { AuthContext } from './AuthProvider';
import type { AuthState } from './AuthProvider';

/**
 * Access the reactive auth context.
 * Must be used within an AuthProvider.
 */
export function useAuth(): AuthState {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
