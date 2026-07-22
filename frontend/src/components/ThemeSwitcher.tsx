import { useTheme } from '../theme/ThemeContext';
import './ThemeSwitcher.css';

export function ThemeSwitcher() {
  const { theme, setTheme } = useTheme();

  return (
    <div className="theme-switcher" role="radiogroup" aria-label="Theme">
      <button
        type="button"
        className={`theme-switcher__btn${theme === 'light' ? ' theme-switcher__btn--active' : ''}`}
        onClick={() => setTheme('light')}
        aria-pressed={theme === 'light'}
        title="Light"
      >
        ☀️
      </button>
      <button
        type="button"
        className={`theme-switcher__btn${theme === 'system' ? ' theme-switcher__btn--active' : ''}`}
        onClick={() => setTheme('system')}
        aria-pressed={theme === 'system'}
        title="System"
      >
        💻
      </button>
      <button
        type="button"
        className={`theme-switcher__btn${theme === 'dark' ? ' theme-switcher__btn--active' : ''}`}
        onClick={() => setTheme('dark')}
        aria-pressed={theme === 'dark'}
        title="Dark"
      >
        🌙
      </button>
    </div>
  );
}
