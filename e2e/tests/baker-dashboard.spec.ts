import { test, expect } from '../fixtures/auth.fixture';
import { DashboardPage } from '../page-objects/DashboardPage';
import { DashboardProductsPage } from '../page-objects/DashboardProductsPage';

test.describe('Baker Dashboard', () => {
  test('dashboard overview loads with page title', async ({ bakerPage }) => {
    const dashboard = new DashboardPage(bakerPage);
    await dashboard.goto();

    await expect(dashboard.title).toBeVisible();
  });

  test('product list is displayed at /dashboard/products', async ({ bakerPage }) => {
    const products = new DashboardProductsPage(bakerPage);
    await products.goto();

    const count = await products.getProductCount();
    expect(count).toBeGreaterThan(0);
  });

  test('adding a new product shows it in the list', async ({ bakerPage }) => {
    const products = new DashboardProductsPage(bakerPage);
    await products.goto();

    const initialCount = await products.getProductCount();
    await products.addProduct('Croissant Test', '2.50', 'Viennoiserie');

    const newCount = await products.getProductCount();
    expect(newCount).toBe(initialCount + 1);
    expect(await products.hasProduct('Croissant Test')).toBeTruthy();
  });

  test('editing product price reflects update', async ({ bakerPage }) => {
    const products = new DashboardProductsPage(bakerPage);
    await products.goto();

    // Add a product to edit
    await products.addProduct('Pain Edit', '3.00', 'Pain');
    await products.editProductPrice('Pain Edit', '4.50');

    // Verify the updated price is visible
    await expect(bakerPage.locator('tr:has-text("Pain Edit")')).toContainText('4.50');
  });

  test('deleting a product removes it from the list', async ({ bakerPage }) => {
    const products = new DashboardProductsPage(bakerPage);
    await products.goto();

    // Add a product to delete
    await products.addProduct('Pain Suppr', '1.50', 'Pain');
    expect(await products.hasProduct('Pain Suppr')).toBeTruthy();

    await products.deleteProduct('Pain Suppr');

    // Wait for the row to disappear
    await expect(bakerPage.locator('tr:has-text("Pain Suppr")')).not.toBeVisible();
  });
});
