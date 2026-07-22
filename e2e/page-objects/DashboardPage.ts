import { Page, Locator } from '@playwright/test';

export class DashboardPage {
  readonly page: Page;
  readonly title: Locator;
  readonly statCards: Locator;
  readonly productsLink: Locator;
  readonly ordersLink: Locator;

  constructor(page: Page) {
    this.page = page;
    this.title = page.locator('h1').first();
    this.statCards = page.locator('.stat-card, .dash-stat-card');
    this.productsLink = page.getByText(/products/i);
    this.ordersLink = page.getByText(/orders/i);
  }

  async goto() {
    await this.page.goto('/dashboard');
  }

  async navigateToProducts() {
    await this.productsLink.first().click();
  }

  async navigateToOrders() {
    await this.ordersLink.first().click();
  }
}
