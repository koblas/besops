import { Typography, Button, Table, Space, Tag, Tooltip, Modal, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useStatusPages, useDeleteStatusPage } from '../../hooks/useStatusPages';
import type { StatusPage } from '../../hooks/useStatusPages';

const { Title } = Typography;

export function StatusPageManage() {
  const navigate = useNavigate();
  const { data: pages = [], isLoading } = useStatusPages();
  const deleteMutation = useDeleteStatusPage();

  function handleDelete(record: StatusPage) {
    Modal.confirm({
      title: 'Delete Status Page',
      content: `Delete "${record.title}"? This cannot be undone.`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.slug).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  const columns = [
    { title: 'Title', dataIndex: 'title', key: 'title' },
    { title: 'Slug', dataIndex: 'slug', key: 'slug' },
    {
      title: 'Published',
      dataIndex: 'published',
      key: 'published',
      width: 100,
      render: (v: boolean) => <Tag color={v ? 'green' : 'default'}>{v ? 'Yes' : 'No'}</Tag>,
    },
    {
      title: 'Groups',
      dataIndex: 'groups',
      key: 'groups',
      width: 80,
      render: (groups: StatusPage['groups']) => groups?.length ?? 0,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 180,
      render: (_: unknown, record: StatusPage) => (
        <Space size="small">
          <Tooltip title="View"><Button size="small" icon={<EyeOutlined />} onClick={() => window.open(`/status/${record.slug}`, '_blank')} /></Tooltip>
          <Tooltip title="Edit"><Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/add-status-page?slug=${record.slug}`)} /></Tooltip>
          <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Status Pages</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/add-status-page')}>
          New Status Page
        </Button>
      </div>

      <Table
        dataSource={pages}
        columns={columns}
        rowKey="slug"
        loading={isLoading}
        size="small"
        pagination={false}
      />
    </div>
  );
}
