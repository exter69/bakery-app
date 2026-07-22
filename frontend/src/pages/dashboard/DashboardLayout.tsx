import { useState, useEffect } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { clearToken } from '../../api/client';
import { ThemeSwitcher } from '../../components/ThemeSwitcher';
import './DashboardLayout.css';

const STORAGE_KEY = 'dashboard_sidebar_collapsed';

const NAV_ITEMS = [
  { to: '/dashboard', label: 'Overview', icon: '📊', end: true },
  { to: '/dashboard/bakery', label: 'My Bakery', icon: '🏪', end: false },
  { to: '/dashboard/products', label: 'Products', icon: '🥐', end: false },
  { to: '/dashboard/schedule', label: 'Schedule', icon: '📅', end: false },
  { to: '/dashboard/orders', label: 'Orders', icon: '📦', end: false },
  { to: '/dashboard/reservations', label: 'Reservations', icon: '📋', end: false },
];

export default function DashboardLayout() {
  const [collapsed, setCollapsed] = useState(() => {
    return localStorage.getItem(STORAGE_KEY) === 'true';
  });

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, String(collapsed));
  }, [collapsed]);

  const handleLogout = () => {
    clearToken();
    window.location.href = '/login';
  };

  return (
    <div className="dashboard-layout">
      <aside className={`dashboard-sidebar ${collapsed ? 'dashboard-sidebar--collapsed' : ''}`}>
        <div className="dashboard-sidebar__header">
          <h2 className="dashboard-sidebar__brand">
            {collapsed ? '🍞' : 'Bakery Portal'}
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
            </NavLink>
          ))}
        </nav>
        <div className="dashboard-sidebar__footer">
          <ThemeSwitcher />
          <button
            type="button"
            className="dashboard-sidebar__logout"
            onClick={handleLogout}
          >
            {collapsed ? '🚪' : 'Sign Out'}
          </button>
        </div>
      </aside>
      <main className="dashboard-main">
        <Outlet />
      </main>
    </div>
  );
}
