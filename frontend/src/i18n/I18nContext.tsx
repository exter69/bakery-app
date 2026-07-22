import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import { translations, type Locale } from './translations';

interface I18nContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string) => string;
}

const I18nContext = createContext<I18nContextType>({
  locale: 'en',
  setLocale: () => {},
  t: (key) => key,
});

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const stored = localStorage.getItem('locale');
    if (stored === 'fr' || stored === 'nl' || stored === 'en') return stored;
    // Auto-detect from browser
    const lang = navigator.language.split('-')[0];
    if (lang === 'fr') return 'fr';
    if (lang === 'nl') return 'nl';
    return 'en';
  });

  const setLocale = useCallback((l: Locale) => {
    setLocaleState(l);
    localStorage.setItem('locale', l);
  }, []);

  const t = useCallback((key: string): string => {
    return translations[locale][key] ?? translations['en'][key] ?? key;
  }, [locale]);

  return (
    <I18nContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  return useContext(I18nContext);
}
