import { Navigate } from 'react-router-dom';
import { Spin, Result, Button } from 'antd';
import { useAuth } from '../hooks/useAuth';
import { AuthenticatedShell } from './AuthenticatedShell';

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, needsSetup, error, retry } = useAuth();

  if (isLoading) {
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}
      >
        <Spin size="large" tip="Connecting to server..." />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Result
          status="error"
          title="Connection Failed"
          subTitle={error}
          extra={<Button type="primary" onClick={retry}>Try Again</Button>}
        />
      </div>
    );
  }

  if (needsSetup) {
    return <Navigate to="/setup" replace />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <AuthenticatedShell>{children}</AuthenticatedShell>;
}
