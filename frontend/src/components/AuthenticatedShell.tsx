import { useSSE } from '../hooks/useSSE';
import { FaviconBadge } from './FaviconBadge';

export function AuthenticatedShell({ children }: { children: React.ReactNode }) {
  useSSE();

  return (
    <>
      <FaviconBadge />
      {children}
    </>
  );
}
