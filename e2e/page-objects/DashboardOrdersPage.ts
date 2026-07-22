import { Page, Locator } from '@playwright/test';

export class DashboardOrdersPage {
  readonly page: Page;
  readonly orderRows: Locator;

  constructor(page: Page) {
    this.page = page;
    this.orderRows = page.locator('.dash-table tbody tr, .order-row');
  }

  async goto() {
    await this.page.goto('/dashboard/orders');
  }

  async getOrderCount() {
    return this.orderRows.count();
  }

  async changeOrderStatus(index: number, newStatus: string) {
    const row = this.orderRows.nth(index);
    const actionBtn = row.locator('.dash-btn--primary').first();
    await actionBtn.click();
  }
}
