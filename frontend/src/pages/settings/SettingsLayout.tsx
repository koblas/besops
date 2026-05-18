import { useContext } from 'react';
import { Layout, Menu } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  SettingOutlined,
  BgColorsOutlined,
  BellOutlined,
  SafetyOutlined,
  KeyOutlined,
  HistoryOutlined,
  TagsOutlined,
  GlobalOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { ThemeContext } from '../../contexts/ThemeContext';

const { Sider, Content } = Layout;

const menuItems = [
  { key: 'general', icon: <SettingOutlined />, label: 'General' },
  { key: 'appearance', icon: <BgColorsOutlined />, label: 'Appearance' },
  { key: 'notifications', icon: <BellOutlined />, label: 'Notifications' },
  { key: 'security', icon: <SafetyOutlined />, label: 'Security' },
  { key: 'api-keys', icon: <KeyOutlined />, label: 'API Keys' },
  { key: 'monitor-history', icon: <HistoryOutlined />, label: 'Monitor History' },
  { key: 'tags', icon: <TagsOutlined />, label: 'Tags' },
  { key: 'proxies', icon: <GlobalOutlined />, label: 'Proxies' },
  { key: 'about', icon: <InfoCircleOutlined />, label: 'About' },
];

export function SettingsLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { isDark } = useContext(ThemeContext);
  const currentKey = location.pathname.split('/settings/')[1] || 'general';

  return (
    <Layout style={{ height: 'calc(100vh - 56px)' }}>
      <Sider
        width={220}
        style={{
          background: isDark ? '#18181b' : '#ffffff',
          borderRight: isDark ? '1px solid rgba(255,255,255,0.06)' : '1px solid rgba(0,0,0,0.05)',
        }}
      >
        <Menu
          mode="inline"
          selectedKeys={[currentKey]}
          items={menuItems}
          onClick={({ key }) => navigate(`/settings/${key}`)}
          style={{ height: '100%', paddingTop: 12, border: 'none' }}
        />
      </Sider>
      <Content style={{ padding: 24, overflow: 'auto' }}>
        <Outlet />
      </Content>
    </Layout>
  );
}
