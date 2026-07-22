import { Page, Browser, BrowserContext } from '@playwright/test';

export interface LoginCredentials {
  username: string;
  password: string;
}

interface AuthToken {
  token: string;
}

/**
 * Logs in via the backend API and injects the JWT token into localStorage.
 * Avoids slow UI-based login for test setup.
 */
export async function loginAsUser(
  page: Page,
  credentials: LoginCredentials
): Promise<void> {
  const response = await page.evaluate(
    async ({ username, password }) => {
      const res = await fetch('http://localhost:8080/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      if (!res.ok) {
        throw new Error(`Login failed: ${res.status} ${res.statusText}`);
      }
      return res.json();
    },
    { username: credentials.username, password: credentials.password }
  );

  const { token } = response as AuthToken;
  await page.evaluate((t) => {
    localStorage.setItem('auth_token', t);
  }, token);
}

/**
 * Creates a new browser context with auth token already set.
 * Used by fixtures to provide pre-authenticated pages.
 */
export async function createAuthenticatedContext(
  browser: Browser,
  credentials: LoginCredentials,
  baseURL: string
): Promise<BrowserContext> {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(baseURL);
  await loginAsUser(page, credentials);
  await page.close();
  return context;
}
