import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, beforeEach } from 'vitest';
import { ThemeSwitcher } from './ThemeSwitcher';
import { ThemeProvider } from '../theme/ThemeContext';

function renderSwitcher() {
  return render(
    <ThemeProvider>
      <ThemeSwitcher />
    </ThemeProvider>
  );
}

describe('ThemeSwitcher', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('renders three theme buttons', () => {
    renderSwitcher();
    const group = screen.getByRole('radiogroup', { name: 'Theme' });
    const buttons = group.querySelectorAll('button');
    expect(buttons).toHaveLength(3);
  });

  it('marks system button as active by default', () => {
    renderSwitcher();
    const systemBtn = screen.getByTitle('System');
    expect(systemBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('clicking dark button activates dark mode', async () => {
    const user = userEvent.setup();
    renderSwitcher();

    const darkBtn = screen.getByTitle('Dark');
    await user.click(darkBtn);

    expect(darkBtn).toHaveAttribute('aria-pressed', 'true');
    expect(localStorage.getItem('theme')).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('clicking light button activates light mode', async () => {
    const user = userEvent.setup();
    renderSwitcher();

    const lightBtn = screen.getByTitle('Light');
    await user.click(lightBtn);

    expect(lightBtn).toHaveAttribute('aria-pressed', 'true');
    expect(localStorage.getItem('theme')).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('switching back to system removes data-theme attribute', async () => {
    const user = userEvent.setup();
    renderSwitcher();

    await user.click(screen.getByTitle('Dark'));
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');

    await user.click(screen.getByTitle('System'));
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
    expect(localStorage.getItem('theme')).toBe('system');
  });

  it('respects persisted dark preference from localStorage', () => {
    localStorage.setItem('theme', 'dark');
    renderSwitcher();
    const darkBtn = screen.getByTitle('Dark');
    expect(darkBtn).toHaveAttribute('aria-pressed', 'true');
  });
});
