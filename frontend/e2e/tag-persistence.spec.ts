import { test, expect } from '@playwright/test';
import { setupAuthenticatedSession, mockAuthAPIs, mockMonitor, mockTags } from './helpers';

test.describe('Tag persistence on save', () => {
  test('tag assignment is sent when form is saved', async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAuthAPIs(page);

    await page.route('**/api/v1/monitors/mon-1', (route) => {
      if (route.request().method() === 'PUT') {
        return route.fulfill({ status: 200, json: mockMonitor });
      }
      return route.fulfill({ json: { ...mockMonitor, tags: [] } });
    });

    await page.route('**/api/v1/monitors', (route) => {
      return route.fulfill({ json: [mockMonitor] });
    });

    await page.route('**/api/v1/tags', (route) => {
      return route.fulfill({ json: mockTags });
    });

    await page.route('**/api/v1/notifications', (route) => {
      return route.fulfill({ json: [] });
    });

    await page.route('**/api/v1/ws/events', (route) => {
      return route.abort();
    });

    let addTagCalled = false;
    let addedTagId = '';
    await page.route('**/api/v1/monitors/mon-1/tags', (route) => {
      if (route.request().method() === 'POST') {
        addTagCalled = true;
        addedTagId = route.request().postDataJSON().tagId;
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.goto('/edit/mon-1');

    // Select a tag — no API call yet
    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const combobox = tagsCard.getByRole('combobox');
    await combobox.click();
    await page.getByTitle('staging').click();

    expect(addTagCalled).toBe(false);

    // Save the form
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Wait for the add-tag API call
    await page.waitForResponse((resp) =>
      resp.url().includes('/monitors/mon-1/tags') && resp.request().method() === 'POST'
    );

    expect(addTagCalled).toBe(true);
    expect(addedTagId).toBe('tag-3');
  });

  test('new monitor form can assign tags on create', async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAuthAPIs(page);

    await page.route('**/api/v1/monitors', (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 201, json: { id: 'mon-new', ...mockMonitor } });
      }
      return route.fulfill({ json: [mockMonitor] });
    });

    await page.route('**/api/v1/monitors/mon-new', (route) => {
      return route.fulfill({ json: { id: 'mon-new', ...mockMonitor } });
    });

    await page.route('**/api/v1/tags', (route) => {
      return route.fulfill({ json: mockTags });
    });

    await page.route('**/api/v1/notifications', (route) => {
      return route.fulfill({ json: [] });
    });

    await page.route('**/api/v1/ws/events', (route) => {
      return route.abort();
    });

    await page.route('**/api/v1/heartbeats/**', (route) => {
      return route.fulfill({ json: [] });
    });

    let addedTagId = '';
    await page.route('**/api/v1/monitors/mon-new/tags', (route) => {
      if (route.request().method() === 'POST') {
        addedTagId = route.request().postDataJSON().tagId;
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.goto('/add');

    // Fill required fields
    await page.getByLabel('Name').fill('New Monitor');
    await page.getByLabel('URL').fill('https://example.com');

    // Select a tag
    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const combobox = tagsCard.getByRole('combobox');
    await combobox.click();
    await page.getByTitle('staging').click();

    // Submit
    await page.getByRole('button', { name: 'Create Monitor' }).click();

    // Wait for the monitor creation and then tag assignment
    await page.waitForResponse((resp) =>
      resp.url().includes('/monitors/mon-new/tags') && resp.request().method() === 'POST'
    );

    expect(addedTagId).toBe('tag-3');
  });

  test('removing a tag is sent on save', async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAuthAPIs(page);

    await page.route('**/api/v1/monitors/mon-1', (route) => {
      if (route.request().method() === 'PUT') {
        return route.fulfill({ status: 200, json: mockMonitor });
      }
      return route.fulfill({ json: mockMonitor });
    });

    await page.route('**/api/v1/monitors', (route) => {
      return route.fulfill({ json: [mockMonitor] });
    });

    await page.route('**/api/v1/tags', (route) => {
      return route.fulfill({ json: mockTags });
    });

    await page.route('**/api/v1/notifications', (route) => {
      return route.fulfill({ json: [] });
    });

    await page.route('**/api/v1/ws/events', (route) => {
      return route.abort();
    });

    await page.route('**/api/v1/monitors/mon-1/tags', (route) => {
      return route.fulfill({ status: 201 });
    });

    let deletedTagId = '';
    await page.route('**/api/v1/monitors/mon-1/tags/**', (route) => {
      if (route.request().method() === 'DELETE') {
        deletedTagId = route.request().url().split('/tags/')[1];
        return route.fulfill({ status: 204 });
      }
      return route.fulfill({ status: 204 });
    });

    await page.goto('/edit/mon-1');

    // Remove a tag
    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const productionTag = tagsCard.locator('.ant-tag', { hasText: 'production' });
    await productionTag.locator('.anticon-close').click();

    // Verify it's visually gone but no API call yet
    await expect(tagsCard.getByText('production')).not.toBeVisible();

    // Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    await page.waitForResponse((resp) =>
      resp.url().includes('/monitors/mon-1/tags/') && resp.request().method() === 'DELETE'
    );

    expect(deletedTagId).toBe('tag-1');
  });
});
