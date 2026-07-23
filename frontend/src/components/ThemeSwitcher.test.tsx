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

  it('renders a single toggle button with system label by default', () => {
    renderSwitcher();
    const btn = screen.getByRole('button', { name: 'System theme' });
    expect(btn).toBeInTheDocument();
  });

  it('cycles from system to dark on click', async () => {
    const user = userEvent.setup();
    renderSwitcher();

    await user.click(screen.getByRole('button', { name: 'System theme' }));
    expect(screen.getByRole('button', { name: 'Dark mode' })).toBeInTheDocument();
    expect(localStorage.getItem('theme')).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('cycles from dark to light on click', async () => {
    localStorage.setItem('theme', 'dark');
    const user = userEvent.setup();
    renderSwitcher();

    await user.click(screen.getByRole('button', { name: 'Dark mode' }));
    expect(screen.getByRole('button', { name: 'Light mode' })).toBeInTheDocument();
    expect(localStorage.getItem('theme')).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('cycles from light back to system on click', async () => {
    localStorage.setItem('theme', 'light');
    const user = userEvent.setup();
    renderSwitcher();

    await user.click(screen.getByRole('button', { name: 'Light mode' }));
    expect(screen.getByRole('button', { name: 'System theme' })).toBeInTheDocument();
    expect(localStorage.getItem('theme')).toBe('system');
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('respects persisted dark preference from localStorage', () => {
    localStorage.setItem('theme', 'dark');
    renderSwitcher();
    expect(screen.getByRole('button', { name: 'Dark mode' })).toBeInTheDocument();
  });

  it('respects persisted light preference from localStorage', () => {
    localStorage.setItem('theme', 'light');
    renderSwitcher();
    expect(screen.getByRole('button', { name: 'Light mode' })).toBeInTheDocument();
  });
});
