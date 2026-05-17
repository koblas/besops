import { test, expect } from '@playwright/test';
import { setupAuthenticatedSession, mockAPIs, mockMonitor } from './helpers';

test.describe('Monitor List', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAPIs(page);
  });

  test('renders monitors in a table with name and type', async ({ page }) => {
    await page.goto('/list');

    await expect(page.getByText('All Monitors')).toBeVisible();
    await expect(page.getByText('My HTTP Monitor')).toBeVisible();
  });

  test('displays tags in table rows', async ({ page }) => {
    await page.goto('/list');

    await expect(page.getByText('production')).toBeVisible();
    await expect(page.getByText('critical')).toBeVisible();
  });

  test('search filters monitors by name', async ({ page }) => {
    const monitors = [
      mockMonitor,
      { ...mockMonitor, id: 'mon-2', name: 'DNS Monitor', type: 'dns', tags: [] },
    ];
    await page.route('**/api/v1/monitors', (route) => {
      if (route.request().method() === 'GET') {
        return route.fulfill({ json: monitors });
      }
      return route.fulfill({ status: 201, json: { id: 'mon-new' } });
    });

    await page.goto('/list');
    await expect(page.getByText('My HTTP Monitor')).toBeVisible();
    await expect(page.getByText('DNS Monitor')).toBeVisible();

    await page.getByPlaceholder('Search monitors...').fill('DNS');
    await expect(page.getByText('My HTTP Monitor')).not.toBeVisible();
    await expect(page.getByText('DNS Monitor')).toBeVisible();
  });

  test('navigates to monitor detail on name click', async ({ page }) => {
    await page.goto('/list');

    await page.getByRole('link', { name: 'My HTTP Monitor' }).or(page.getByText('My HTTP Monitor')).first().click();
    await expect(page).toHaveURL(/\/dashboard\/mon-1/);
  });

  test('pause button triggers pause API call', async ({ page }) => {
    let paused = false;
    await page.route('**/api/v1/monitors/mon-1/pause', (route) => {
      paused = true;
      return route.fulfill({ status: 204 });
    });

    await page.goto('/list');
    await page.locator('[aria-label="Pause"]').or(page.getByRole('button', { name: /pause/i })).first().click();
    expect(paused).toBe(true);
  });

  test('delete button shows confirmation modal', async ({ page }) => {
    await page.goto('/list');
    await page.locator('[aria-label="Delete"]').or(page.getByRole('button', { name: /delete/i })).first().click();

    await expect(page.locator('.ant-modal-confirm-title')).toHaveText(/Delete.*My HTTP Monitor/);
  });
});
