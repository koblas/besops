import { useState } from 'react';
import { Typography, Button, Table, Switch, Modal, Form, Input, DatePicker, Empty, Tooltip, message } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useAPIKeys, useCreateAPIKey, useDeleteAPIKey, useEnableAPIKey, useDisableAPIKey } from '../../hooks/useAPIKeys';
import type { APIKey, APIKeyInput } from '../../hooks/useAPIKeys';
import { CopyableText } from '../../components/CopyableText';

const { Title } = Typography;

export function APIKeysSettings() {
  const { data: keys = [], isLoading } = useAPIKeys();
  const createMutation = useCreateAPIKey();
  const deleteMutation = useDeleteAPIKey();
  const enableMutation = useEnableAPIKey();
  const disableMutation = useDisableAPIKey();
  const [modalOpen, setModalOpen] = useState(false);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [form] = Form.useForm();

  function handleDelete(record: APIKey) {
    Modal.confirm({
      title: 'Delete API Key',
      content: `Delete "${record.name}"? This cannot be undone.`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.id).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  function handleToggle(record: APIKey) {
    if (record.active) {
      disableMutation.mutate(record.id, { onError: () => message.error('Failed to disable API key') });
    } else {
      enableMutation.mutate(record.id, { onError: () => message.error('Failed to enable API key') });
    }
  }

  async function handleCreate() {
    const values = await form.validateFields();
    const input: APIKeyInput = {
      name: values.name,
      expires: values.expires?.toISOString(),
    };
    createMutation.mutate(input, {
      onSuccess: (data) => {
        setModalOpen(false);
        if (data.key) {
          setCreatedKey(data.key);
        }
        form.resetFields();
      },
      onError: () => message.error('Failed to create API key'),
    });
  }

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'Active',
      key: 'active',
      width: 80,
      render: (_: unknown, record: APIKey) => (
        <Switch checked={record.active} size="small" onChange={() => handleToggle(record)} />
      ),
    },
    {
      title: 'Expires',
      dataIndex: 'expires',
      key: 'expires',
      render: (v: string | undefined) => v ? new Date(v).toLocaleDateString() : 'Never',
    },
    {
      title: 'Created',
      dataIndex: 'createdDate',
      key: 'createdDate',
      render: (v: string | undefined) => v ? new Date(v).toLocaleDateString() : '—',
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_: unknown, record: APIKey) => (
        <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>API Keys</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          Create API Key
        </Button>
      </div>

      <Table
        dataSource={keys}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={false}
        locale={{ emptyText: <Empty description="No API keys yet. Create one to access the API programmatically." /> }}
      />

      <Modal
        open={modalOpen}
        title="Create API Key"
        onCancel={() => setModalOpen(false)}
        onOk={handleCreate}
        confirmLoading={createMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="My API Key" />
          </Form.Item>
          <Form.Item name="expires" label="Expires (optional)">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        open={!!createdKey}
        title="API Key Created"
        onCancel={() => setCreatedKey(null)}
        onOk={() => setCreatedKey(null)}
        cancelButtonProps={{ style: { display: 'none' } }}
      >
        <p>Copy this key now. It will not be shown again.</p>
        {createdKey && <CopyableText text={createdKey} />}
      </Modal>
    </div>
  );
}
