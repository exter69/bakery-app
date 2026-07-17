import { Navigate } from 'react-router-dom';
import { isAuthenticated, getToken, decodeTokenRole } from '../api/client';

interface RoleRouteProps {
  children: React.ReactNode;
  /** Allowed roles for this route */
  allowedRoles: number[];
  /** Where to redirect if role doesn't match (default: /) */
  fallback?: string;
}

/**
 * Protects routes by requiring authentication AND a specific role.
 * Redirects to /login if not authenticated, or to fallback if role doesn't match.
 */
export default function RoleRoute({ children, allowedRoles, fallback = '/' }: RoleRouteProps) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }

  const token = getToken();
  if (!token) {
    return <Navigate to="/login" replace />;
  }

  const role = decodeTokenRole(token);
  if (role === null || !allowedRoles.includes(role)) {
    return <Navigate to={fallback} replace />;
  }

  return <>{children}</>;
}
