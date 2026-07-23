import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import * as Sentry from '@sentry/react'
import { I18nProvider } from './i18n'
import { ThemeProvider } from './theme/ThemeContext'
import './index.css'
import App from './App.tsx'

const sentryDsn = import.meta.env.VITE_SENTRY_DSN;
if (sentryDsn) {
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

const AppWithProviders = () => (
  <StrictMode>
    <ThemeProvider>
      <I18nProvider>
        <App />
      </I18nProvider>
    </ThemeProvider>
  </StrictMode>
);

createRoot(document.getElementById('root')!).render(
  sentryDsn
    ? <Sentry.ErrorBoundary fallback={<p>An error occurred. Please refresh.</p>}><AppWithProviders /></Sentry.ErrorBoundary>
    : <AppWithProviders />
);
