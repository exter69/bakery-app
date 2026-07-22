import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { ThemeProvider, useTheme } from './ThemeContext';

// Helper component to read and control theme from tests
function ThemeConsumer() {
  const { theme, setTheme, resolvedTheme } = useTheme();
  return (
    <div>
      <span data-testid="theme">{theme}</span>
      <span data-testid="resolved">{resolvedTheme}</span>
      <button onClick={() => setTheme('dark')}>set-dark</button>
      <button onClick={() => setTheme('light')}>set-light</button>
      <button onClick={() => setTheme('system')}>set-system</button>
    </div>
  );
}

describe('ThemeProvider', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  afterEach(() => {
    document.documentElement.removeAttribute('data-theme');
  });

  it('defaults to system when localStorage is empty', () => {
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );
    expect(screen.getByTestId('theme').textContent).toBe('system');
  });

  it('reads persisted theme from localStorage', () => {
    localStorage.setItem('theme', 'dark');
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );
    expect(screen.getByTestId('theme').textContent).toBe('dark');
  });

  it('ignores invalid localStorage values and defaults to system', () => {
    localStorage.setItem('theme', 'invalid-value');
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );
    expect(screen.getByTestId('theme').textContent).toBe('system');
  });

  it('persists theme to localStorage when setTheme is called', async () => {
    const user = userEvent.setup();
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );

    await user.click(screen.getByText('set-dark'));
    expect(localStorage.getItem('theme')).toBe('dark');

    await user.click(screen.getByText('set-light'));
    expect(localStorage.getItem('theme')).toBe('light');
  });

  it('sets data-theme attribute on html element for explicit theme', async () => {
    const user = userEvent.setup();
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );

    await user.click(screen.getByText('set-dark'));
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');

    await user.click(screen.getByText('set-light'));
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('removes data-theme attribute when theme is set to system', async () => {
    const user = userEvent.setup();
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );

    await user.click(screen.getByText('set-dark'));
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');

    await user.click(screen.getByText('set-system'));
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('resolvedTheme reflects light when system prefers light', () => {
    // jsdom matchMedia defaults to not matching dark
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );
    expect(screen.getByTestId('resolved').textContent).toBe('light');
  });

  it('resolvedTheme reflects explicit dark when theme is dark', async () => {
    const user = userEvent.setup();
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    );

    await user.click(screen.getByText('set-dark'));
    expect(screen.getByTestId('resolved').textContent).toBe('dark');
  });
});
