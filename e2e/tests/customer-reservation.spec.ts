import { test, expect } from '../fixtures/auth.fixture';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';
import { SchedulePage } from '../page-objects/SchedulePage';
import { BAKERIES } from '../helpers/test-data';

test.describe('Customer Reservation Flow', () => {
  test('can switch to pickup/reservation mode in modal', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);
    await detail.clickOrder();

    // Switch to reservation/pickup mode
    const reservationTab = customerPage.locator('.psm__mode-toggle, .psm__tab--reservation, [data-mode="reservation"]');
    await reservationTab.click();

    await expect(reservationTab).toBeVisible();
  });

  test('can submit reservation with products and pickup time', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);
    await detail.clickOrder();

    // Switch to reservation mode
    const reservationTab = customerPage.locator('.psm__mode-toggle, .psm__tab--reservation, [data-mode="reservation"]');
    await reservationTab.click();

    // Select a product
    await customerPage.locator('.psm__card').first().click();

    // Choose pickup day
    await customerPage.locator('.psm__day-chip').first().click();

    // Choose pickup time
    await customerPage.locator('.psm__time-select').selectOption({ index: 1 });

    // Submit reservation
    await customerPage.locator('.psm__submit-btn').click();

    // Modal should close
    await expect(detail.productSelectionModal).not.toBeVisible();
  });

  test('reservation appears on schedule page', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);
    await detail.clickOrder();

    // Switch to reservation mode and place reservation
    const reservationTab = customerPage.locator('.psm__mode-toggle, .psm__tab--reservation, [data-mode="reservation"]');
    await reservationTab.click();
    await customerPage.locator('.psm__card').first().click();
    await customerPage.locator('.psm__day-chip').first().click();
    await customerPage.locator('.psm__time-select').selectOption({ index: 1 });
    await customerPage.locator('.psm__submit-btn').click();

    // Navigate to schedule and verify
    const schedule = new SchedulePage(customerPage);
    await schedule.goto();
    const count = await schedule.getOrderCount();
    expect(count).toBeGreaterThan(0);
  });
});
