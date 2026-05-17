import { useState } from 'react';
import { Form, Input, Select, Button, Modal, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useMonitors, useCreateMonitor } from '../../../hooks/useMonitors';

const { Text } = Typography;

const monitorTypeGroups = [
  {
    label: 'HTTP',
    options: [
      { value: 'http', label: 'HTTP(s)' },
    ],
  },
  {
    label: 'Network',
    options: [
      { value: 'port', label: 'TCP Port' },
      { value: 'ping', label: 'Ping' },
      { value: 'dns', label: 'DNS' },
      { value: 'smtp', label: 'SMTP' },
      { value: 'tailscale-ping', label: 'Tailscale Ping' },
    ],
  },
  {
    label: 'Database',
    options: [
      { value: 'redis', label: 'Redis' },
    ],
  },
  {
    label: 'Messaging',
    options: [
      { value: 'mqtt', label: 'MQTT' },
    ],
  },
  {
    label: 'Infrastructure',
    options: [
      { value: 'grpc-keyword', label: 'gRPC' },
    ],
  },
  {
    label: 'Other',
    options: [
      { value: 'push', label: 'Push (passive)' },
      { value: 'group', label: 'Group' },
    ],
  },
];

interface GeneralFieldsProps {
  excludeId?: string;
}

export function GeneralFields({ excludeId }: GeneralFieldsProps) {
  const { data: monitors = [] } = useMonitors();
  const createMonitor = useCreateMonitor();
  const [modalOpen, setModalOpen] = useState(false);
  const [groupName, setGroupName] = useState('');
  const form = Form.useFormInstance();

  const groups = monitors.filter(m => m.type === 'group' && m.id !== excludeId);

  function handleCreateGroup() {
    if (!groupName.trim()) return;
    createMonitor.mutate(
      {
        name: groupName.trim(),
        type: 'group',
        active: false,
        interval: 60,
        maxRetries: 0,
        timeout: 48,
        retryInterval: 60,
        maxRedirects: 10,
        method: 'GET',
        resendInterval: 0,
        packetSize: 56,
      },
      {
        onSuccess: (data) => {
          message.success(`Group "${groupName.trim()}" created`);
          form.setFieldValue('parentId', data.id);
          setGroupName('');
          setModalOpen(false);
        },
        onError: () => message.error('Failed to create group'),
      },
    );
  }

  return (
    <>
      <Form.Item
        name="type"
        label="Monitor Type"
        rules={[{ required: true, message: 'Please select a monitor type' }]}
        extra="Choose how this monitor will check your service."
      >
        <Select
          options={monitorTypeGroups}
          placeholder="Select monitor type"
          showSearch
          optionFilterProp="label"
        />
      </Form.Item>

      <Form.Item
        name="name"
        label="Friendly Name"
        rules={[{ required: true, message: 'Please enter a name' }]}
        extra="A name to identify this monitor in the dashboard."
      >
        <Input placeholder="My Website" />
      </Form.Item>

      <Form.Item label="Parent Group" extra="Organize monitors into groups for easier navigation.">
        <div style={{ display: 'flex', gap: 8 }}>
          <Form.Item name="parentId" noStyle>
            <Select
              allowClear
              placeholder="None (top level)"
              options={groups.map(g => ({ value: g.id, label: g.name }))}
              style={{ flex: 1 }}
            />
          </Form.Item>
          <Button icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            New Group
          </Button>
        </div>
      </Form.Item>

      <Modal
        title="New Group"
        open={modalOpen}
        onOk={handleCreateGroup}
        onCancel={() => { setModalOpen(false); setGroupName(''); }}
        okText="Create"
        okButtonProps={{ disabled: !groupName.trim(), loading: createMonitor.isPending }}
      >
        <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
          Groups let you organize monitors into collapsible sections.
        </Text>
        <Input
          value={groupName}
          onChange={e => setGroupName(e.target.value)}
          placeholder="e.g. Production Servers"
          onPressEnter={handleCreateGroup}
          autoFocus
        />
      </Modal>
    </>
  );
}
