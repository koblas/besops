import { createContext, useCallback, useEffect, useState, type ReactNode } from 'react';
import { ConfigProvider, theme as antTheme } from 'antd';

export type ThemeMode = 'light' | 'dark' | 'auto';

interface ThemeContextValue {
  mode: ThemeMode;
  isDark: boolean;
  setMode: (mode: ThemeMode) => void;
}

export const ThemeContext = createContext<ThemeContextValue>({
  mode: 'auto',
  isDark: false,
  setMode: () => {},
});

const STORAGE_KEY = 'theme_mode';

function getSystemDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(() => {
    return (localStorage.getItem(STORAGE_KEY) as ThemeMode) || 'auto';
  });
  const [systemDark, setSystemDark] = useState(getSystemDark);

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = (e: MediaQueryListEvent) => setSystemDark(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  const setMode = useCallback((m: ThemeMode) => {
    setModeState(m);
    localStorage.setItem(STORAGE_KEY, m);
  }, []);

  const isDark = mode === 'dark' || (mode === 'auto' && systemDark);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light');
  }, [isDark]);

  return (
    <ThemeContext.Provider value={{ mode, isDark, setMode }}>
      <ConfigProvider
        theme={{
          algorithm: isDark ? antTheme.darkAlgorithm : antTheme.defaultAlgorithm,
          token: {
            colorPrimary: '#34d399',
            colorSuccess: '#34d399',
            colorWarning: '#fbbf24',
            colorError: '#f87171',
            colorInfo: '#60a5fa',
            borderRadius: 10,
            borderRadiusLG: 14,
            borderRadiusSM: 6,
            fontFamily: '-apple-system, BlinkMacSystemFont, "SF Pro Text", "SF Pro Display", "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
            fontSize: 14,
            controlHeight: 36,
            wireframe: false,
          },
          components: {
            Layout: {
              headerBg: isDark ? 'rgba(20, 20, 22, 0.85)' : 'rgba(255, 255, 255, 0.85)',
              headerColor: isDark ? 'rgba(255, 255, 255, 0.88)' : 'rgba(0, 0, 0, 0.88)',
              headerHeight: 56,
              bodyBg: isDark ? '#111113' : '#f8f9fa',
              siderBg: isDark ? '#18181b' : '#ffffff',
            },
            Menu: {
              itemBorderRadius: 8,
              itemMarginInline: 4,
              activeBarBorderWidth: 0,
            },
            Card: {
              borderRadiusLG: 12,
              boxShadowTertiary: isDark
                ? '0 1px 3px rgba(0, 0, 0, 0.3)'
                : '0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.02)',
            },
            Button: {
              borderRadius: 8,
              controlHeight: 36,
              primaryShadow: 'none',
              defaultShadow: 'none',
              dangerShadow: 'none',
            },
            Input: {
              borderRadius: 8,
              controlHeight: 36,
            },
            Select: {
              borderRadius: 8,
              controlHeight: 36,
            },
            Table: {
              borderRadius: 12,
              headerBorderRadius: 12,
            },
          },
        }}
      >
        {children}
      </ConfigProvider>
    </ThemeContext.Provider>
  );
}
