import { useContext } from 'react';
import { Layout, Button, Space, Dropdown, Tooltip } from 'antd';
import {
  DashboardOutlined,
  FileTextOutlined,
  SettingOutlined,
  LogoutOutlined,
  ToolOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { ThemeContext } from '../contexts/ThemeContext';

const { Header, Content } = Layout;

export function AppLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();
  const { isDark } = useContext(ThemeContext);

  const currentPath = location.pathname;

  const navItems = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: 'Dashboard' },
    { key: '/manage-status-page', icon: <FileTextOutlined />, label: 'Status Pages' },
    { key: '/maintenance', icon: <ToolOutlined />, label: 'Maintenance' },
  ];

  function isActive(key: string) {
    if (key === '/dashboard') return currentPath.startsWith('/dashboard') || currentPath === '/add' || currentPath.startsWith('/edit/') || currentPath.startsWith('/clone/');
    if (key === '/manage-status-page') return currentPath.startsWith('/manage-status-page') || currentPath.startsWith('/add-status-page');
    if (key === '/maintenance') return currentPath.startsWith('/maintenance') || currentPath.startsWith('/add-maintenance');
    return false;
  }

  const userMenuItems = [
    { key: 'settings', icon: <SettingOutlined />, label: 'Settings' },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: 'Sign Out', danger: true },
  ];

  const activeBg = isDark ? 'rgba(52, 211, 153, 0.12)' : 'rgba(52, 211, 153, 0.08)';
  const activeColor = '#34d399';
  const inactiveColor = isDark ? 'rgba(255,255,255,0.55)' : 'rgba(0,0,0,0.5)';

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          padding: '0 20px',
          backdropFilter: 'blur(12px)',
          WebkitBackdropFilter: 'blur(12px)',
          borderBottom: isDark ? '1px solid rgba(255,255,255,0.06)' : '1px solid rgba(0,0,0,0.06)',
          position: 'sticky',
          top: 0,
          zIndex: 100,
        }}
      >
        {/* Logo */}
        <div
          onClick={() => navigate('/dashboard')}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginRight: 32,
            cursor: 'pointer',
            userSelect: 'none',
          }}
        >
          <div style={{
            width: 28,
            height: 28,
            borderRadius: 8,
            background: 'linear-gradient(135deg, #34d399 0%, #059669 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}>
            <DashboardOutlined style={{ color: '#fff', fontSize: 14 }} />
          </div>
          <span style={{
            fontSize: 15,
            fontWeight: 600,
            color: isDark ? 'rgba(255,255,255,0.88)' : 'rgba(0,0,0,0.88)',
            letterSpacing: '-0.02em',
          }}>
            Bes Ops
          </span>
        </div>

        {/* Nav items */}
        <Space size={4} style={{ flex: 1 }}>
          {navItems.map(item => {
            const active = isActive(item.key);
            return (
              <Tooltip key={item.key} title={item.label} placement="bottom" mouseEnterDelay={0.5}>
                <Button
                  type="text"
                  icon={item.icon}
                  onClick={() => navigate(item.key)}
                  style={{
                    color: active ? activeColor : inactiveColor,
                    background: active ? activeBg : 'transparent',
                    fontWeight: active ? 600 : 400,
                    borderRadius: 8,
                    height: 36,
                    padding: '0 14px',
                  }}
                >
                  {item.label}
                </Button>
              </Tooltip>
            );
          })}
        </Space>

        {/* User menu */}
        <Dropdown
          menu={{
            items: userMenuItems,
            onClick: ({ key }) => {
              if (key === 'settings') navigate('/settings');
              if (key === 'logout') logout();
            },
          }}
          placement="bottomRight"
        >
          <Button
            type="text"
            style={{
              color: inactiveColor,
              borderRadius: 8,
              height: 36,
              padding: '0 10px',
            }}
          >
            <Space size={6}>
              <UserOutlined />
              <span style={{ fontSize: 13 }}>Admin</span>
            </Space>
          </Button>
        </Dropdown>
      </Header>
      <Content>
        <Outlet />
      </Content>
    </Layout>
  );
}
