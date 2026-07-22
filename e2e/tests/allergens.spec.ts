import { test, expect } from '../fixtures/auth.fixture';
import { BakeryDetailPage } from '../page-objects/BakeryDetailPage';
import { BAKERIES } from '../helpers/test-data';

test.describe('Allergen Indicators', () => {
  test('hovering allergen indicator shows tooltip', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);

    // Wait for allergen indicators to be present
    await detail.allergenIndicators.first().waitFor({ state: 'visible' });

    await detail.hoverAllergenIcon(0);

    const tooltip = customerPage.locator('.allergen-indicator__tooltip').first();
    await expect(tooltip).toBeVisible();
  });

  test('clicking allergen icon opens detail modal', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);

    await detail.allergenIndicators.first().waitFor({ state: 'visible' });
    await detail.clickAllergenIcon(0);

    const modal = customerPage.locator('.allergen-detail-modal, .allergen-modal');
    await expect(modal).toBeVisible();
  });

  test('closing allergen modal removes it', async ({ customerPage }) => {
    const detail = new BakeryDetailPage(customerPage);
    await detail.goto(BAKERIES.bakery1.id);

    await detail.allergenIndicators.first().waitFor({ state: 'visible' });
    await detail.clickAllergenIcon(0);

    const modal = customerPage.locator('.allergen-detail-modal, .allergen-modal');
    await expect(modal).toBeVisible();

    // Close the modal
    const closeBtn = customerPage.locator('.allergen-detail-modal__close, .allergen-modal__close, [aria-label="Close"]');
    await closeBtn.click();

    await expect(modal).not.toBeVisible();
  });
});
