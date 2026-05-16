import { Typography, Card, Descriptions, Spin } from 'antd';
import { useServerInfo, useDatabaseSize } from '../../hooks/useSystem';

const { Title } = Typography;

export function AboutSettings() {
  const { data: info, isLoading } = useServerInfo();
  const { data: dbSizeData } = useDatabaseSize();

  if (isLoading) return <Spin />;

  const sizeMB = dbSizeData?.size ? (dbSizeData.size / 1024 / 1024).toFixed(2) : '—';

  return (
    <div>
      <Title level={4}>About</Title>
      <Card size="small">
        <Descriptions column={1} size="small">
          <Descriptions.Item label="Version">{info?.version ?? '—'}</Descriptions.Item>
          <Descriptions.Item label="Latest Version">{info?.latestVersion ?? '—'}</Descriptions.Item>
          <Descriptions.Item label="Primary Base URL">{info?.primaryBaseURL ?? '—'}</Descriptions.Item>
          <Descriptions.Item label="Database Size">{sizeMB} MB</Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
}
