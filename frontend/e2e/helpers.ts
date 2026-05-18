import { type Page } from '@playwright/test';

export const mockMonitor = {
  id: 'mon-1',
  name: 'My HTTP Monitor',
  type: 'http',
  url: 'https://example.com',
  active: true,
  interval: 60,
  timeout: 48,
  maxRetries: 0,
  retryInterval: 60,
  maxRedirects: 10,
  method: 'GET',
  resendInterval: 0,
  packetSize: 56,
  tags: [
    { tagId: 'tag-1', name: 'production', color: '#f50', value: '' },
    { tagId: 'tag-2', name: 'critical', color: '#2db7f5', value: '' },
  ],
};

export const mockTags = [
  { id: 'tag-1', name: 'production', color: '#f50' },
  { id: 'tag-2', name: 'critical', color: '#2db7f5' },
  { id: 'tag-3', name: 'staging', color: '#87d068' },
];

export async function setupAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('access_token', 'test-token');
    localStorage.setItem('refresh_token', 'test-refresh');
  });
}

export async function mockAuthAPIs(page: Page) {
  await page.route('**/api/v1/auth/setup', (route) => {
    return route.fulfill({ json: { needSetup: false } });
  });

  await page.route('**/api/v1/auth/token/refresh', (route) => {
    return route.fulfill({ json: { token: 'test-token' } });
  });

  await page.route('**/api/v1/auth/logout', (route) => {
    return route.fulfill({ status: 204 });
  });
}

export async function mockAPIs(page: Page) {
  await mockAuthAPIs(page);

  await page.route('**/api/v1/monitors', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: [mockMonitor] });
    }
    return route.fulfill({ status: 201, json: { id: 'mon-new' } });
  });

  await page.route('**/api/v1/monitors/mon-1', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: mockMonitor });
    }
    return route.fulfill({ status: 200, json: mockMonitor });
  });

  await page.route('**/api/v1/tags', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({ json: mockTags });
    }
    return route.fulfill({ status: 201, json: { id: 'tag-new', name: 'new-tag', color: '#597ef7' } });
  });

  await page.route('**/api/v1/monitors/*/tags', (route) => {
    return route.fulfill({ status: 201 });
  });

  await page.route('**/api/v1/monitors/*/tags/**', (route) => {
    return route.fulfill({ status: 204 });
  });

  await page.route('**/api/v1/notifications', (route) => {
    return route.fulfill({ json: [] });
  });

  await page.route('**/api/v1/heartbeats/mon-1**', (route) => {
    return route.fulfill({ json: [] });
  });

  await page.route('**/api/v1/ws/events', (route) => {
    return route.abort();
  });
}
