import { Page, Locator } from '@playwright/test';

export class SchedulePage {
  readonly page: Page;
  readonly orderRows: Locator;
  readonly deleteButtons: Locator;
  readonly confirmDialog: Locator;

  constructor(page: Page) {
    this.page = page;
    this.orderRows = page.locator('.schedule-orders__table tbody tr, .schedule-entry');
    this.deleteButtons = page.locator('.schedule-orders__delete-btn');
    this.confirmDialog = page.locator('.schedule-orders__confirm-dialog, [role="dialog"]');
  }

  async goto() {
    await this.page.goto('/schedule');
  }

  async getOrderCount() {
    return this.orderRows.count();
  }

  async cancelOrder(index = 0) {
    await this.deleteButtons.nth(index).click();
  }

  async confirmCancellation() {
    await this.page.locator('.schedule-orders__confirm-delete').click();
  }
}
