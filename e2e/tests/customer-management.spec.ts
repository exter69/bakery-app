import { test, expect } from '../fixtures/auth.fixture';
import { SchedulePage } from '../page-objects/SchedulePage';

test.describe('Customer Order Management', () => {
  test('existing orders are displayed on schedule page', async ({ customerPage }) => {
    const schedule = new SchedulePage(customerPage);
    await schedule.goto();

    const count = await schedule.getOrderCount();
    expect(count).toBeGreaterThan(0);
  });

  test('cancel button shows confirmation', async ({ customerPage }) => {
    const schedule = new SchedulePage(customerPage);
    await schedule.goto();

    // Click the cancel/delete button on the first order
    await schedule.cancelOrder(0);

    // Confirmation dialog should appear
    await expect(schedule.confirmDialog).toBeVisible();
  });

  test('confirming cancellation removes order', async ({ customerPage }) => {
    const schedule = new SchedulePage(customerPage);
    await schedule.goto();

    const initialCount = await schedule.getOrderCount();

    // Cancel the first order
    await schedule.cancelOrder(0);
    await schedule.confirmCancellation();

    // Order count should decrease
    const finalCount = await schedule.getOrderCount();
    expect(finalCount).toBe(initialCount - 1);
  });
});
