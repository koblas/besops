import { useState } from 'react';
import { Typography, Button, Table, Space, Switch, Empty, Tooltip, message, Modal, Form, Input, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined } from '@ant-design/icons';
import {
  useNotifications,
  useCreateNotification,
  useUpdateNotification,
  useDeleteNotification,
  useTestNotification,
} from '../../hooks/useNotifications';
import type { Notification, NotificationInput } from '../../hooks/useNotifications';
import { JsonEditor } from '../../components/JsonEditor';

const { Title } = Typography;

const notificationTypes = [
  'webhook', 'smtp', 'telegram', 'discord', 'slack', 'pushover',
  'signal', 'gotify', 'ntfy', 'apprise', 'pushbullet', 'line',
  'mattermost', 'matrix', 'teams', 'opsgenie', 'pagerduty',
  'rocket.chat', 'lunasea', 'feishu', 'alerta', 'onebot',
];

export function NotificationSettings() {
  const { data: notifications = [], isLoading } = useNotifications();
  const createMutation = useCreateNotification();
  const updateMutation = useUpdateNotification();
  const deleteMutation = useDeleteNotification();
  const testMutation = useTestNotification();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Notification | null>(null);
  const [form] = Form.useForm();

  function openCreate() {
    setEditing(null);
    form.resetFields();
    setModalOpen(true);
  }

  function openEdit(record: Notification) {
    setEditing(record);
    form.setFieldsValue({
      name: record.name,
      type: record.type,
      active: record.active,
      config: JSON.stringify(record.config ?? {}, null, 2),
    });
    setModalOpen(true);
  }

  function handleDelete(record: Notification) {
    Modal.confirm({
      title: 'Delete Notification',
      content: `Delete "${record.name}"? This cannot be undone.`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.id).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  function handleTest(id: string) {
    testMutation.mutate(id, {
      onSuccess: () => message.success('Test notification sent'),
      onError: () => message.error('Test failed'),
    });
  }

  async function handleSubmit() {
    const values = await form.validateFields();
    let config: Record<string, unknown>;
    try {
      config = values.config ? JSON.parse(values.config) : {};
    } catch {
      message.error('Invalid JSON in config');
      return;
    }
    const input: NotificationInput = {
      name: values.name,
      type: values.type,
      active: values.active ?? true,
      config,
      applyExisting: false,
    };

    if (editing) {
      updateMutation.mutate(
        { id: editing.id, input },
        {
          onSuccess: () => {
            message.success('Updated');
            setModalOpen(false);
          },
          onError: () => message.error('Failed to save notification'),
        },
      );
    } else {
      createMutation.mutate(input, {
        onSuccess: () => {
          message.success('Created');
          setModalOpen(false);
        },
        onError: () => message.error('Failed to create notification'),
      });
    }
  }

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    {
      title: 'Active',
      dataIndex: 'active',
      key: 'active',
      width: 80,
      render: (active: boolean) => <Switch checked={active} disabled size="small" />,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: unknown, record: Notification) => (
        <Space size="small">
          <Button size="small" icon={<SendOutlined />} onClick={() => handleTest(record.id)} loading={testMutation.isPending}>
            Test
          </Button>
          <Tooltip title="Edit"><Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)} /></Tooltip>
          <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Notifications</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Add Notification
        </Button>
      </div>

      <Table
        dataSource={notifications}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={false}
        locale={{ emptyText: <Empty description="No notifications configured. Add one to get alerted when a monitor goes down." /> }}
      />

      <Modal
        open={modalOpen}
        title={editing ? 'Edit Notification' : 'New Notification'}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical" initialValues={{ active: true }}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="My Webhook" />
          </Form.Item>
          <Form.Item name="type" label="Type" rules={[{ required: true }]}>
            <Select
              showSearch
              placeholder="Select provider"
              options={notificationTypes.map(t => ({ value: t, label: t }))}
            />
          </Form.Item>
          <Form.Item name="active" label="Active" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="config" label="Configuration (JSON)">
            <JsonEditor placeholder='{"webhookUrl": "https://..."}' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
