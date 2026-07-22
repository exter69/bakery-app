import { test, expect } from '@playwright/test';
import { LoginPage } from '../page-objects/LoginPage';
import { USERS, REGISTRATION_CODE } from '../helpers/test-data';

test.describe('Authentication & Registration', () => {
  test('registration form is displayed at /register', async ({ page }) => {
    await page.goto('/register');
    await expect(page.getByRole('heading', { name: /register/i })).toBeVisible();
    await expect(page.getByLabel(/username/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
  });

  test('registration with valid code DEMO1234 succeeds', async ({ page }) => {
    await page.goto('/register');
    await page.getByLabel(/username/i).fill('newbaker');
    await page.getByLabel(/password/i).fill('password123');
    // Select baker/seller role if role selector is present
    const roleSelect = page.locator('select, [name="role"]');
    if (await roleSelect.isVisible()) {
      await roleSelect.selectOption({ label: 'baker' });
    }
    await page.getByLabel(/code|registration/i).fill(REGISTRATION_CODE);
    await page.getByRole('button', { name: /register|sign up/i }).click();

    // Verify redirect after successful registration
    await expect(page).not.toHaveURL(/\/register/);
  });

  test('registration with invalid code shows error', async ({ page }) => {
    await page.goto('/register');
    await page.getByLabel(/username/i).fill('invaliduser');
    await page.getByLabel(/password/i).fill('password123');
    const codeInput = page.getByLabel(/code|registration/i);
    if (await codeInput.isVisible()) {
      await codeInput.fill('INVALIDCODE');
    }
    await page.getByRole('button', { name: /register|sign up/i }).click();

    await expect(page.locator('[role="alert"], .error, .register__error')).toBeVisible();
  });

  test('login with invalid credentials shows error', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('wronguser', 'wrongpassword');

    await expect(loginPage.errorMessage).toBeVisible();
  });

  test('login with empty fields shows validation error', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('', '');

    // Either browser validation or custom error message should be visible
    const hasError = await loginPage.errorMessage.isVisible().catch(() => false);
    const hasValidation = await page.locator(':invalid').first().isVisible().catch(() => false);
    expect(hasError || hasValidation).toBeTruthy();
  });

  test('login with valid credentials redirects to home or dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login(USERS.customer.username, USERS.customer.password);

    // Should navigate away from /login
    await expect(page).not.toHaveURL(/\/login/);
  });
});
