import { Typography, Radio, Space } from 'antd';
import { useTheme } from '../../hooks/useTheme';
import type { ThemeMode } from '../../contexts/ThemeContext';

const { Title, Text } = Typography;

export function AppearanceSettings() {
  const { mode, setMode } = useTheme();

  return (
    <div>
      <Title level={4}>Appearance</Title>
      <Space direction="vertical" size="large">
        <div>
          <Text strong>Theme</Text>
          <br />
          <Radio.Group
            value={mode}
            onChange={e => setMode(e.target.value as ThemeMode)}
            style={{ marginTop: 8 }}
          >
            <Radio.Button value="light">Light</Radio.Button>
            <Radio.Button value="auto">Auto</Radio.Button>
            <Radio.Button value="dark">Dark</Radio.Button>
          </Radio.Group>
        </div>
      </Space>
    </div>
  );
}
