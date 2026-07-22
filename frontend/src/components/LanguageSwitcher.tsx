import { useI18n } from '../i18n';
import type { Locale } from '../i18n';
import './LanguageSwitcher.css';

const locales: { code: Locale; label: string }[] = [
  { code: 'en', label: 'EN' },
  { code: 'fr', label: 'FR' },
  { code: 'nl', label: 'NL' },
];

export default function LanguageSwitcher() {
  const { locale, setLocale } = useI18n();

  return (
    <div className="lang-switcher" role="group" aria-label="Language">
      {locales.map(({ code, label }) => (
        <button
          key={code}
          type="button"
          className={`lang-switcher__btn${locale === code ? ' lang-switcher__btn--active' : ''}`}
          onClick={() => setLocale(code)}
          aria-pressed={locale === code}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
