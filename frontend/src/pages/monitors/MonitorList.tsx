import { useState } from 'react';
import { Typography, Table, Button, Space, Input, Tag, Tooltip, Result, message, Modal } from 'antd';
import { PlusOutlined, DeleteOutlined, PauseCircleOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useMonitors, usePauseMonitor, useResumeMonitor, useDeleteMonitor } from '../../hooks/useMonitors';
import type { Monitor } from '../../hooks/useMonitors';
import { StatusBadge } from '../../components/StatusBadge';
import { MonitorTypeIcon } from '../../components/MonitorTypeIcon';
import type { StatusValue } from '../../lib/constants';

const { Title } = Typography;
const { Search } = Input;

export function MonitorList() {
  const navigate = useNavigate();
  const { data: monitors = [], isLoading, isError } = useMonitors();
  const pauseMutation = usePauseMonitor();
  const resumeMutation = useResumeMonitor();
  const deleteMutation = useDeleteMonitor();
  const [search, setSearch] = useState('');
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  if (isError) {
    return <Result status="error" title="Failed to load monitors" subTitle="Check your connection and try again." />;
  }

  const filtered = monitors.filter(m =>
    m.name.toLowerCase().includes(search.toLowerCase()),
  );

  function handleBulkDelete() {
    Modal.confirm({
      title: `Delete ${selectedIds.length} monitor(s)?`,
      content: 'This cannot be undone.',
      okText: 'Delete',
      okType: 'danger',
      onOk: async () => {
        try {
          for (const id of selectedIds) {
            await deleteMutation.mutateAsync(id);
          }
          setSelectedIds([]);
          message.success('Monitors deleted');
        } catch {
          message.error('Failed to delete some monitors');
        }
      },
    });
  }

  const columns = [
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'status',
      width: 70,
      render: (active: boolean) => <StatusBadge status={(active ? 1 : 2) as StatusValue} />,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: Monitor, b: Monitor) => a.name.localeCompare(b.name),
      render: (name: string, record: Monitor) => (
        <a onClick={() => navigate(`/dashboard/${record.id}`)}>{name}</a>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 140,
      render: (type: string) => (
        <Space size={4}>
          <MonitorTypeIcon type={type} />
          <span>{type}</span>
        </Space>
      ),
      filters: [...new Set(monitors.map(m => m.type))].map(t => ({ text: t, value: t })),
      onFilter: (value: unknown, record: Monitor) => record.type === value,
    },
    {
      title: 'Interval',
      dataIndex: 'interval',
      key: 'interval',
      width: 90,
      render: (v: number) => `${v}s`,
      sorter: (a: Monitor, b: Monitor) => a.interval - b.interval,
    },
    {
      title: 'Tags',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: Monitor['tags']) =>
        tags?.map(t => (
          <Tag key={t.tagId} color={t.color}>{t.name}{t.value ? `: ${t.value}` : ''}</Tag>
        )),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 140,
      render: (_: unknown, record: Monitor) => (
        <Space size="small">
          {record.active ? (
            <Tooltip title="Pause">
              <Button
                size="small"
                icon={<PauseCircleOutlined />}
                onClick={() => pauseMutation.mutate(record.id, {
                  onSuccess: () => message.success('Monitor paused'),
                  onError: () => message.error('Failed to pause monitor'),
                })}
              />
            </Tooltip>
          ) : (
            <Tooltip title="Resume">
              <Button
                size="small"
                icon={<PlayCircleOutlined />}
                onClick={() => resumeMutation.mutate(record.id, {
                  onSuccess: () => message.success('Monitor resumed'),
                  onError: () => message.error('Failed to resume monitor'),
                })}
              />
            </Tooltip>
          )}
          <Tooltip title="Delete">
            <Button
              size="small"
              icon={<DeleteOutlined />}
              danger
              onClick={() => {
                Modal.confirm({
                  title: `Delete "${record.name}"?`,
                  okText: 'Delete',
                  okType: 'danger',
                  onOk: () => deleteMutation.mutateAsync(record.id).catch(() => message.error('Failed to delete monitor')),
                });
              }}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>All Monitors</Title>
        <Space>
          {selectedIds.length > 0 && (
            <Button danger icon={<DeleteOutlined />} onClick={handleBulkDelete}>
              Delete ({selectedIds.length})
            </Button>
          )}
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/add')}>
            Add Monitor
          </Button>
        </Space>
      </div>

      <Search
        placeholder="Search monitors..."
        value={search}
        onChange={e => setSearch(e.target.value)}
        style={{ marginBottom: 16, maxWidth: 300 }}
        allowClear
      />

      <Table
        dataSource={filtered}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={{ pageSize: 20, showSizeChanger: true }}
        rowSelection={{
          selectedRowKeys: selectedIds,
          onChange: (keys) => setSelectedIds(keys as string[]),
        }}
      />
    </div>
  );
}
