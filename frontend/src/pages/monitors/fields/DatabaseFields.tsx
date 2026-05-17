import { Form, Input } from 'antd';

export function DatabaseFields() {
  return (
    <Form.Item
      name="hostname"
      label="Connection String"
      rules={[{ required: true, message: 'Please enter a connection string' }]}
      extra="Format: redis://host:6379"
    >
      <Input.Password
        placeholder="redis://localhost:6379"
        visibilityToggle
      />
    </Form.Item>
  );
}
