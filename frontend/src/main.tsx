import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import * as Sentry from '@sentry/react'
import { I18nProvider } from './i18n'
import { ThemeProvider } from './theme/ThemeContext'
import { getConsentValue } from './components/CookieConsent'
import './index.css'
import App from './App.tsx'

const sentryDsn = import.meta.env.VITE_SENTRY_DSN;

/**
 * Initialize Sentry only when the user has given "all" cookie consent.
 * Called on load (if consent was previously granted) and again after the user
 * interacts with the consent banner.
 */
export function initSentry() {
  if (!sentryDsn) return;
  if (Sentry.getClient()) return; // already initialized

  const consent = getConsentValue();
  if (consent !== 'all') return;

  Sentry.init({
    dsn: sentryDsn,
    environment: import.meta.env.VITE_APP_ENV || 'development',
    release: import.meta.env.VITE_APP_VERSION || 'dev',
    integrations: [
      Sentry.browserTracingIntegration(),
    ],
    tracesSampleRate: 0.1,
    // Don't send events in development builds
    enabled: import.meta.env.PROD,
    beforeSend(event) {
      // Scrub PII: remove user IP address
      if (event.user) {
        delete event.user.ip_address;
      }
      return event;
    },
  });
}

// Initialize Sentry on page load if consent was previously given
initSentry();

const AppWithProviders = () => (
  <StrictMode>
    <ThemeProvider>
      <I18nProvider>
        <App />
      </I18nProvider>
    </ThemeProvider>
  </StrictMode>
);

const isSentryActive = !!Sentry.getClient();

createRoot(document.getElementById('root')!).render(
  isSentryActive
    ? <Sentry.ErrorBoundary fallback={<p>An error occurred. Please refresh.</p>}><AppWithProviders /></Sentry.ErrorBoundary>
    : <AppWithProviders />
);
