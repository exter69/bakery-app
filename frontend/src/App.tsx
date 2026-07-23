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
import OrderHistoryPage from './pages/OrderHistoryPage';
import RecurringOrdersPage from './pages/RecurringOrdersPage';
import AboutPage from './pages/AboutPage';
import GuidePage from './pages/GuidePage';
import DashboardLayout from './pages/dashboard/DashboardLayout';
import DashboardOverview from './pages/dashboard/DashboardOverview';
import DashboardBakery from './pages/dashboard/DashboardBakery';
import DashboardProducts from './pages/dashboard/DashboardProducts';
import DashboardSchedule from './pages/dashboard/DashboardSchedule';
import DashboardOrders from './pages/dashboard/DashboardOrders';
import DashboardBundles from './pages/dashboard/DashboardBundles';
import DashboardB2BPage from './pages/dashboard/DashboardB2BPage';
import BundlePage from './pages/BundlePage';
import ComptoirLayout from './pages/comptoir/ComptoirLayout';
import CommanderPage from './pages/comptoir/CommanderPage';
import RecurrencesPage from './pages/comptoir/RecurrencesPage';
import LivraisonsPage from './pages/comptoir/LivraisonsPage';
import FacturesPage from './pages/comptoir/FacturesPage';
import ComptoirProfilePage from './pages/comptoir/ComptoirProfilePage';
import { SiteProvider } from './components/comptoir/SiteSwitcher';
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

        {/* Home page renders its own nav (floating pill over hero) */}
        <Route path="/" element={<HomePage />} />

        {/* Customer routes wrapped in CustomerLayout (pill navbar at top) */}
        <Route element={<CustomerLayout />}>
          <Route path="/bakeries" element={<BakeriesPage />} />
          <Route path="/bakeries/:id" element={<BakeryDetailPage />} />
          <Route path="/paniers-du-soir" element={<BundlePage />} />
          <Route path="/schedule" element={<ProtectedRoute><ScheduleOrdersPage /></ProtectedRoute>} />
          <Route path="/history" element={<ProtectedRoute><OrderHistoryPage /></ProtectedRoute>} />
          <Route path="/recurring" element={<ProtectedRoute><RecurringOrdersPage /></ProtectedRoute>} />
          <Route path="/about" element={<AboutPage />} />
          <Route path="/guide" element={<GuidePage />} />
        </Route>

        {/* Seller dashboard (role 0 or 1) */}
        <Route path="/dashboard" element={<RoleRoute allowedRoles={[0, 1]}><DashboardLayout /></RoleRoute>}>
          <Route index element={<DashboardOverview />} />
          <Route path="bakery" element={<DashboardBakery />} />
          <Route path="products" element={<DashboardProducts />} />
          <Route path="stats" element={<DashboardSchedule />} />
          <Route path="orders" element={<DashboardOrders />} />
          <Route path="bundles" element={<DashboardBundles />} />
          <Route path="b2b" element={<DashboardB2BPage />} />
        </Route>

        {/* B2B Comptoir (role 3) */}
        <Route path="/comptoir" element={<RoleRoute allowedRoles={[3]}><SiteProvider><ComptoirLayout /></SiteProvider></RoleRoute>}>
          <Route index element={<CommanderPage />} />
          <Route path="recurrences" element={<RecurrencesPage />} />
          <Route path="livraisons" element={<LivraisonsPage />} />
          <Route path="factures" element={<FacturesPage />} />
          <Route path="profile" element={<ComptoirProfilePage />} />
        </Route>

        {/* Catch-all */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
