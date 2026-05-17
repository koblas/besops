import { test, expect } from '@playwright/test';
import { setupAuthenticatedSession, mockAuthAPIs, mockMonitor, mockTags } from './helpers';

test.describe('Tag persistence on edit form', () => {
  test('selected tag appears in assigned list after API roundtrip', async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAuthAPIs(page);

    let fetchCount = 0;
    const monitorWithNewTag = {
      ...mockMonitor,
      tags: [
        ...mockMonitor.tags,
        { tagId: 'tag-3', name: 'staging', color: '#87d068', value: '' },
      ],
    };

    await page.route('**/api/v1/monitors/mon-1', (route) => {
      fetchCount++;
      const data = fetchCount > 1 ? monitorWithNewTag : mockMonitor;
      return route.fulfill({ json: data });
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
    await page.route('**/api/v1/monitors/mon-1/tags', (route) => {
      if (route.request().method() === 'POST') {
        addTagCalled = true;
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await expect(tagsCard.getByText('production')).toBeVisible();
    await expect(tagsCard.getByText('critical')).toBeVisible();

    const combobox = tagsCard.getByRole('combobox');
    await combobox.click();
    await page.getByTitle('staging').click();

    expect(addTagCalled).toBe(true);

    // After the mutation succeeds and query refetches, the new tag should appear
    await expect(tagsCard.getByText('staging')).toBeVisible({ timeout: 5000 });
  });

  test('tag persists after adding and is visible when page is revisited', async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAuthAPIs(page);

    // Simulate: first visit has no tags, after add the server always includes it
    let tagAdded = false;

    await page.route('**/api/v1/monitors/mon-1', (route) => {
      const tags = tagAdded
        ? [{ tagId: 'tag-3', name: 'staging', color: '#87d068', value: '' }]
        : [];
      return route.fulfill({ json: { ...mockMonitor, tags } });
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
      if (route.request().method() === 'POST') {
        tagAdded = true;
        return route.fulfill({ status: 201 });
      }
      return route.fulfill({ status: 201 });
    });

    // First visit: no tags
    await page.goto('/edit/mon-1');
    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    await expect(tagsCard.getByText('No tags assigned')).toBeVisible();

    // Add a tag
    const combobox = tagsCard.getByRole('combobox');
    await combobox.click();
    await page.getByTitle('staging').click();

    // Tag should appear after refetch
    await expect(tagsCard.getByText('staging')).toBeVisible({ timeout: 5000 });

    // Navigate away and back — tag should still be there
    await page.goto('/dashboard');
    await page.goto('/edit/mon-1');
    await expect(tagsCard.getByText('staging')).toBeVisible({ timeout: 5000 });
  });

  test('add tag request sends correct tagId in body', async ({ page }) => {
    await setupAuthenticatedSession(page);
    await mockAuthAPIs(page);

    await page.route('**/api/v1/monitors/mon-1', (route) => {
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

    const requestPromise = page.waitForRequest((req) =>
      req.url().includes('/monitors/mon-1/tags') && req.method() === 'POST'
    );

    await page.route('**/api/v1/monitors/mon-1/tags', (route) => {
      return route.fulfill({ status: 201 });
    });

    await page.goto('/edit/mon-1');

    const tagsCard = page.locator('.ant-card', { hasText: 'Tags' });
    const combobox = tagsCard.getByRole('combobox');
    await combobox.click();
    await page.getByTitle('staging').click();

    const req = await requestPromise;
    const body = req.postDataJSON();

    // Verify correct tag ID is sent
    expect(body.tagId).toBe('tag-3');
  });
});
