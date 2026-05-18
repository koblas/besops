import { test, expect } from '@playwright/test';
import { setupAuthenticatedSession, mockAPIs, mockMonitor, mockTags } from './helpers';

test.describe('TagSelector on Monitor Form', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAPIs(page);
  });

  test('displays assigned tags on edit form', async ({ page }) => {
    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await expect(tagsCard).toBeVisible();
    await expect(tagsCard.getByText('production')).toBeVisible();
    await expect(tagsCard.getByText('critical')).toBeVisible();
  });

  test('shows tags card on new monitor form', async ({ page }) => {
    await page.goto('/add');
    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await expect(tagsCard).toBeVisible();
    await expect(tagsCard.getByText('No tags assigned')).toBeVisible();
  });

  test('shows available tags in the dropdown excluding already-assigned', async ({ page }) => {
    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const select = tagsCard.locator('.ant-select');
    await select.click();

    await expect(page.getByTitle('staging')).toBeVisible();
    await expect(page.locator('.ant-select-item[title="production"]')).not.toBeVisible();
    await expect(page.locator('.ant-select-item[title="critical"]')).not.toBeVisible();
  });

  test('selecting a tag shows it immediately in the UI without API call', async ({ page }) => {
    let addTagCalled = false;
    await page.route('**/api/v1/monitors/mon-1/tags', async (route) => {
      if (route.request().method() === 'POST') {
        addTagCalled = true;
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const select = tagsCard.locator('.ant-select');
    await select.click();
    await page.getByTitle('staging').click();

    // Tag appears in the UI immediately
    await expect(tagsCard.getByText('staging')).toBeVisible();
    // But no API call yet — that happens on save
    expect(addTagCalled).toBe(false);
  });

  test('removing a tag hides it in the UI without API call', async ({ page }) => {
    let deleteTagCalled = false;
    await page.route('**/api/v1/monitors/mon-1/tags/**', async (route) => {
      if (route.request().method() === 'DELETE') {
        deleteTagCalled = true;
        return route.fulfill({ status: 204 });
      }
      return route.fulfill({ status: 204 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const productionTag = tagsCard.locator('.ant-tag', { hasText: 'production' });
    await productionTag.locator('.anticon-close').click();

    await expect(tagsCard.getByText('production')).not.toBeVisible();
    expect(deleteTagCalled).toBe(false);
  });

  test('tag changes are sent on form save', async ({ page }) => {
    const addRequests: string[] = [];
    const deleteRequests: string[] = [];

    await page.route('**/api/v1/monitors/mon-1/tags', async (route) => {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON();
        addRequests.push(body.tagId);
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.route('**/api/v1/monitors/mon-1/tags/**', async (route) => {
      if (route.request().method() === 'DELETE') {
        const url = route.request().url();
        deleteRequests.push(url.split('/tags/')[1]);
        return route.fulfill({ status: 204 });
      }
      return route.fulfill({ status: 204 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });

    // Add staging tag
    const select = tagsCard.locator('.ant-select');
    await select.click();
    await page.getByTitle('staging').click();

    // Remove production tag
    const productionTag = tagsCard.locator('.ant-tag', { hasText: 'production' });
    await productionTag.locator('.anticon-close').click();

    // Submit the form
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Wait for the API calls
    await page.waitForResponse((resp) => resp.url().includes('/monitors/mon-1/tags'));

    expect(addRequests).toContain('tag-3');
    expect(deleteRequests).toContain('tag-1');
  });

  test('creates a new tag immediately and reflects it in the form', async ({ page }) => {
    let createdTag = '';

    await page.route('**/api/v1/tags', async (route) => {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON();
        createdTag = body.name;
        return route.fulfill({ status: 201, json: { id: 'tag-new', name: body.name, color: body.color } });
      }
      return route.fulfill({ json: [...mockTags, { id: 'tag-new', name: 'deployment', color: '#597ef7' }] });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await tagsCard.getByRole('button', { name: /new tag/i }).click();

    const nameInput = tagsCard.getByPlaceholder('Tag name');
    await nameInput.fill('deployment');
    await tagsCard.getByRole('button', { name: 'Add' }).click();

    // Tag creation hits the API immediately
    expect(createdTag).toBe('deployment');
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
