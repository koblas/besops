import createClient, { type Middleware } from 'openapi-fetch';
import type { paths } from './generated/v1';

const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

export function getToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setTokens(accessToken: string, refreshToken?: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  if (refreshToken) {
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

let refreshPromise: Promise<boolean> | null = null;

async function doRefresh(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  const resp = await fetch('/api/v1/auth/token/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  });

  if (!resp.ok) {
    clearTokens();
    return false;
  }

  const data = await resp.json();
  if (data.token) {
    setTokens(data.token);
    return true;
  }
  clearTokens();
  return false;
}

async function refreshAccessToken(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => { refreshPromise = null; });
  }
  return refreshPromise;
}

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    const token = getToken();
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`);
    }
    return request;
  },
  async onResponse({ request, response }) {
    if (response.status === 401 && getRefreshToken()) {
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        const newToken = getToken();
        if (newToken) {
          request.headers.set('Authorization', `Bearer ${newToken}`);
          return fetch(request);
        }
      }
      window.location.href = '/login';
    }
    return response;
  },
};

export const api = createClient<paths>({ baseUrl: '/api/v1' });
api.use(authMiddleware);
