import { test, expect } from '@playwright/test';
import { BakeriesPage } from '../page-objects/BakeriesPage';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';

test.describe('Guest Access', () => {
  test('guest can view bakeries page', async ({ page }) => {
    const bakeries = new BakeriesPage(page);
    await bakeries.goto();

    const count = await bakeries.getBakeryCount();
    expect(count).toBeGreaterThan(0);
  });

  test('guest can view bakery detail page', async ({ page }) => {
    const detail = new BakeryDetailPage(page);
    await detail.goto('bakery-1');

    await expect(detail.bakeryName).toBeVisible();
  });

  test('guest is redirected to /login when accessing /schedule', async ({ page }) => {
    await page.goto('/schedule');
    await expect(page).toHaveURL(/\/login/);
  });

  test('guest is redirected to /login when accessing /recurring', async ({ page }) => {
    await page.goto('/recurring');
    await expect(page).toHaveURL(/\/login/);
  });

  test('guest is redirected to /login when accessing /dashboard', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });
});
