import { useState } from 'react';
import { Typography, Button, Table, Space, Switch, Empty, Tooltip, message, Modal, Form, Input, Select, InputNumber } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined } from '@ant-design/icons';
import {
  useNotifications,
  useCreateNotification,
  useUpdateNotification,
  useDeleteNotification,
  useTestNotification,
} from '../../hooks/useNotifications';
import type { Notification, NotificationInput } from '../../hooks/useNotifications';

const { Title } = Typography;

const notificationTypes = [
  { value: 'webhook', label: 'Webhook' },
  { value: 'slack', label: 'Slack' },
  { value: 'discord', label: 'Discord' },
  { value: 'telegram', label: 'Telegram' },
  { value: 'smtp', label: 'Email (SMTP)' },
  { value: 'pagerduty', label: 'PagerDuty' },
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
  const selectedType = Form.useWatch('type', form);

  function openCreate() {
    setEditing(null);
    form.resetFields();
    setModalOpen(true);
  }

  function openEdit(record: Notification) {
    setEditing(record);
    const cfg = record.config ?? {};
    form.setFieldsValue({
      name: record.name,
      type: record.type,
      active: record.active,
      ...cfg,
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
    const config = buildConfig(values.type, values);
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
          onSuccess: () => { message.success('Updated'); setModalOpen(false); },
          onError: () => message.error('Failed to save notification'),
        },
      );
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Created'); setModalOpen(false); },
        onError: () => message.error('Failed to create notification'),
      });
    }
  }

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'Type', dataIndex: 'type', key: 'type',
      render: (type: string) => notificationTypes.find(t => t.value === type)?.label ?? type,
    },
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
        destroyOnClose
      >
        <Form form={form} layout="vertical" initialValues={{ active: true }} preserve={false}>
          <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Give this notification a name' }]}>
            <Input placeholder="e.g. Production Alerts" />
          </Form.Item>
          <Form.Item name="type" label="Provider" rules={[{ required: true, message: 'Select a provider' }]}>
            <Select placeholder="Select provider" options={notificationTypes} />
          </Form.Item>
          <Form.Item name="active" label="Active" valuePropName="checked">
            <Switch />
          </Form.Item>

          {selectedType === 'webhook' && <WebhookFields />}
          {selectedType === 'slack' && <SlackFields />}
          {selectedType === 'discord' && <DiscordFields />}
          {selectedType === 'telegram' && <TelegramFields />}
          {selectedType === 'smtp' && <SmtpFields />}
          {selectedType === 'pagerduty' && <PagerDutyFields />}
        </Form>
      </Modal>
    </div>
  );
}

function WebhookFields() {
  return (
    <>
      <Form.Item name="webhookURL" label="URL" rules={[{ required: true, message: 'Webhook URL is required' }]}>
        <Input placeholder="https://example.com/webhook" />
      </Form.Item>
      <Form.Item name="webhookContentType" label="Content Type" initialValue="application/json">
        <Select options={[
          { value: 'application/json', label: 'application/json' },
          { value: 'application/x-www-form-urlencoded', label: 'application/x-www-form-urlencoded' },
        ]} />
      </Form.Item>
      <Form.Item
        name="webhookAdditionalHeaders"
        label="Additional Headers"
        extra='JSON object of headers to include, e.g. {"Authorization": "Bearer ..."}'
      >
        <Input.TextArea rows={3} placeholder='{"Authorization": "Bearer token"}' style={{ fontFamily: 'monospace' }} />
      </Form.Item>
    </>
  );
}

function SlackFields() {
  return (
    <>
      <Form.Item name="slackwebhookURL" label="Webhook URL" rules={[{ required: true, message: 'Slack webhook URL is required' }]}>
        <Input placeholder="https://hooks.slack.com/services/..." />
      </Form.Item>
      <Form.Item name="slackchannel" label="Channel" extra="Override the default channel (optional).">
        <Input placeholder="#alerts" />
      </Form.Item>
      <Form.Item name="slackusername" label="Bot Username" extra="Display name for the message (default: Bes Ops).">
        <Input placeholder="Bes Ops" />
      </Form.Item>
      <Form.Item name="slackiconemo" label="Icon Emoji" extra="Custom emoji for the bot avatar.">
        <Input placeholder=":rotating_light:" />
      </Form.Item>
    </>
  );
}

