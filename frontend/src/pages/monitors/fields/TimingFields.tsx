import { Form, InputNumber, Switch } from 'antd';

export function TimingFields() {
  return (
    <>
      <Form.Item
        name="interval"
        label="Check Interval (seconds)"
        initialValue={60}
        extra="How often to check this monitor. Minimum 20 seconds."
      >
        <InputNumber min={20} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="retryInterval"
        label="Retry Interval (seconds)"
        initialValue={60}
        extra="How long to wait before retrying after a failed check."
      >
        <InputNumber min={20} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="maxRetries"
        label="Retries Before Alert"
        initialValue={0}
        extra="Number of consecutive failures before marking as down. 0 = alert immediately."
      >
        <InputNumber min={0} max={10} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="timeout"
        label="Request Timeout (seconds)"
        initialValue={48}
        extra="How long to wait for a response before considering it failed."
      >
        <InputNumber min={1} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="upsideDown"
        label="Invert Status"
        valuePropName="checked"
        extra="Treat DOWN responses as UP and vice versa. Useful for monitoring that a service is intentionally offline."
      >
        <Switch />
      </Form.Item>
    </>
  );
}
