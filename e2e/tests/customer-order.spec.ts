import { test, expect } from '../fixtures/auth.fixture';
import { BakeriesPage } from '../page-objects/BakeriesPage';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';
import { SchedulePage } from '../page-objects/SchedulePage';
import { BAKERIES, SEEDED_COUNTS } from '../helpers/test-data';

test.describe('Customer Order Flow', () => {
  test('bakery list displays all seeded bakeries', async ({ customerPage }) => {
    const bakeries = new BakeriesPage(customerPage);
    await bakeries.goto();

    const count = await bakeries.getBakeryCount();
    expect(count).toBe(SEEDED_COUNTS.bakeries);
  });

  test('clicking bakery navigates to detail page', async ({ customerPage }) => {
    const bakeries = new BakeriesPage(customerPage);
    await bakeries.goto();
    await bakeries.clickBakery(BAKERIES.bakery1.name);

    await expect(customerPage).toHaveURL(new RegExp(`/bakeries/${BAKERIES.bakery1.id}`));
  });

  test('clicking order button opens product selection modal', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);
    await detail.clickOrder();

    await expect(detail.productSelectionModal).toBeVisible();
  });

  test('can select products, choose day/time, and submit order', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);
    await detail.clickOrder();

    // Select a product
    await customerPage.locator('.psm__card').first().click();

    // Choose a day
    await customerPage.locator('.psm__day-chip').first().click();

    // Choose a time slot
    await customerPage.locator('.psm__time-select').selectOption({ index: 1 });

    // Submit the order
    await customerPage.locator('.psm__submit-btn').click();

    // Modal should close or confirmation shown
    await expect(detail.productSelectionModal).not.toBeVisible();
  });

  test('new order appears on schedule page', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);
    await detail.clickOrder();

    // Place an order
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
