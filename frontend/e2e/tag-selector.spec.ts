import { test, expect } from '@playwright/test';
import { setupAuthenticatedSession, mockAPIs, mockMonitor, mockTags } from './helpers';

test.describe('TagSelector on Monitor Edit', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAPIs(page);
  });

  test('displays assigned tags with remove buttons', async ({ page }) => {
    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await expect(tagsCard).toBeVisible();
    await expect(tagsCard.getByText('production')).toBeVisible();
    await expect(tagsCard.getByText('critical')).toBeVisible();
  });

  test('shows available tags in the dropdown excluding already-assigned', async ({ page }) => {
    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const select = tagsCard.locator('.ant-select');
    await select.click();

    // 'staging' is available (not assigned), 'production' and 'critical' should not be
    await expect(page.getByTitle('staging')).toBeVisible();
    await expect(page.locator('.ant-select-item[title="production"]')).not.toBeVisible();
    await expect(page.locator('.ant-select-item[title="critical"]')).not.toBeVisible();
  });

  test('adds a tag by selecting from dropdown', async ({ page }) => {
    const addRequests: string[] = [];
    await page.route('**/api/v1/monitors/mon-1/tags', async (route) => {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON();
        addRequests.push(body.tagId);
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const select = tagsCard.locator('.ant-select');
    await select.click();
    await page.getByTitle('staging').click();

    expect(addRequests).toContain('tag-3');
  });

  test('removes a tag when close icon is clicked', async ({ page }) => {
    let deleteTagId = '';
    await page.route('**/api/v1/monitors/mon-1/tags/**', async (route) => {
      if (route.request().method() === 'DELETE') {
        const url = route.request().url();
        deleteTagId = url.split('/tags/')[1];
        return route.fulfill({ status: 204 });
      }
      return route.fulfill({ status: 204 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const productionTag = tagsCard.locator('.ant-tag', { hasText: 'production' });
    await productionTag.locator('.anticon-close').click();

    expect(deleteTagId).toBe('tag-1');
  });

  test('creates a new tag and assigns it', async ({ page }) => {
    let createdTag = '';

    await page.route('**/api/v1/tags', async (route) => {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON();
        createdTag = body.name;
        return route.fulfill({ status: 201, json: { id: 'tag-new', name: body.name, color: body.color } });
      }
      return route.fulfill({ json: mockTags });
    });

    const addTagPromise = page.waitForRequest((req) =>
      req.url().includes('/monitors/mon-1/tags') && req.method() === 'POST'
    );

    await page.route('**/api/v1/monitors/mon-1/tags', async (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await tagsCard.getByRole('button', { name: /new tag/i }).click();

    const nameInput = tagsCard.getByPlaceholder('Tag name');
    await nameInput.fill('deployment');
    await tagsCard.getByRole('button', { name: 'Add' }).click();

    const addReq = await addTagPromise;
    const assignedBody = addReq.postDataJSON();

    expect(createdTag).toBe('deployment');
    expect(assignedBody.tagId).toBe('tag-new');
  });

  test('tag section not shown on new monitor form', async ({ page }) => {
    await page.goto('/add');
    await expect(page.locator('.ant-card', { hasText: 'Tags' })).not.toBeVisible();
  });

  test('empty state shows helpful message when no tags assigned', async ({ page }) => {
    await page.route('**/api/v1/monitors/mon-1', (route) => {
      return route.fulfill({ json: { ...mockMonitor, tags: [] } });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await expect(tagsCard.getByText('No tags assigned')).toBeVisible();
  });
});
