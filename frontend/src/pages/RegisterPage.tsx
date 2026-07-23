import { useState } from 'react';
import { useNavigate, Link, useSearchParams } from 'react-router-dom';
import { apiFetch, setToken, decodeTokenRole } from '../api/client';
import { useI18n } from '../i18n';
import BakerCard from '../components/BakerCard';
import './RegisterPage.css';

interface LoginResponse {
  token: string;
  user: { id: string; username: string; role: number };
}

export default function RegisterPage() {
  const [searchParams] = useSearchParams();
  const isBakeryMode = searchParams.get('role') === 'bakery';
  const { t } = useI18n();

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [role, setRole] = useState<number>(isBakeryMode ? 1 : 2);
  const [token, setBakeryToken] = useState('');
  const [acceptTerms, setAcceptTerms] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  function redirectByRole(token: string) {
    const tokenRole = decodeTokenRole(token);
    if (tokenRole === 1) {
      navigate('/dashboard', { replace: true });
    } else {
      navigate('/', { replace: true });
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!username.trim()) {
      setError('Username is required.');
      return;
    }
    if (password.length < 6) {
      setError('Password must be at least 6 characters.');
      return;
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    if (!acceptTerms) {
      setError(t('register.acceptTermsRequired'));
      return;
    }

    setLoading(true);
    try {
      // Register
      await apiFetch('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username: username.trim(), password, role, ...(role === 1 && token.trim() ? { code: token.trim() } : {}) }),
      });

      // Auto-login after registration
      const loginRes = await apiFetch<LoginResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username: username.trim(), password }),
      });

      setToken(loginRes.token);
      redirectByRole(loginRes.token);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="register-page">
      <div className="register-page__hero">
        <img
          src="https://images.unsplash.com/photo-1509440159596-0249088772ff?w=1200"
          alt="Fresh baked goods"
          className="register-page__hero-img"
        />
      </div>

      <div className="register-page__content">
        <div className="register-card">
          <h1 className="register-card__title">
            {role === 1 ? t('register.titleBakery') : t('register.title')}
          </h1>
          <p className="register-card__subtitle">
            {role === 1 ? t('register.subtitleBakery') : t('register.subtitle')}
          </p>

          <form className="register-card__form" onSubmit={handleSubmit}>
            <div className="register-card__field">
              <label htmlFor="reg-username" className="register-card__label">
                {t('login.username')}
              </label>
              <input
                id="reg-username"
                type="text"
                className="register-card__input"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Choose a username"
                autoComplete="username"
                autoFocus
              />
            </div>

            <div className="register-card__field">
              <label htmlFor="reg-password" className="register-card__label">
                {t('login.password')}
              </label>
              <input
                id="reg-password"
                type="password"
                className="register-card__input"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="At least 6 characters"
                autoComplete="new-password"
              />
            </div>

            <div className="register-card__field">
              <label htmlFor="reg-confirm" className="register-card__label">
                {t('register.confirmPassword')}
              </label>
              <input
                id="reg-confirm"
                type="password"
                className="register-card__input"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Re-enter your password"
                autoComplete="new-password"
              />
            </div>

            <fieldset className="register-card__role-group">
              <legend className="register-card__label">{t('register.iAmA')}</legend>
              <div className="register-card__role-options">
                <label className={`register-card__role-option ${role === 2 ? 'register-card__role-option--active' : ''}`}>
                  <input
                    type="radio"
                    name="role"
                    value={2}
                    checked={role === 2}
                    onChange={() => setRole(2)}
                    className="register-card__role-radio"
                  />
                  <span className="register-card__role-text">{t('register.customer')}</span>
                </label>
                <label className={`register-card__role-option ${role === 1 ? 'register-card__role-option--active' : ''}`}>
                  <input
                    type="radio"
                    name="role"
                    value={1}
                    checked={role === 1}
                    onChange={() => setRole(1)}
                    className="register-card__role-radio"
                  />
                  <span className="register-card__role-text">{t('register.bakeryOwner')}</span>
                </label>
              </div>
            </fieldset>

            {role === 1 && (
              <div className="register-card__field">
                <label htmlFor="reg-token" className="register-card__label">
                  {t('register.bakeryCode')}
                </label>
                <input
                  id="reg-token"
                  type="text"
                  className="register-card__input"
                  value={token}
                  onChange={(e) => setBakeryToken(e.target.value)}
                  placeholder="Enter the code you received"
                />
              </div>
            )}

            <div className="register-card__field register-card__terms">
              <label className="register-card__checkbox-label">
                <input
                  type="checkbox"
                  checked={acceptTerms}
                  onChange={(e) => setAcceptTerms(e.target.checked)}
                  className="register-card__checkbox"
                />
                <span>
                  {t('register.acceptTerms')}{' '}
                  <Link to="/terms" className="register-card__link" target="_blank">
                    {t('register.termsLink')}
                  </Link>
                  {' & '}
                  <Link to="/privacy" className="register-card__link" target="_blank">
                    {t('register.privacyLink')}
                  </Link>
                </span>
              </label>
            </div>

            {error && (
              <div className="register-card__error" role="alert">
                {error}
              </div>
            )}

            <button
              type="submit"
              className="register-card__submit"
              disabled={loading}
            >
              {loading ? t('register.creating') : t('register.createAccount')}
            </button>
          </form>

          <p className="register-card__footer">
            {t('register.haveAccount')}{' '}
            <Link to="/login" className="register-card__link">
              {t('login.signIn')}
            </Link>
          </p>
        </div>

        {/* Baker registration card */}
        <BakerCard />
      </div>
    </div>
  );
}
