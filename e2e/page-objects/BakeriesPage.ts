import { Page, Locator } from '@playwright/test';

export class BakeriesPage {
  readonly page: Page;
  readonly bakeryCards: Locator;

  constructor(page: Page) {
    this.page = page;
    this.bakeryCards = page.locator('.bakery-card, [data-testid="bakery-card"], article');
  }

  async goto() {
    await this.page.goto('/bakeries');
  }

  async getBakeryCount() {
    return this.bakeryCards.count();
  }

  async clickBakery(name: string) {
    await this.page.getByText(name).first().click();
  }
}