function SmtpFields() {
  return (
    <>
      <Form.Item name="smtpHost" label="SMTP Host" rules={[{ required: true, message: 'SMTP host is required' }]}>
        <Input placeholder="smtp.example.com" />
      </Form.Item>
      <Form.Item name="smtpPort" label="Port" initialValue={587}>
        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
      </Form.Item>
      <Form.Item name="smtpSecurity" label="Security" initialValue="">
        <Select options={[
          { value: '', label: 'None' },
          { value: 'STARTTLS', label: 'STARTTLS' },
          { value: 'TLS', label: 'TLS' },
        ]} />
      </Form.Item>
      <Form.Item name="smtpUsername" label="Username">
        <Input autoComplete="off" />
      </Form.Item>
      <Form.Item name="smtpPassword" label="Password">
        <Input.Password autoComplete="new-password" />
      </Form.Item>
      <Form.Item name="smtpFrom" label="From Address" rules={[{ required: true, message: 'From address is required' }]}>
        <Input placeholder="alerts@example.com" />
      </Form.Item>
      <Form.Item name="smtpTo" label="To Address" rules={[{ required: true, message: 'Recipient is required' }]} extra="Comma-separated for multiple recipients.">
        <Input placeholder="team@example.com" />
      </Form.Item>
      <Form.Item name="smtpIgnoreTLSError" label="Ignore TLS Errors" valuePropName="checked">
        <Switch />
      </Form.Item>
    </>
  );
}

function PagerDutyFields() {
  return (
    <>
      <Form.Item name="pagerdutyIntegrationKey" label="Integration Key" rules={[{ required: true, message: 'Integration key is required' }]} extra="Events API v2 integration key from your PagerDuty service.">
        <Input.Password placeholder="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" visibilityToggle />
      </Form.Item>
      <Form.Item name="pagerdutySeverity" label="Severity" initialValue="critical">
        <Select options={[
          { value: 'critical', label: 'Critical' },
          { value: 'error', label: 'Error' },
          { value: 'warning', label: 'Warning' },
          { value: 'info', label: 'Info' },
        ]} />
      </Form.Item>
    </>
  );
}

function DiscordFields() {
  return (
    <>
      <Form.Item name="discordWebhookURL" label="Webhook URL" rules={[{ required: true, message: 'Discord webhook URL is required' }]} extra="Create a webhook in your Discord channel settings under Integrations.">
        <Input placeholder="https://discord.com/api/webhooks/..." />
      </Form.Item>
      <Form.Item name="discordUsername" label="Bot Username" extra="Display name for the message (default: Bes Ops).">
        <Input placeholder="Bes Ops" />
      </Form.Item>
      <Form.Item name="discordPrefixMessage" label="Prefix Message" extra="Optional text before the embed (e.g. @everyone to ping).">
        <Input placeholder="@everyone" />
      </Form.Item>
    </>
  );
}

function TelegramFields() {
  return (
    <>
      <Form.Item name="telegramBotToken" label="Bot Token" rules={[{ required: true, message: 'Bot token is required' }]} extra="Get this from @BotFather on Telegram.">
        <Input.Password placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz" visibilityToggle />
      </Form.Item>
      <Form.Item name="telegramChatID" label="Chat ID" rules={[{ required: true, message: 'Chat ID is required' }]} extra="Use @userinfobot or @getidsbot to find your chat ID.">
        <Input placeholder="-1001234567890" />
      </Form.Item>
    </>
  );
}

function buildConfig(type: string, values: Record<string, unknown>): Record<string, unknown> {
  switch (type) {
    case 'webhook':
      return pick(values, ['webhookURL', 'webhookContentType', 'webhookAdditionalHeaders']);
    case 'slack':
      return pick(values, ['slackwebhookURL', 'slackchannel', 'slackusername', 'slackiconemo']);
    case 'discord':
      return pick(values, ['discordWebhookURL', 'discordUsername', 'discordPrefixMessage']);
    case 'telegram':
      return pick(values, ['telegramBotToken', 'telegramChatID']);
    case 'smtp':
      return pick(values, ['smtpHost', 'smtpPort', 'smtpSecurity', 'smtpUsername', 'smtpPassword', 'smtpFrom', 'smtpTo', 'smtpIgnoreTLSError']);
    case 'pagerduty':
      return pick(values, ['pagerdutyIntegrationKey', 'pagerdutySeverity']);
    default:
      return {};
  }
}

function pick(obj: Record<string, unknown>, keys: string[]): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const key of keys) {
    if (obj[key] !== undefined && obj[key] !== '' && obj[key] !== null) {
      result[key] = obj[key];
    }
  }
  return result;
}
