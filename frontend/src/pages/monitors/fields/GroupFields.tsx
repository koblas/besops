import { Form, InputNumber, Alert } from 'antd';

export function GroupFields() {
  return (
    <>
      <Alert
        type="info"
        showIcon
        message="Group monitors aggregate the status of their children"
        description="Add child monitors by setting their Parent Group to this monitor. The group's status will reflect the worst status among its active children."
        style={{ marginBottom: 16 }}
      />

      <Form.Item
        name="interval"
        label="Status Check Interval (seconds)"
        extra="How often to re-evaluate children's status. A shorter interval means faster propagation of child status changes to the group."
      >
        <InputNumber min={20} style={{ width: '100%' }} />
      </Form.Item>
    </>
  );
}
