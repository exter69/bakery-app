import { test, expect } from '@playwright/test';

test.describe('Internationalization', () => {
  test('switching language updates page content', async ({ page }) => {
    await page.goto('/');

    // Capture initial text content
    const initialContent = await page.locator('body').textContent();

    // Find and click language switcher
    const langSwitcher = page.locator('.lang-switcher, [data-testid="lang-switcher"], select[name="language"]');
    if (await langSwitcher.isVisible()) {
      // Try selecting a different language option
      const isSelect = await langSwitcher.evaluate((el) => el.tagName === 'SELECT');
      if (isSelect) {
        await langSwitcher.selectOption({ index: 1 });
      } else {
        await langSwitcher.click();
        await page.locator('.lang-option, [data-testid="lang-option"]').first().click();
      }
    } else {
      // Try finding language buttons (e.g., FR / EN)
      const langBtn = page.getByRole('button', { name: /^(FR|EN|NL)$/i }).first();
      await langBtn.click();
    }

    // Verify page content changed
    const updatedContent = await page.locator('body').textContent();
    expect(updatedContent).not.toBe(initialContent);
  });

  test('language selection persists across navigation', async ({ page }) => {
    await page.goto('/');

    // Switch language
    const langSwitcher = page.locator('.lang-switcher, [data-testid="lang-switcher"], select[name="language"]');
    if (await langSwitcher.isVisible()) {
      const isSelect = await langSwitcher.evaluate((el) => el.tagName === 'SELECT');
      if (isSelect) {
        await langSwitcher.selectOption({ index: 1 });
      } else {
        await langSwitcher.click();
        await page.locator('.lang-option, [data-testid="lang-option"]').first().click();
      }
    } else {
      const langBtn = page.getByRole('button', { name: /^(FR|EN|NL)$/i }).first();
      await langBtn.click();
    }

    // Capture content after switch
    const contentAfterSwitch = await page.locator('body').textContent();

    // Navigate to another page
    await page.goto('/bakeries');

    // Navigate back to home
    await page.goto('/');

    // Verify language persisted
    const contentAfterNav = await page.locator('body').textContent();
    expect(contentAfterNav).toBe(contentAfterSwitch);
  });
});
