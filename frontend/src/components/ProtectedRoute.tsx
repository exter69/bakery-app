import { Navigate, useLocation } from 'react-router-dom';
import { isAuthenticated, isGuestMode } from '../api/client';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

/**
 * Wraps routes that require authentication or guest access.
 * Redirects to /login if the user has no token and is not in guest mode.
 */
export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation();

  if (!isAuthenticated() && !isGuestMode()) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }

  return <>{children}</>;
}
