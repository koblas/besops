import { Form, Input, Select } from 'antd';

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
];

export function GeneralFields() {
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
    </>
  );
}
