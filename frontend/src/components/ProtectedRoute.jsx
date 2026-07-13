import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

// Protected Route Component. Pass `allowedRoles` to also enforce that the
// logged-in user's role matches the page - otherwise they're redirected to
// their own dashboard instead of landing on a page where every action 403s.
export const ProtectedRoute = ({ children, allowedRoles }) => {
  const { isAuthenticated, user } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (allowedRoles && !allowedRoles.includes(user?.role)) {
    const fallback = user?.role === 'hr_manager' ? '/dashboard/hr' : '/dashboard/employee';
    return <Navigate to={fallback} replace />;
  }

  return children;
};

export default ProtectedRoute;
