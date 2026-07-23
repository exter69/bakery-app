import { NavLink, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import { SiteSwitcher } from '../../components/comptoir/SiteSwitcher';
import { getProfile } from '../../api/b2b-client';
import { clearToken } from '../../api/client';
import './ComptoirNav.css';

export function ComptoirNav() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [companyName, setCompanyName] = useState('');
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    getProfile()
      .then((p) => setCompanyName(p.companyName))
      .catch(() => {});
  }, []);

  const handleLogout = () => {
    clearToken();
    window.location.href = '/login';
  };

  const NAV_TABS = [
    { to: '/comptoir', label: t('comptoir.nav.commander'), end: true },
    { to: '/comptoir/recurrences', label: t('comptoir.nav.recurrences'), end: false },
    { to: '/comptoir/livraisons', label: t('comptoir.nav.livraisons'), end: false },
    { to: '/comptoir/factures', label: t('comptoir.nav.factures'), end: false },
  ];

  return (
    <header className="comptoir-nav">
      <div className="comptoir-nav__inner">
        <div className="comptoir-nav__brand">
          <span className="comptoir-nav__brand-name">Mie & Beurre</span>
          <span className="comptoir-nav__brand-badge">Comptoir</span>
        </div>

        <nav className="comptoir-nav__tabs">
          {NAV_TABS.map((tab) => (
            <NavLink
              key={tab.to}
              to={tab.to}
              end={tab.end}
              className={({ isActive }) =>
                `comptoir-nav__tab ${isActive ? 'comptoir-nav__tab--active' : ''}`
              }
            >
              {tab.label}
            </NavLink>
          ))}
        </nav>

        <div className="comptoir-nav__right">
          <SiteSwitcher />
          <div className="comptoir-nav__account">
            <button
              type="button"
              className="comptoir-nav__account-btn"
              onClick={() => setMenuOpen(!menuOpen)}
              aria-expanded={menuOpen}
            >
              {companyName || '...'}
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="6 9 12 15 18 9" />
              </svg>
            </button>
            {menuOpen && (
              <div className="comptoir-nav__dropdown">
                <button
                  type="button"
                  className="comptoir-nav__dropdown-item"
                  onClick={() => { setMenuOpen(false); navigate('/comptoir/profile'); }}
                >
                  {t('comptoir.nav.profile')}
                </button>
                <button
                  type="button"
                  className="comptoir-nav__dropdown-item"
                  onClick={handleLogout}
                >
                  {t('nav.signOut')}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
