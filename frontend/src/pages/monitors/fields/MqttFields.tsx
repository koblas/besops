import { Form, Input, InputNumber } from 'antd';

export function MqttFields() {
  return (
    <>
      <Form.Item
        name="hostname"
        label="MQTT Broker Host"
        rules={[{ required: true, message: 'Please enter a broker hostname' }]}
      >
        <Input placeholder="mqtt.example.com" />
      </Form.Item>

      <Form.Item name="port" label="Port" initialValue={1883} extra="Default: 1883 (unencrypted) or 8883 (TLS).">
        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="mqttTopic"
        label="Topic"
        rules={[{ required: true, message: 'Please enter a topic' }]}
        extra="The MQTT topic to subscribe to."
      >
        <Input placeholder="sensors/temperature" />
      </Form.Item>

      <Form.Item
        name="mqttSuccessMessage"
        label="Expected Message"
        extra="If set, the monitor reports UP only when this exact text is received on the topic. Leave empty to accept any message."
      >
        <Input placeholder="OK" />
      </Form.Item>

      <Form.Item name="mqttUsername" label="Username" extra="Leave empty if no authentication is required.">
        <Input autoComplete="off" />
      </Form.Item>

      <Form.Item name="mqttPassword" label="Password">
        <Input.Password autoComplete="new-password" />
      </Form.Item>
    </>
  );
}
