import { Form, Select, Typography } from 'antd';
import { Link } from 'react-router-dom';
import { useNotifications } from '../../../hooks/useNotifications';

const { Text } = Typography;

export function NotificationSelector() {
  const { data: notifications = [] } = useNotifications();

  return (
    <Form.Item
      name="notificationIds"
      label="Notifications"
      extra={notifications.length === 0
        ? <Text type="secondary">No notification providers configured. <Link to="/settings/notifications">Set one up</Link> to get alerted when this monitor goes down.</Text>
        : undefined
      }
    >
      <Select
        mode="multiple"
        placeholder={notifications.length === 0 ? 'No providers available' : 'Select notification providers'}
        options={notifications.map(n => ({ value: n.id, label: `${n.name} (${n.type})` }))}
        disabled={notifications.length === 0}
      />
    </Form.Item>
  );
}
