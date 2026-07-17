import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { apiFetch, setToken, decodeTokenRole } from '../api/client';
import './RegisterPage.css';

interface LoginResponse {
  token: string;
  user: { id: string; username: string; role: number };
}

export default function RegisterPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [role, setRole] = useState<number>(2); // default: customer
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

    setLoading(true);
    try {
      // Register
      await apiFetch('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username: username.trim(), password, role }),
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
          <h1 className="register-card__title">Create Account</h1>
          <p className="register-card__subtitle">Join our bakery community</p>

          <form className="register-card__form" onSubmit={handleSubmit}>
            <div className="register-card__field">
              <label htmlFor="reg-username" className="register-card__label">
                Username
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
                Password
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
                Confirm Password
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
              <legend className="register-card__label">I am a...</legend>
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
                  <span className="register-card__role-text">Customer</span>
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
                  <span className="register-card__role-text">Bakery Owner</span>
                </label>
              </div>
            </fieldset>

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
              {loading ? 'Creating account...' : 'Create Account'}
            </button>
          </form>

          <p className="register-card__footer">
            Already have an account?{' '}
            <Link to="/login" className="register-card__link">
              Sign In
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
