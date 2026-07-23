import { test, expect } from '../fixtures/auth.fixture';

test.describe('Pro Baker Portal (Votre Boulangerie)', () => {
  test('baker can access the dashboard overview', async ({ bakerPage }) => {
    await bakerPage.goto('/dashboard');
    // Dashboard layout should render with sidebar brand
    await expect(bakerPage.locator('.dashboard-sidebar')).toBeVisible();
    await expect(bakerPage.locator('.dashboard-main')).toBeVisible();
  });

  test('dashboard sidebar shows navigation links', async ({ bakerPage }) => {
    await bakerPage.goto('/dashboard');
    await expect(bakerPage.getByRole('link', { name: /Commandes/ })).toBeVisible();
    await expect(bakerPage.getByRole('link', { name: /Menu & stock/ })).toBeVisible();
    await expect(bakerPage.getByRole('link', { name: /Paniers du soir/ })).toBeVisible();
  });

  test('baker can navigate to products page', async ({ bakerPage }) => {
    await bakerPage.goto('/dashboard/products');
    await expect(bakerPage).toHaveURL(/\/dashboard\/products/);
    // Product page should load with content
    await expect(bakerPage.locator('.dashboard-main')).toBeVisible();
  });

  test('baker can navigate to bundles page', async ({ bakerPage }) => {
    await bakerPage.goto('/dashboard/bundles');
    await expect(bakerPage).toHaveURL(/\/dashboard\/bundles/);
    await expect(bakerPage.locator('.dashboard-main')).toBeVisible();
  });

  test('non-seller user is redirected away from dashboard', async ({ customerPage }) => {
    await customerPage.goto('/dashboard');
    // RoleRoute should redirect non-sellers
    await expect(customerPage).not.toHaveURL(/\/dashboard$/);
  });
});
