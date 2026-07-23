import { useState, useEffect } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { clearToken } from '../../api/client';
import { fetchMyBakery } from '../../api/seller';
import { useI18n } from '../../i18n';
import { ThemeSwitcher } from '../../components/ThemeSwitcher';
import type { Bakery } from '../../types/bakery';
import './pro-theme.css';
import './DashboardLayout.css';

const STORAGE_KEY = 'dashboard_sidebar_collapsed';

export default function DashboardLayout() {
  const { t } = useI18n();
  const [collapsed, setCollapsed] = useState(() => {
    return localStorage.getItem(STORAGE_KEY) === 'true';
  });
  const [bakery, setBakery] = useState<Bakery | null>(null);
  const [orderCount, setOrderCount] = useState<number>(0);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, String(collapsed));
  }, [collapsed]);

  useEffect(() => {
    fetchMyBakery().then((b) => {
      if (b) setBakery(b);
    });
  }, []);

  // Fetch pending order count for the badge
  useEffect(() => {
    if (!bakery) return;
    import('../../api/seller').then(({ fetchBakeryOrders }) => {
      fetchBakeryOrders(bakery.id, { status: 'confirmed' }).then((res) => {
        setOrderCount(res.total);
      }).catch(() => { /* silently ignore */ });
    });
  }, [bakery]);

  const handleLogout = () => {
    clearToken();
    window.location.href = '/login';
  };

  /** Get bakery initials for avatar fallback */
  const getInitials = (name: string) => {
    return name.split(' ').map((w) => w[0]).slice(0, 2).join('').toUpperCase();
  };

  /** Determine if the bakery is currently open based on today's schedule */
  const isBakeryOpen = (): boolean => {
    if (!bakery?.schedule?.length) return false;
    const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
    const today = days[new Date().getDay()];
    const todaySchedule = bakery.schedule.find((s) => s.day.toLowerCase() === today);
    if (!todaySchedule?.isOpen) return false;
    const now = new Date();
    const currentMinutes = now.getHours() * 60 + now.getMinutes();
    const open = todaySchedule.openTime.hour * 60 + todaySchedule.openTime.minute;
    const close = todaySchedule.closeTime.hour * 60 + todaySchedule.closeTime.minute;
    return currentMinutes >= open && currentMinutes <= close;
  };

  const getOpenHours = (): string => {
    if (!bakery?.schedule?.length) return '';
    const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
    const today = days[new Date().getDay()];
    const todaySchedule = bakery.schedule.find((s) => s.day.toLowerCase() === today);
    if (!todaySchedule?.isOpen) return '';
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${pad(todaySchedule.openTime.hour)}:${pad(todaySchedule.openTime.minute)}–${pad(todaySchedule.closeTime.hour)}:${pad(todaySchedule.closeTime.minute)}`;
  };

  const NAV_ITEMS: { to: string; label: string; icon: React.ReactNode; end: boolean; badge?: number }[] = [
    { to: '/dashboard', label: 'Tableau de bord', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>, end: true },
    { to: '/dashboard/orders', label: 'Commandes', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16v-2"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>, end: false, badge: orderCount || undefined },
    { to: '/dashboard/products', label: 'Menu & stock', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 13.87A4 4 0 0 1 7.41 6.6a5.11 5.11 0 0 1 8.57-1.81A3.13 3.13 0 0 1 20 8.1a2.5 2.5 0 0 1 0 5"/><path d="M4 19.5h16"/><path d="M4 22h16"/></svg>, end: false },
    { to: '/dashboard/bundles', label: 'Paniers du soir', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>, end: false },
    { to: '/dashboard/payouts', label: 'Paiements', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>, end: false },
    { to: '/dashboard/schedule', label: 'Planning', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>, end: false },
    { to: '/dashboard/bakery', label: 'Boutique', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>, end: false },
    { to: '/dashboard/b2b', label: 'B2B Comptoir', icon: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/></svg>, end: false },
  ];

  const open = isBakeryOpen();
  const hours = getOpenHours();

  return (
    <div className="dashboard-layout pro-portal">
      <aside className={`dashboard-sidebar ${collapsed ? 'dashboard-sidebar--collapsed' : ''}`}>
        <div className="dashboard-sidebar__header">
          <h2 className="dashboard-sidebar__brand">
            {collapsed ? <span className="dashboard-sidebar__brand-abbr">VB</span> : (
              <span className="dashboard-sidebar__brand-name">Votre Boulangerie</span>
            )}
          </h2>
          <button
            type="button"
            className="dashboard-sidebar__toggle"
            onClick={() => setCollapsed((c) => !c)}
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {collapsed ? '»' : '«'}
          </button>
        </div>
        <nav className="dashboard-sidebar__nav">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `dashboard-sidebar__link ${isActive ? 'dashboard-sidebar__link--active' : ''}`
              }
              title={collapsed ? item.label : undefined}
            >
              <span className="dashboard-sidebar__icon">{item.icon}</span>
              <span className="dashboard-sidebar__label">{item.label}</span>
              {item.badge && (
                <span className="dashboard-sidebar__badge">{item.badge}</span>
              )}
            </NavLink>
          ))}
        </nav>
        <div className="dashboard-sidebar__footer">
          {bakery && (
            <div className="dashboard-sidebar__bakery-info">
              <div className="dashboard-sidebar__avatar">
                {bakery.photoUrl
                  ? <img src={bakery.photoUrl} alt={bakery.name} />
                  : getInitials(bakery.name)
                }
              </div>
              <div className="dashboard-sidebar__bakery-detail">
                <span className="dashboard-sidebar__bakery-name">{bakery.name}</span>
                <span className={`dashboard-sidebar__bakery-status ${open ? 'dashboard-sidebar__bakery-status--open' : ''}`}>
                  {open ? `Ouvert · ${hours}` : 'Fermé'}
                </span>
              </div>
            </div>
          )}
          <ThemeSwitcher />
          <button
            type="button"
            className="dashboard-sidebar__logout"
            onClick={handleLogout}
          >
            {collapsed ? '🚪' : t('nav.signOut')}
          </button>
        </div>
      </aside>
      <main className="dashboard-main">
        <Outlet />
      </main>
    </div>
  );
}
