import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';

export function PublicLayout() {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Layout.Content>
        <Outlet />
      </Layout.Content>
    </Layout>
  );
}
