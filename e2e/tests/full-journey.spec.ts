import { test, expect } from '../fixtures/auth.fixture';
import { BakeriesPage } from '../page-objects/BakeriesPage';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';
import { SchedulePage } from '../page-objects/SchedulePage';
import { DashboardOrdersPage } from '../page-objects/DashboardOrdersPage';
import { BAKERIES } from '../helpers/test-data';

test.describe('Full Delivery Journey — A to Z', () => {
  test('complete order lifecycle: place order, baker advances status, customer sees it', async ({
    customerPage,
    bakerPage,
  }) => {
    const bakeries = new BakeriesPage(customerPage);
    const detail = new BakeryDetailPage(customerPage);
    const schedule = new SchedulePage(customerPage);
    const bakerOrders = new DashboardOrdersPage(bakerPage);

    // Step 1: Browse bakeries
    await bakeries.goto();
    await expect(customerPage.getByText(BAKERIES.bakery1.name)).toBeVisible();

    // Step 2: Select a bakery
    await bakeries.clickBakery(BAKERIES.bakery1.name);
    await expect(customerPage).toHaveURL(/\/bakeries\/bakery-1/);

    // Step 3: Open order modal
    await detail.clickOrder();
    await expect(customerPage.locator('.psm')).toBeVisible();

    // Step 4: Select products
    await customerPage.locator('.psm__card').first().click();

    // Step 5: Choose delivery day
    await customerPage.locator('.psm__day-chip').first().click();

    // Step 6: Choose time slot
    await customerPage.locator('.psm__time-select').selectOption({ index: 1 });

    // Step 7: Submit order (stub gateway auto-confirms payment)
    await customerPage.locator('.psm__submit-btn').click();

    // Step 8: Verify order appears in customer schedule
    await schedule.goto();
    const orderCount = await schedule.getOrderCount();
    expect(orderCount).toBeGreaterThan(0);

    // Step 9: Baker sees the order in their dashboard
    await bakerOrders.goto();
    const bakerOrderCount = await bakerOrders.getOrderCount();
    expect(bakerOrderCount).toBeGreaterThan(0);

    // Step 10: Baker advances order status (confirmed -> preparing)
    const firstRow = bakerOrders.orderRows.first();
    const initialText = await firstRow.textContent();
    await bakerOrders.changeOrderStatus(0, 'confirmed');

    // Step 11: Verify the status has changed in the display
    await expect(firstRow).not.toHaveText(initialText!);
  });
});
