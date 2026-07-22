import { Page, Locator } from '@playwright/test';

export class BakeryDetailPage {
  readonly page: Page;
  readonly bakeryName: Locator;
  readonly orderButton: Locator;
  readonly productCards: Locator;
  readonly productSelectionModal: Locator;
  readonly allergenIndicators: Locator;

  constructor(page: Page) {
    this.page = page;
    this.bakeryName = page.locator('.bakery-page__name').first();
    this.orderButton = page.locator('.bakery-page__order-btn');
    this.productCards = page.locator('.product-card, .psm__card');
    this.productSelectionModal = page.locator('.psm');
    this.allergenIndicators = page.locator('.allergen-indicator');
  }

  async goto(id: string) {
    await this.page.goto(`/bakeries/${id}`);
  }

  async clickOrder() {
    await this.orderButton.click();
  }

  async isModalOpen() {
    return this.productSelectionModal.isVisible();
  }

  async getProductNames() {
    return this.page.locator('.product-card__name, .psm__card-name').allTextContents();
  }

  async hoverAllergenIcon(index = 0) {
    await this.allergenIndicators.nth(index).hover();
  }

  async clickAllergenIcon(index = 0) {
    await this.allergenIndicators.nth(index).click();
  }

  async getAllergenTooltipText(index = 0) {
    await this.allergenIndicators.nth(index).hover();
    return this.page.locator('.allergen-indicator__tooltip').nth(index).textContent();
  }
}
