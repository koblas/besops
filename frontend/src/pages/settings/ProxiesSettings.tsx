import { useState } from 'react';
import { Typography, Button, Table, Space, Switch, Modal, Form, Input, Select, InputNumber, Empty, Tooltip, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useProxies, useCreateProxy, useUpdateProxy, useDeleteProxy } from '../../hooks/useProxies';
import type { Proxy, ProxyInput } from '../../hooks/useProxies';

const { Title } = Typography;

const protocols = ['http', 'https', 'socks', 'socks5', 'socks5h', 'socks4'];

export function ProxiesSettings() {
  const { data: proxies = [], isLoading } = useProxies();
  const createMutation = useCreateProxy();
  const updateMutation = useUpdateProxy();
  const deleteMutation = useDeleteProxy();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Proxy | null>(null);
  const [form] = Form.useForm();

  function openCreate() {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ protocol: 'http', port: 8080, active: true, default: false, auth: false, applyExisting: false });
    setModalOpen(true);
  }

  function openEdit(record: Proxy) {
    setEditing(record);
    form.setFieldsValue(record);
    setModalOpen(true);
  }

  function handleDelete(record: Proxy) {
    Modal.confirm({
      title: 'Delete Proxy',
      content: `Delete "${record.host}:${record.port}"?`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.id).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  async function handleSubmit() {
    const values = await form.validateFields();
    const input: ProxyInput = {
      protocol: values.protocol,
      host: values.host,
      port: values.port,
      active: values.active ?? true,
      default: values.default ?? false,
      auth: values.auth ?? false,
      username: values.username,
      password: values.password,
      applyExisting: values.applyExisting ?? false,
    };

    if (editing) {
      updateMutation.mutate({ id: editing.id, input }, {
        onSuccess: () => { message.success('Updated'); setModalOpen(false); },
        onError: () => message.error('Failed to save proxy'),
      });
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Created'); setModalOpen(false); },
        onError: () => message.error('Failed to create proxy'),
      });
    }
  }

  const authValue = Form.useWatch('auth', form);

  const columns = [
    {
      title: 'Proxy',
      key: 'proxy',
      render: (_: unknown, record: Proxy) => `${record.protocol}://${record.host}:${record.port}`,
    },
    {
      title: 'Active',
      dataIndex: 'active',
      key: 'active',
      width: 80,
      render: (v: boolean) => <Switch checked={v} disabled size="small" />,
    },
    {
      title: 'Default',
      dataIndex: 'default',
      key: 'default',
      width: 80,
      render: (v: boolean) => v ? 'Yes' : 'No',
    },
    {
      title: 'Auth',
      dataIndex: 'auth',
      key: 'auth',
      width: 80,
      render: (v: boolean) => v ? 'Yes' : 'No',
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: unknown, record: Proxy) => (
        <Space size="small">
          <Tooltip title="Edit"><Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)} /></Tooltip>
          <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Proxies</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Add Proxy
        </Button>
      </div>

      <Table
        dataSource={proxies}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={false}
        locale={{ emptyText: <Empty description="No proxies configured. Add one to route monitor checks through a proxy." /> }}
      />

      <Modal
        open={modalOpen}
        title={editing ? 'Edit Proxy' : 'New Proxy'}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="protocol" label="Protocol" rules={[{ required: true }]}>
            <Select options={protocols.map(p => ({ value: p, label: p }))} />
          </Form.Item>
          <Form.Item name="host" label="Host" rules={[{ required: true }]}>
            <Input placeholder="proxy.example.com" />
          </Form.Item>
          <Form.Item name="port" label="Port" rules={[{ required: true }]}>
            <InputNumber min={1} max={65535} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="active" label="Active" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="default" label="Default Proxy" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="auth" label="Requires Authentication" valuePropName="checked">
            <Switch />
          </Form.Item>
          {authValue && (
            <>
              <Form.Item name="username" label="Username">
                <Input />
              </Form.Item>
              <Form.Item name="password" label="Password">
                <Input.Password />
              </Form.Item>
            </>
          )}
          <Form.Item name="applyExisting" label="Apply to Existing Monitors" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
