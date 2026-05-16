import { Form, Input, InputNumber } from 'antd';

export function PingFields() {
  return (
    <>
      <Form.Item
        name="hostname"
        label="Hostname or IP"
        rules={[{ required: true, message: 'Please enter a hostname or IP address' }]}
      >
        <Input placeholder="example.com" />
      </Form.Item>

      <Form.Item
        name="packetSize"
        label="Packet Size (bytes)"
        initialValue={56}
        extra="Size of the ICMP packet payload. Default (56) works for most cases."
      >
        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
      </Form.Item>
    </>
  );
}
