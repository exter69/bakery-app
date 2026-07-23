import { useState, useEffect } from 'react';
import { useI18n } from '../i18n';
import { initSentry } from '../main';
import './CookieConsent.css';

const CONSENT_KEY = 'cookie_consent';

export type ConsentValue = 'all' | 'essential';

export function getConsentValue(): ConsentValue | null {
  const stored = localStorage.getItem(CONSENT_KEY);
  if (stored === 'all' || stored === 'essential') return stored;
  return null;
}

export default function CookieConsent() {
  const { t } = useI18n();
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!getConsentValue()) {
      setVisible(true);
    }
  }, []);

  function handleAccept(value: ConsentValue) {
    localStorage.setItem(CONSENT_KEY, value);
    setVisible(false);
    if (value === 'all') {
      initSentry();
    }
  }

  if (!visible) return null;

  return (
    <div className="cookie-consent" role="banner" aria-label={t('consent.title')}>
      <div className="cookie-consent__inner">
        <p className="cookie-consent__text">{t('consent.text')}</p>
        <div className="cookie-consent__actions">
          <button
            className="cookie-consent__btn cookie-consent__btn--essential"
            onClick={() => handleAccept('essential')}
          >
            {t('consent.essentialOnly')}
          </button>
          <button
            className="cookie-consent__btn cookie-consent__btn--accept"
            onClick={() => handleAccept('all')}
          >
            {t('consent.acceptAll')}
          </button>
        </div>
      </div>
    </div>
  );
}
