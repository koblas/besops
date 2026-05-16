import { Typography, Button, Table, Space, Tag, Tooltip, Result, Modal, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PauseCircleOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import {
  useMaintenanceList,
  useDeleteMaintenance,
  usePauseMaintenance,
  useResumeMaintenance,
} from '../../hooks/useMaintenance';
import type { Maintenance } from '../../hooks/useMaintenance';

const { Title } = Typography;

const strategyLabels: Record<string, string> = {
  manual: 'Manual',
  single: 'Single',
  'recurring-interval': 'Recurring (Interval)',
  'recurring-weekday': 'Recurring (Weekday)',
  'recurring-day-of-month': 'Recurring (Day of Month)',
  cron: 'Cron',
};

export function MaintenanceList() {
  const navigate = useNavigate();
  const { data: items = [], isLoading, isError } = useMaintenanceList();
  const deleteMutation = useDeleteMaintenance();
  const pauseMutation = usePauseMaintenance();
  const resumeMutation = useResumeMaintenance();

  function handleDelete(record: Maintenance) {
    Modal.confirm({
      title: 'Delete Maintenance',
      content: `Delete "${record.title}"?`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.id).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  const columns = [
    { title: 'Title', dataIndex: 'title', key: 'title' },
    {
      title: 'Strategy',
      dataIndex: 'strategy',
      key: 'strategy',
      render: (s: string) => strategyLabels[s] ?? s,
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      width: 100,
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>{active ? 'Active' : 'Paused'}</Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: unknown, record: Maintenance) => (
        <Space size="small">
          {record.active ? (
            <Button size="small" icon={<PauseCircleOutlined />} onClick={() => pauseMutation.mutate(record.id, { onSuccess: () => message.success('Paused'), onError: () => message.error('Failed to pause') })}>
              Pause
            </Button>
          ) : (
            <Button size="small" icon={<PlayCircleOutlined />} onClick={() => resumeMutation.mutate(record.id, { onSuccess: () => message.success('Resumed'), onError: () => message.error('Failed to resume') })}>
              Resume
            </Button>
          )}
          <Tooltip title="Edit"><Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/maintenance/edit/${record.id}`)} /></Tooltip>
          <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
        </Space>
      ),
    },
  ];

  if (isError) {
    return <Result status="error" title="Failed to load maintenance windows" subTitle="Check your connection and try again." />;
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Maintenance</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/add-maintenance')}>
          Add Maintenance
        </Button>
      </div>

      <Table
        dataSource={items}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={{ pageSize: 20 }}
      />
    </div>
  );
}
