import { lazy, Suspense, useEffect } from 'react';
import { BrowserRouter, Routes, Route, useNavigate } from 'react-router-dom';
import { isGuestMode } from './api/client';
import ProtectedRoute from './components/ProtectedRoute';
import RoleRoute from './components/RoleRoute';
import CustomerLayout from './components/CustomerLayout';
import LoadingSpinner from './components/LoadingSpinner';
import { SiteProvider } from './components/comptoir/SiteSwitcher';
import './App.css';

// Route-level code splitting: each page is a separate chunk loaded on demand
const LoginPage = lazy(() => import('./pages/LoginPage'));
const RegisterPage = lazy(() => import('./pages/RegisterPage'));
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'));
const OAuthCallbackPage = lazy(() => import('./pages/OAuthCallbackPage'));
const HomePage = lazy(() => import('./pages/HomePage'));
const BakeriesPage = lazy(() => import('./pages/BakeriesPage'));
const BakeryDetailPage = lazy(() => import('./pages/BakeryDetailPage'));
const ScheduleOrdersPage = lazy(() => import('./pages/ScheduleOrdersPage'));
const OrderHistoryPage = lazy(() => import('./pages/OrderHistoryPage'));
const RecurringOrdersPage = lazy(() => import('./pages/RecurringOrdersPage'));
const AboutPage = lazy(() => import('./pages/AboutPage'));
const GuidePage = lazy(() => import('./pages/GuidePage'));
const BundlePage = lazy(() => import('./pages/BundlePage'));
const PrivacyPage = lazy(() => import('./pages/PrivacyPage'));
const TermsPage = lazy(() => import('./pages/TermsPage'));
const AccountSettingsPage = lazy(() => import('./pages/AccountSettingsPage'));

// Dashboard pages (seller portal)
const DashboardLayout = lazy(() => import('./pages/dashboard/DashboardLayout'));
const DashboardOverview = lazy(() => import('./pages/dashboard/DashboardOverview'));
const DashboardBakery = lazy(() => import('./pages/dashboard/DashboardBakery'));
const DashboardProducts = lazy(() => import('./pages/dashboard/DashboardProducts'));
const DashboardSchedule = lazy(() => import('./pages/dashboard/DashboardSchedule'));
const DashboardOrders = lazy(() => import('./pages/dashboard/DashboardOrders'));
const DashboardBundles = lazy(() => import('./pages/dashboard/DashboardBundles'));
const DashboardB2BPage = lazy(() => import('./pages/dashboard/DashboardB2BPage'));
const DashboardPayouts = lazy(() => import('./pages/dashboard/DashboardPayouts'));

// B2B Comptoir pages
const ComptoirLayout = lazy(() => import('./pages/comptoir/ComptoirLayout'));
const CommanderPage = lazy(() => import('./pages/comptoir/CommanderPage'));
const RecurrencesPage = lazy(() => import('./pages/comptoir/RecurrencesPage'));
const LivraisonsPage = lazy(() => import('./pages/comptoir/LivraisonsPage'));
const FacturesPage = lazy(() => import('./pages/comptoir/FacturesPage'));
const ComptoirProfilePage = lazy(() => import('./pages/comptoir/ComptoirProfilePage'));

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
      <Suspense fallback={<LoadingSpinner />}>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/auth/callback" element={<OAuthCallbackPage />} />

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
            <Route path="/settings" element={<ProtectedRoute><AccountSettingsPage /></ProtectedRoute>} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/guide" element={<GuidePage />} />
            <Route path="/privacy" element={<PrivacyPage />} />
            <Route path="/terms" element={<TermsPage />} />
          </Route>

          {/* Seller dashboard (role 0 or 1) */}
          <Route path="/dashboard" element={<RoleRoute allowedRoles={[0, 1]}><DashboardLayout /></RoleRoute>}>
            <Route index element={<DashboardOverview />} />
            <Route path="bakery" element={<DashboardBakery />} />
            <Route path="products" element={<DashboardProducts />} />
            <Route path="stats" element={<DashboardSchedule />} />
            <Route path="orders" element={<DashboardOrders />} />
            <Route path="bundles" element={<DashboardBundles />} />
            <Route path="payouts" element={<DashboardPayouts />} />
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
      </Suspense>
    </BrowserRouter>
  );
}

export default App;
