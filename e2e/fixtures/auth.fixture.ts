import { test as base, expect, Page } from '@playwright/test';
import { USERS } from '../helpers/test-data';
import { loginAsUser } from '../helpers/auth';

type AuthFixtures = {
  customerPage: Page;
  bakerPage: Page;
  adminPage: Page;
};

export const test = base.extend<AuthFixtures>({
  customerPage: async ({ browser, baseURL }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(baseURL!);
    await loginAsUser(page, USERS.customer);
    await use(page);
    await context.close();
  },

  bakerPage: async ({ browser, baseURL }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(baseURL!);
    await loginAsUser(page, USERS.baker);
    await use(page);
    await context.close();
  },

  adminPage: async ({ browser, baseURL }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(baseURL!);
    await loginAsUser(page, USERS.admin);
    await use(page);
    await context.close();
  },
});

export { expect };
