import { test, expect } from '../fixtures/auth.fixture';
import { DashboardOrdersPage } from '../page-objects/DashboardOrdersPage';

test.describe('Baker Orders', () => {
  test('orders are displayed at /dashboard/orders', async ({ bakerPage }) => {
    const orders = new DashboardOrdersPage(bakerPage);
    await orders.goto();

    const count = await orders.getOrderCount();
    expect(count).toBeGreaterThan(0);
  });

  test('changing order status updates the display', async ({ bakerPage }) => {
    const orders = new DashboardOrdersPage(bakerPage);
    await orders.goto();

    // Get initial state of first order row
    const firstRow = orders.orderRows.first();
    const initialText = await firstRow.textContent();

    // Change the order status
    await orders.changeOrderStatus(0, 'confirmed');

    // Verify the row text has changed (status update reflected)
    await expect(firstRow).not.toHaveText(initialText!);
  });
});
