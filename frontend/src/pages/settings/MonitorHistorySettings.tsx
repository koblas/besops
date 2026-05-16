import { Typography, Card, Button, Space, Statistic, message, Modal } from 'antd';
import { DeleteOutlined, CompressOutlined } from '@ant-design/icons';
import { useDatabaseSize, useShrinkDatabase, useClearStatistics } from '../../hooks/useSystem';
import { useSettings, useUpdateSettings } from '../../hooks/useSettings';

const { Title } = Typography;

export function MonitorHistorySettings() {
  const { data: settings } = useSettings();
  const { data: dbSizeData, refetch: refetchSize } = useDatabaseSize();
  const shrinkMutation = useShrinkDatabase();
  const clearMutation = useClearStatistics();
  const updateMutation = useUpdateSettings();

  function handleShrink() {
    shrinkMutation.mutate(undefined, {
      onSuccess: () => {
        message.success('Database vacuumed');
        refetchSize();
      },
      onError: () => message.error('Failed to shrink database'),
    });
  }

  function handleClear() {
    Modal.confirm({
      title: 'Clear All Statistics',
      content: 'This will permanently delete all heartbeat data. This cannot be undone.',
      okText: 'Clear',
      okType: 'danger',
      onOk: () =>
        clearMutation.mutateAsync().then(() => {
          message.success('Statistics cleared');
          refetchSize();
        }).catch(() => message.error('Failed to clear statistics')),
    });
  }

  function handleSaveRetention(days: number) {
    updateMutation.mutate(
      { keepDataPeriodDays: days },
      {
        onSuccess: () => message.success('Retention period updated'),
        onError: () => message.error('Failed to update retention period'),
      },
    );
  }

  const sizeMB = dbSizeData?.size ? (dbSizeData.size / 1024 / 1024).toFixed(2) : '—';

  return (
    <div>
      <Title level={4}>Monitor History</Title>

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card size="small">
          <Statistic title="Database Size" value={`${sizeMB} MB`} />
        </Card>

        <Card title="Data Retention" size="small">
          <Space>
            {[7, 30, 90, 180, 365].map(d => (
              <Button
                key={d}
                type={settings?.keepDataPeriodDays === d ? 'primary' : 'default'}
                onClick={() => handleSaveRetention(d)}
              >
                {d} days
              </Button>
            ))}
          </Space>
        </Card>

        <Card title="Database Management" size="small">
          <Space>
            <Button icon={<CompressOutlined />} onClick={handleShrink} loading={shrinkMutation.isPending}>
              Shrink Database
            </Button>
            <Button icon={<DeleteOutlined />} danger onClick={handleClear} loading={clearMutation.isPending}>
              Clear Statistics
            </Button>
          </Space>
        </Card>
      </Space>
    </div>
  );
}
