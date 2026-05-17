import { test, expect } from '@playwright/test';
import { setupAuthenticatedSession, mockAPIs, mockMonitor } from './helpers';

test.describe('Monitor Form', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAPIs(page);
  });

  test('shows add form with default values', async ({ page }) => {
    await page.goto('/add');

    await expect(page.getByText('Add New Monitor')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create Monitor' })).toBeVisible();
  });

  test('shows edit form when editing existing monitor', async ({ page }) => {
    await page.goto('/edit/mon-1');

    await expect(page.getByText('Edit Monitor')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Save Changes' })).toBeVisible();
  });

  test('populates form fields from existing monitor data', async ({ page }) => {
    await page.goto('/edit/mon-1');

    const nameInput = page.locator('input#name');
    await expect(nameInput).toHaveValue('My HTTP Monitor');
  });

  test('shows connection card with type-specific fields for HTTP', async ({ page }) => {
    await page.goto('/edit/mon-1');

    const connectionCard = page.locator('.ant-card', { hasText: 'Connection' });
    await expect(connectionCard).toBeVisible();
    await expect(connectionCard.getByLabel('URL')).toBeVisible();
  });

  test('submits create request for new monitor', async ({ page }) => {
    let created = false;
    await page.route('**/api/v1/monitors', async (route) => {
      if (route.request().method() === 'POST') {
        created = true;
        return route.fulfill({ status: 201, json: { id: 'mon-new' } });
      }
      return route.fulfill({ json: [] });
    });

    await page.goto('/add');

    await page.locator('#name').fill('New Monitor');
    await page.locator('#url').fill('https://example.com');
    await page.getByRole('button', { name: 'Create Monitor' }).click();

    expect(created).toBe(true);
  });

  test('submits update request for existing monitor', async ({ page }) => {
    let updated = false;
    await page.route('**/api/v1/monitors/mon-1', async (route) => {
      if (route.request().method() === 'PUT') {
        updated = true;
        return route.fulfill({ status: 200, json: mockMonitor });
      }
      return route.fulfill({ json: mockMonitor });
    });

    await page.goto('/edit/mon-1');
    await page.getByRole('button', { name: 'Save Changes' }).click();

    expect(updated).toBe(true);
  });

  test('cancel button navigates back', async ({ page }) => {
    await page.goto('/edit/mon-1');
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page).toHaveURL(/\/dashboard\/mon-1/);
  });
});
