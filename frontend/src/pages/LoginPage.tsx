import { useState } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { apiFetch, setToken, decodeTokenRole, setGuestMode } from '../api/client';
import { useI18n } from '../i18n';
import BakerCard from '../components/BakerCard';
import LanguageSwitcher from '../components/LanguageSwitcher';
import GoogleIcon from '../components/icons/GoogleIcon';
import './LoginPage.css';

interface LoginResponse {
  token: string;
  user: { id: string; username: string; role: number };
}

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { t } = useI18n();

  const from = (location.state as { from?: string })?.from || '/';

  function redirectByRole(token: string) {
    const role = decodeTokenRole(token);
    if (role === 1) {
      navigate('/dashboard', { replace: true });
    } else {
      navigate(from === '/login' ? '/' : from, { replace: true });
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!username.trim() || !password.trim()) {
      setError(t('login.error'));
      return;
    }

    setLoading(true);
    try {
      const res = await apiFetch<LoginResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username: username.trim(), password }),
      });
      setToken(res.token);
      redirectByRole(res.token);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleGuestAccess = () => {
    setGuestMode();
    navigate('/', { replace: true });
  };

  const handleOAuthLogin = async (provider: string) => {
    setError(null);
    setLoading(true);
    try {
      const res = await apiFetch<{ url: string }>(`/auth/oauth/${provider}?state=${encodeURIComponent(from)}`);
      window.location.href = res.url;
    } catch (err) {
      setError(err instanceof Error ? err.message : t('login.oauthError'));
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      {/* Language switcher top-right */}
      <div className="login-page__lang">
        <LanguageSwitcher />
      </div>

      {/* Hero image top */}
      <div className="login-page__hero">
        <img
          src="https://images.unsplash.com/photo-1509440159596-0249088772ff?w=1200"
          alt="Fresh baked goods"
          className="login-page__hero-img"
        />
      </div>

      {/* Content: sign-in card + baker card side by side */}
      <div className="login-page__content">
        {/* Sign-in card */}
        <div className="login-card">
          <h1 className="login-card__title">{t('login.welcome')}</h1>
          <p className="login-card__subtitle">{t('login.subtitle')}</p>

          <form className="login-card__form" onSubmit={handleSubmit}>
            <div className="login-card__field">
              <label htmlFor="username-input" className="login-card__label">
                {t('login.username')}
              </label>
              <input
                id="username-input"
                type="text"
                className="login-card__input"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder={t('login.username')}
                autoComplete="username"
                autoFocus
              />
            </div>

            <div className="login-card__field">
              <label htmlFor="password-input" className="login-card__label">
                {t('login.password')}
              </label>
              <input
                id="password-input"
                type="password"
                className="login-card__input"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t('login.password')}
                autoComplete="current-password"
              />
            </div>

            {error && (
              <div className="login-card__error" role="alert">
                {error}
              </div>
            )}

            <button
              type="submit"
              className="login-card__submit"
              disabled={loading}
            >
              {loading ? t('login.signingIn') : t('login.signIn')}
            </button>
          </form>

          <div className="login-card__divider"><span>{t('login.or')}</span></div>

          <div className="login-card__social">
            <button
              type="button"
              className="login-card__social-btn login-card__social-btn--google"
              onClick={() => handleOAuthLogin('google')}
              disabled={loading}
            >
              <GoogleIcon className="login-card__social-icon" />
              <span>{t('login.signInWithGoogle')}</span>
            </button>
          </div>

          <div className="login-card__divider"><span>{t('login.oauthSeparator')}</span></div>

          <div className="login-card__actions">
            <Link to="/register" className="login-card__link">
              {t('login.createAccount')}
            </Link>
            <button
              type="button"
              className="login-card__guest-btn"
              onClick={handleGuestAccess}
            >
              {t('login.guest')}
            </button>
          </div>
        </div>

        <BakerCard />
      </div>
    </div>
  );
}
