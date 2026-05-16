import { Form, Input, InputNumber } from 'antd';

export function TcpFields() {
  return (
    <>
      <Form.Item
        name="hostname"
        label="Hostname or IP"
        rules={[{ required: true, message: 'Please enter a hostname' }]}
      >
        <Input placeholder="example.com" />
      </Form.Item>

      <Form.Item
        name="port"
        label="Port"
        rules={[{ required: true, message: 'Please enter a port' }]}
        extra="The TCP port to connect to. The monitor reports UP if a connection can be established."
      >
        <InputNumber min={1} max={65535} placeholder="443" style={{ width: '100%' }} />
      </Form.Item>
    </>
  );
}
