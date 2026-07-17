import { NavLink, Outlet } from 'react-router-dom';
import { clearToken } from '../../api/client';
import './DashboardLayout.css';

const NAV_ITEMS = [
  { to: '/dashboard', label: 'Overview', end: true },
  { to: '/dashboard/bakery', label: 'My Bakery', end: false },
  { to: '/dashboard/products', label: 'Products', end: false },
  { to: '/dashboard/schedule', label: 'Schedule', end: false },
  { to: '/dashboard/orders', label: 'Orders', end: false },
  { to: '/dashboard/reservations', label: 'Reservations', end: false },
];

export default function DashboardLayout() {
  const handleLogout = () => {
    clearToken();
    window.location.href = '/login';
  };

  return (
    <div className="dashboard-layout">
      <aside className="dashboard-sidebar">
        <div className="dashboard-sidebar__header">
          <h2 className="dashboard-sidebar__brand">Bakery Portal</h2>
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
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="dashboard-sidebar__footer">
          <button
            type="button"
            className="dashboard-sidebar__logout"
            onClick={handleLogout}
          >
            Sign Out
          </button>
        </div>
      </aside>
      <main className="dashboard-main">
        <Outlet />
      </main>
    </div>
  );
}
