import { useContext } from 'react';
import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';
import { MonitorListSidebar } from '../../components/MonitorListSidebar';
import { ThemeContext } from '../../contexts/ThemeContext';

const { Sider, Content } = Layout;

export function DashboardPage() {
  const { isDark } = useContext(ThemeContext);

  return (
    <Layout style={{ height: 'calc(100vh - 56px)' }}>
      <Sider
        width={280}
        style={{
          background: isDark ? '#18181b' : '#ffffff',
          borderRight: isDark ? '1px solid rgba(255,255,255,0.06)' : '1px solid rgba(0,0,0,0.05)',
          overflow: 'auto',
        }}
      >
        <MonitorListSidebar />
      </Sider>
      <Content style={{ padding: 24, overflow: 'auto' }}>
        <Outlet />
      </Content>
    </Layout>
  );
}
