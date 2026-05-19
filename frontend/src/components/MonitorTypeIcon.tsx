import {
  GlobalOutlined,
  ApiOutlined,
  WifiOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  MailOutlined,
  CodeOutlined,
  GroupOutlined,
} from '@ant-design/icons';

const iconMap: Record<string, React.ReactNode> = {
  http: <GlobalOutlined />,
  port: <ApiOutlined />,
  ping: <WifiOutlined />,
  'tailscale-ping': <WifiOutlined />,
  dns: <CloudServerOutlined />,
  smtp: <MailOutlined />,
  mqtt: <MailOutlined />,
  'grpc-keyword': <CodeOutlined />,
  redis: <DatabaseOutlined />,
  group: <GroupOutlined />,
};

interface MonitorTypeIconProps {
  type: string;
}

export function MonitorTypeIcon({ type }: MonitorTypeIconProps) {
  return <>{iconMap[type] ?? <GlobalOutlined />}</>;
}
