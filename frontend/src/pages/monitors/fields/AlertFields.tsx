import { Form, InputNumber, Switch } from 'antd';

export function AlertFields() {
  return (
    <>
      <Form.Item
        name="resendInterval"
        label="Resend Notification Every (checks)"
        initialValue={0}
        extra="Resend the alert every N checks while still down. 0 = notify only once."
      >
        <InputNumber min={0} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="expiryNotification"
        label="TLS Certificate Expiry Alert"
        valuePropName="checked"
        extra="Get notified before the TLS certificate expires (HTTP monitors only)."
      >
        <Switch />
      </Form.Item>
    </>
  );
}
