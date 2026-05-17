import { createContext, useCallback, useEffect, useState, type ReactNode } from 'react';
import { api, getToken, getRefreshToken, setTokens, clearTokens } from '../api/client';

export interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  needsSetup: boolean;
  error: string | null;
}

export interface AuthContextValue extends AuthState {
  login: (
    username: string,
    password: string,
    twoFAToken?: string,
  ) => Promise<{ tokenRequired?: boolean }>;
  logout: () => Promise<void>;
  setup: (username: string, password: string) => Promise<void>;
  retry: () => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    token: getToken(),
    isAuthenticated: false,
    isLoading: true,
    needsSetup: false,
    error: null,
  });
  const [initAttempt, setInitAttempt] = useState(0);

  useEffect(() => {
    async function init() {
      setState(s => ({ ...s, isLoading: true, error: null }));
      try {
        const { data: setupData } = await api.GET('/auth/setup');
        if (setupData?.needSetup) {
          setState(s => ({ ...s, isLoading: false, needsSetup: true }));
          return;
        }

        const token = getToken();
        const refreshToken = getRefreshToken();
        if (!token || !refreshToken) {
          clearTokens();
          setState(s => ({ ...s, token: null, isLoading: false }));
          return;
        }

        // Validate the session by refreshing the access token
        const resp = await fetch('/api/v1/auth/token/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refreshToken }),
        });

        if (!resp.ok) {
          clearTokens();
          setState(s => ({ ...s, token: null, isLoading: false }));
        } else {
          const data = await resp.json();
          if (data.token) {
            setTokens(data.token);
            setState(s => ({
              ...s,
              token: data.token,
              isAuthenticated: true,
              isLoading: false,
            }));
          } else {
            clearTokens();
            setState(s => ({ ...s, token: null, isLoading: false }));
          }
        }
      } catch {
        setState(s => ({ ...s, isLoading: false, error: 'Unable to connect to the server. Check your network and try again.' }));
      }
    }
    init();
  }, [initAttempt]);

  const retry = useCallback(() => setInitAttempt(n => n + 1), []);

  const login = useCallback(async (username: string, password: string, twoFAToken?: string) => {
    const body: { username: string; password: string; token?: string } = { username, password };
    if (twoFAToken) body.token = twoFAToken;
    const { data, error } = await api.POST('/auth/login', { body });

    if (error) {
      const msg = (error as { error?: string }).error || 'Login failed';
      throw new Error(msg);
    }

    const resp = data as { token?: string; refreshToken?: string; tokenRequired?: boolean };
    if (resp.tokenRequired) {
      return { tokenRequired: true };
    }

    if (resp.token) {
      setTokens(resp.token, resp.refreshToken);
      setState(s => ({ ...s, token: resp.token!, isAuthenticated: true }));
    }
    return {};
  }, []);

  const logout = useCallback(async () => {
    await api.POST('/auth/logout').catch(() => {});
    clearTokens();
    setState(s => ({ ...s, token: null, isAuthenticated: false }));
  }, []);

  const setup = useCallback(async (username: string, password: string) => {
    const { error } = await api.POST('/auth/setup', {
      body: { username, password },
    });
    if (error) {
      throw new Error((error as { message?: string }).message || 'Setup failed');
    }

    // After setup, log in with the new credentials
    await login(username, password);
    setState(s => ({ ...s, needsSetup: false }));
  }, [login]);

  return (
    <AuthContext.Provider value={{ ...state, login, logout, setup, retry }}>
      {children}
    </AuthContext.Provider>
  );
}
