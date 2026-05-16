import {
  GlobalOutlined,
  ApiOutlined,
  WifiOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  MailOutlined,
  CodeOutlined,
  DesktopOutlined,
  UploadOutlined,
  GroupOutlined,
} from '@ant-design/icons';

const iconMap: Record<string, React.ReactNode> = {
  http: <GlobalOutlined />,
  keyword: <GlobalOutlined />,
  'json-query': <GlobalOutlined />,
  'real-browser': <GlobalOutlined />,
  port: <ApiOutlined />,
  ping: <WifiOutlined />,
  'tailscale-ping': <WifiOutlined />,
  dns: <CloudServerOutlined />,
  docker: <DesktopOutlined />,
  push: <UploadOutlined />,
  steam: <DesktopOutlined />,
  gamedig: <DesktopOutlined />,
  mqtt: <MailOutlined />,
  'grpc-keyword': <CodeOutlined />,
  sqlserver: <DatabaseOutlined />,
  postgres: <DatabaseOutlined />,
  mysql: <DatabaseOutlined />,
  mongodb: <DatabaseOutlined />,
  redis: <DatabaseOutlined />,
  radius: <ApiOutlined />,
  rabbitmq: <MailOutlined />,
  snmp: <ApiOutlined />,
  group: <GroupOutlined />,
  manual: <CodeOutlined />,
};

interface MonitorTypeIconProps {
  type: string;
}

export function MonitorTypeIcon({ type }: MonitorTypeIconProps) {
  return <>{iconMap[type] ?? <GlobalOutlined />}</>;
}
