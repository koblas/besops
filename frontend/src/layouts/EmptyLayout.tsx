import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';

export function EmptyLayout() {
  return (
    <Layout
      style={{
        minHeight: '100vh',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        background: '#f5f5f5',
      }}
    >
      <Outlet />
    </Layout>
  );
}
