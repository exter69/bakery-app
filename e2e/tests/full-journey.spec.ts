import { test, expect } from '../fixtures/auth.fixture';
import { BakeriesPage } from '../page-objects/BakeriesPage';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';
import { SchedulePage } from '../page-objects/SchedulePage';
import { BAKERIES } from '../helpers/test-data';

test.describe('Full Delivery Journey — A to Z', () => {
  test('complete order lifecycle', async ({ customerPage }) => {
    const bakeries = new BakeriesPage(customerPage);
    const detail = new BakeryDetailPage(customerPage);
    const schedule = new SchedulePage(customerPage);

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

    // Step 7: Submit order
    await customerPage.locator('.psm__submit-btn').click();

    // Step 8: Payment processing
    // TODO: Payment flow not yet implemented
    // Expected: redirect to payment gateway → complete payment → return to app
    // For now, the stub gateway auto-confirms

    // Step 9: Order confirmed
    // TODO: Verify confirmation screen/toast after payment

    // Step 10: Baker receives order notification
    // TODO: Real-time notifications not yet implemented

    // Step 11: Baker marks order as preparing
    // TODO: Test this in baker-orders.spec.ts (status: confirmed → preparing)

    // Step 12: Baker marks order as ready
    // TODO: Test this in baker-orders.spec.ts (status: preparing → ready)

    // Step 13: Delivery dispatch
    // TODO: Delivery/logistics system not yet implemented
    // Expected: assign delivery partner, provide tracking

    // Step 14: Customer receives delivery
    // TODO: Delivery tracking & confirmation not yet implemented
    // Expected: customer confirms receipt, order marked as delivered

    // Step 15: Order appears in history
    // Verify the order shows in the customer's schedule
    await schedule.goto();
    const count = await schedule.getOrderCount();
    expect(count).toBeGreaterThan(0);
  });
});
