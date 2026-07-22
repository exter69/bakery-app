import { useRef, useEffect, useState, useCallback } from 'react';
import { Link, NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { isAuthenticated, clearToken } from '../api/client';
import { useI18n } from '../i18n';
import Footer from './Footer';
import LanguageSwitcher from './LanguageSwitcher';
import './CustomerLayout.css';

export default function CustomerLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const authenticated = isAuthenticated();
  const { t } = useI18n();

  const linksRef = useRef<HTMLDivElement>(null);
  const [indicator, setIndicator] = useState<{ left: number; width: number } | null>(null);

  const updateIndicator = useCallback(() => {
    if (!linksRef.current) return;
    const active = linksRef.current.querySelector('.pill-nav__link--active') as HTMLElement | null;
    if (active) {
      const containerRect = linksRef.current.getBoundingClientRect();
      const activeRect = active.getBoundingClientRect();
      setIndicator({
        left: activeRect.left - containerRect.left,
        width: activeRect.width,
      });
    } else {
      setIndicator(null);
    }
  }, []);

  useEffect(() => {
    // Small delay to let NavLink update its active class
    const timer = setTimeout(updateIndicator, 20);
    return () => clearTimeout(timer);
  }, [location.pathname, updateIndicator]);

  function handleSignOut() {
    clearToken();
    navigate('/login');
  }

  return (
    <div className="customer-layout">
      {/* Floating pill navbar */}
      <nav className="pill-nav">
        <Link to="/" className="pill-nav__brand">
          Mie &amp; Beurre
        </Link>

        <div className="pill-nav__spacer" />

        <div className="pill-nav__links" ref={linksRef}>
          {/* Sliding indicator */}
          {indicator && (
            <span
              className="pill-nav__indicator"
              style={{ left: indicator.left, width: indicator.width }}
            />
          )}
          <NavLink
            to="/bakeries"
            className={({ isActive }) =>
              `pill-nav__link${isActive ? ' pill-nav__link--active' : ''}`
            }
          >
            {t('nav.bakeries')}
          </NavLink>
          {authenticated && (
            <NavLink
              to="/schedule"
              className={({ isActive }) =>
                `pill-nav__link${isActive ? ' pill-nav__link--active' : ''}`
              }
            >
              {t('nav.orders')}
            </NavLink>
          )}
          <NavLink
            to="/about"
            className={({ isActive }) =>
              `pill-nav__link${isActive ? ' pill-nav__link--active' : ''}`
            }
          >
            {t('nav.about')}
          </NavLink>
          <NavLink
            to="/guide"
            className={({ isActive }) =>
              `pill-nav__link${isActive ? ' pill-nav__link--active' : ''}`
            }
          >
            {t('nav.guide')}
          </NavLink>
        </div>

        <LanguageSwitcher />

        <div className="pill-nav__actions">
          {authenticated ? (
            <button
              onClick={handleSignOut}
              className="pill-nav__btn pill-nav__btn--accent"
            >
              {t('nav.signOut')}
            </button>
          ) : (
            <Link to="/login" className="pill-nav__btn pill-nav__btn--accent">
              {t('nav.signIn')}
            </Link>
          )}
        </div>
      </nav>

      {/* Page content */}
      <main className="customer-layout__content">
        <Outlet />
      </main>

      <Footer />
    </div>
  );
}
