import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { isAuthenticated, clearToken } from '../api/client';
import './CustomerLayout.css';

export default function CustomerLayout() {
  const navigate = useNavigate();
  const authenticated = isAuthenticated();

  function handleSignOut() {
    clearToken();
    navigate('/login');
  }

  return (
    <div className="customer-layout">
      {/* Navbar */}
      <nav className="navbar">
        <Link to="/" className="navbar__brand">
          BakeryApp
        </Link>

        <div className="navbar__links">
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `navbar__link${isActive ? ' navbar__link--active' : ''}`
            }
          >
            Home
          </NavLink>
          <NavLink
            to="/bakeries"
            className={({ isActive }) =>
              `navbar__link${isActive ? ' navbar__link--active' : ''}`
            }
          >
            Bakeries
          </NavLink>
          {authenticated && (
            <NavLink
              to="/recurring"
              className={({ isActive }) =>
                `navbar__link${isActive ? ' navbar__link--active' : ''}`
              }
            >
              Recurring
            </NavLink>
          )}
          <NavLink
            to="/about"
            className={({ isActive }) =>
              `navbar__link${isActive ? ' navbar__link--active' : ''}`
            }
          >
            About
          </NavLink>
        </div>

        <div className="navbar__actions">
          {authenticated ? (
            <>
              <Link to="/schedule" className="navbar__btn navbar__btn--ghost">
                My Orders
              </Link>
              <button
                onClick={handleSignOut}
                className="navbar__btn navbar__btn--primary"
              >
                Sign Out
              </button>
            </>
          ) : (
            <Link to="/login" className="navbar__btn navbar__btn--primary">
              Sign In
            </Link>
          )}
        </div>
      </nav>

      {/* Hero strip */}
      <div className="hero-strip">
        <img
          src="https://images.unsplash.com/photo-1517433670267-08bbd4be890f?w=1600"
          alt="Bakery background"
          className="hero-strip__img"
        />
        <div className="hero-strip__overlay" />
      </div>

      {/* Page content */}
      <main className="customer-layout__content">
        <Outlet />
      </main>
    </div>
  );
}
