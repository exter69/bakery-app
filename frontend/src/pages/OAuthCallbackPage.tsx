import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { apiFetch, setToken, decodeTokenRole } from '../api/client';
import { useI18n } from '../i18n';
import './OAuthCallbackPage.css';

interface OAuthCallbackResponse {
  token: string;
  user: { id: string; username: string; role: number };
}

export default function OAuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { t } = useI18n();

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    // Determine provider from state or default to google
    const provider = searchParams.get('provider') || 'google';

    if (!code) {
      setError('No authorization code received.');
      return;
    }

    const exchangeCode = async () => {
      try {
        const res = await apiFetch<OAuthCallbackResponse>(
          `/auth/oauth/${provider}/callback`,
          {
            method: 'POST',
            body: JSON.stringify({ code }),
          }
        );
        setToken(res.token);

        // Redirect based on role or state
        const role = decodeTokenRole(res.token);
        if (role === 1) {
          navigate('/dashboard', { replace: true });
        } else {
          const redirectTo = state && state !== 'oauth' ? state : '/';
          navigate(redirectTo, { replace: true });
        }
      } catch (err) {
        setError(
          err instanceof Error ? err.message : t('login.oauthError')
        );
      }
    };

    exchangeCode();
  }, [searchParams, navigate, t]);

  if (error) {
    return (
      <div className="oauth-callback">
        <div className="oauth-callback__card">
          <h1 className="oauth-callback__title">Authentication failed</h1>
          <p className="oauth-callback__error" role="alert">
            {error}
          </p>
          <Link to="/login" className="oauth-callback__link">
            Back to login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="oauth-callback">
      <div className="oauth-callback__card">
        <p className="oauth-callback__loading">Signing you in...</p>
      </div>
    </div>
  );
}
