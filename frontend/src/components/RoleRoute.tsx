import { Navigate } from 'react-router-dom';
import { UserRole } from '../auth/roles';
import { useAuth } from '../auth/useAuth';

interface RoleRouteProps {
  children: React.ReactNode;
  /** Allowed roles for this route */
  allowedRoles: UserRole[];
  /** Where to redirect if role doesn't match (default: /) */
  fallback?: string;
}

/**
 * Protects routes by requiring authentication AND a specific role.
 * Redirects to /login if not authenticated, or to fallback if role doesn't match.
 */
export default function RoleRoute({ children, allowedRoles, fallback = '/' }: RoleRouteProps) {
  const { isLoggedIn, role } = useAuth();

  if (!isLoggedIn) {
    return <Navigate to="/login" replace />;
  }

  if (role === null || !allowedRoles.includes(role)) {
    return <Navigate to={fallback} replace />;
  }

  return <>{children}</>;
}
