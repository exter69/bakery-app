import { test, expect } from '../fixtures/auth.fixture';
import { BakeriesPage } from '../page-objects/BakeriesPage';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';
import { BAKERIES } from '../helpers/test-data';

test.use({ viewport: { width: 375, height: 812 } });

test.describe('Responsive Layout', () => {
  test('home page renders without horizontal overflow on mobile', async ({ customerPage }) => {
    await customerPage.goto('/');

    const bodyWidth = await customerPage.evaluate(() => document.body.scrollWidth);
    const viewportWidth = await customerPage.evaluate(() => window.innerWidth);
    expect(bodyWidth).toBeLessThanOrEqual(viewportWidth);
  });

  test('bakery list displays in narrow layout on mobile', async ({ customerPage }) => {
    const bakeries = new BakeriesPage(customerPage);
    await bakeries.goto();

    const count = await bakeries.getBakeryCount();
    expect(count).toBeGreaterThan(0);

    // Verify cards fit within viewport width
    const firstCard = customerPage.locator('.baker-card').first();
    if (await firstCard.isVisible()) {
      const box = await firstCard.boundingBox();
      expect(box!.width).toBeLessThanOrEqual(375);
    }
  });

  test('full order flow completes on mobile viewport', async ({ customerPage }) => {
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
});
