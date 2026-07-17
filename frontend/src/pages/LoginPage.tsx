import { useState } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { apiFetch, setToken, decodeTokenRole, setGuestMode } from '../api/client';
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
      setError('Please enter username and password.');
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

  return (
    <div className="login-page">
      <div className="login-page__hero">
        <img
          src="https://images.unsplash.com/photo-1509440159596-0249088772ff?w=1200"
          alt="Fresh baked goods"
          className="login-page__hero-img"
        />
      </div>

      <div className="login-page__content">
        <div className="login-card">
          <h1 className="login-card__title">Welcome Back</h1>
          <p className="login-card__subtitle">Sign in to your account</p>

          <form className="login-card__form" onSubmit={handleSubmit}>
            <div className="login-card__field">
              <label htmlFor="username-input" className="login-card__label">
                Username
              </label>
              <input
                id="username-input"
                type="text"
                className="login-card__input"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter your username"
                autoComplete="username"
                autoFocus
              />
            </div>

            <div className="login-card__field">
              <label htmlFor="password-input" className="login-card__label">
                Password
              </label>
              <input
                id="password-input"
                type="password"
                className="login-card__input"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter your password"
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
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          <div className="login-card__divider">
            <span>or</span>
          </div>

          <div className="login-card__actions">
            <Link to="/register" className="login-card__link">
              Create Account
            </Link>
            <button
              type="button"
              className="login-card__guest-btn"
              onClick={handleGuestAccess}
            >
              Visit without account
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
