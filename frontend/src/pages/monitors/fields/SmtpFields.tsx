import { Form, Input, InputNumber, Select, Switch } from 'antd';

export function SmtpFields() {
  return (
    <>
      <Form.Item
        name="hostname"
        label="SMTP Host"
        rules={[{ required: true, message: 'Please enter the SMTP server hostname' }]}
      >
        <Input placeholder="smtp.example.com" />
      </Form.Item>

      <Form.Item
        name="port"
        label="Port"
        initialValue={25}
        extra="Common ports: 25 (plain), 465 (SSL/TLS), 587 (STARTTLS)"
      >
        <InputNumber min={1} max={65535} style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        name="smtpSecurity"
        label="Security"
        initialValue="none"
        extra="How to secure the connection to the SMTP server."
      >
        <Select
          options={[
            { value: 'none', label: 'None / Plain' },
            { value: 'starttls', label: 'STARTTLS' },
            { value: 'secure', label: 'SSL / TLS' },
          ]}
        />
      </Form.Item>

      <Form.Item
        name="ignoreTls"
        label="Ignore TLS Errors"
        valuePropName="checked"
        extra="Accept self-signed or expired certificates."
      >
        <Switch />
      </Form.Item>
    </>
  );
}
