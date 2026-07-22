import { Page, Locator } from '@playwright/test';

export class DashboardProductsPage {
  readonly page: Page;
  readonly addButton: Locator;
  readonly productRows: Locator;
  readonly modal: Locator;
  readonly nameInput: Locator;
  readonly priceInput: Locator;
  readonly categoryInput: Locator;
  readonly saveButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.addButton = page.getByRole('button', { name: /add product/i });
    this.productRows = page.locator('.dash-table tbody tr');
    this.modal = page.locator('.dash-modal');
    this.nameInput = page.locator('#prod-name');
    this.priceInput = page.locator('#prod-price');
    this.categoryInput = page.locator('#prod-category');
    this.saveButton = page.locator('.dash-modal button[type="submit"]');
  }

  async goto() {
    await this.page.goto('/dashboard/products');
  }

  async getProductCount() {
    return this.productRows.count();
  }

  async addProduct(name: string, price: string, category: string) {
    await this.addButton.click();
    await this.modal.waitFor({ state: 'visible' });
    await this.nameInput.fill(name);
    await this.priceInput.fill(price);
    await this.categoryInput.fill(category);
    await this.saveButton.click();
    await this.modal.waitFor({ state: 'hidden' });
  }

  async hasProduct(name: string) {
    return this.page.getByText(name).isVisible();
  }

  async editProductPrice(name: string, newPrice: string) {
    const row = this.page.locator(`tr:has-text("${name}")`);
    await row.getByText('Edit').click();
    await this.modal.waitFor({ state: 'visible' });
    await this.priceInput.clear();
    await this.priceInput.fill(newPrice);
    await this.saveButton.click();
    await this.modal.waitFor({ state: 'hidden' });
  }

  async deleteProduct(name: string) {
    const row = this.page.locator(`tr:has-text("${name}")`);
    this.page.once('dialog', (dialog) => dialog.accept());
    await row.getByText('Delete').click();
  }
}
