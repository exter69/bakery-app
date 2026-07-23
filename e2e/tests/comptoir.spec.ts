import { test, expect } from '../fixtures/auth.fixture';

test.describe('B2B Comptoir Portal', () => {
  test('B2B user can access the comptoir and sees the commander page', async ({ b2bPage }) => {
    await b2bPage.goto('/comptoir');
    // The CommanderPage should render — verify its heading is visible
    await expect(b2bPage.locator('h1')).toBeVisible();
  });

  test('comptoir navigation links are rendered', async ({ b2bPage }) => {
    await b2bPage.goto('/comptoir');
    // Nav should have links to main sections
    await expect(b2bPage.locator('nav')).toBeVisible();
    await expect(b2bPage.getByRole('link', { name: /livraisons/i })).toBeVisible();
    await expect(b2bPage.getByRole('link', { name: /factures/i })).toBeVisible();
  });

  test('livraisons page loads without error', async ({ b2bPage }) => {
    await b2bPage.goto('/comptoir/livraisons');
    await expect(b2bPage.locator('h1')).toBeVisible();
  });

  test('factures page loads without error', async ({ b2bPage }) => {
    await b2bPage.goto('/comptoir/factures');
    await expect(b2bPage.locator('h1')).toBeVisible();
  });

  test('non-B2B user is redirected away from comptoir', async ({ customerPage }) => {
    await customerPage.goto('/comptoir');
    // RoleRoute should redirect to login or show forbidden
    await expect(customerPage).not.toHaveURL(/\/comptoir$/);
  });
});
