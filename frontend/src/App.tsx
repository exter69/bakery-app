import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, useNavigate } from 'react-router-dom';
import { isGuestMode } from './api/client';
import ProtectedRoute from './components/ProtectedRoute';
import RoleRoute from './components/RoleRoute';
import CustomerLayout from './components/CustomerLayout';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import NotFoundPage from './pages/NotFoundPage';
import HomePage from './pages/HomePage';
import BakeriesPage from './pages/BakeriesPage';
import BakeryDetailPage from './pages/BakeryDetailPage';
import ScheduleOrdersPage from './pages/ScheduleOrdersPage';
import RecurringOrdersPage from './pages/RecurringOrdersPage';
import AboutPage from './pages/AboutPage';
import DashboardLayout from './pages/dashboard/DashboardLayout';
import DashboardOverview from './pages/dashboard/DashboardOverview';
import DashboardBakery from './pages/dashboard/DashboardBakery';
import DashboardProducts from './pages/dashboard/DashboardProducts';
import DashboardSchedule from './pages/dashboard/DashboardSchedule';
import DashboardOrders from './pages/dashboard/DashboardOrders';
import DashboardReservations from './pages/dashboard/DashboardReservations';
import './App.css';

/** Listens for auth:unauthorized events and redirects to login */
function AuthRedirectListener() {
  const navigate = useNavigate();

  useEffect(() => {
    const handler = () => {
      if (isGuestMode()) return;
      if (window.location.pathname === '/login') return;
      navigate('/login', { state: { from: window.location.pathname }, replace: true });
    };
    window.addEventListener('auth:unauthorized', handler);
    return () => window.removeEventListener('auth:unauthorized', handler);
  }, [navigate]);

  return null;
}

function App() {
  return (
    <BrowserRouter>
      <AuthRedirectListener />
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        {/* Customer routes wrapped in CustomerLayout (navbar + hero strip) */}
        <Route element={<CustomerLayout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/bakeries" element={<BakeriesPage />} />
          <Route path="/bakeries/:id" element={<BakeryDetailPage />} />
          <Route path="/schedule" element={<ProtectedRoute><ScheduleOrdersPage /></ProtectedRoute>} />
          <Route path="/recurring" element={<ProtectedRoute><RecurringOrdersPage /></ProtectedRoute>} />
          <Route path="/about" element={<AboutPage />} />
        </Route>

        {/* Seller dashboard (role 0 or 1) */}
        <Route path="/dashboard" element={<RoleRoute allowedRoles={[0, 1]}><DashboardLayout /></RoleRoute>}>
          <Route index element={<DashboardOverview />} />
          <Route path="bakery" element={<DashboardBakery />} />
          <Route path="products" element={<DashboardProducts />} />
          <Route path="schedule" element={<DashboardSchedule />} />
          <Route path="orders" element={<DashboardOrders />} />
          <Route path="reservations" element={<DashboardReservations />} />
        </Route>

        {/* Catch-all */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
