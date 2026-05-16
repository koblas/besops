import { Form, Typography } from 'antd';
import { CopyableText } from '../../../components/CopyableText';

const { Text } = Typography;

export function PushFields() {
  const form = Form.useFormInstance();
  const pushToken = Form.useWatch('pushToken', form);

  if (!pushToken) {
    return (
      <Form.Item label="Push URL">
        <Text type="secondary">
          Save this monitor first — a unique push URL will be generated that your service can call to report its status.
        </Text>
      </Form.Item>
    );
  }

  const pushUrl = `${window.location.origin}/api/v1/push/${pushToken}?status=up&msg=OK&ping=`;

  return (
    <Form.Item label="Push URL" extra="Your service should send a GET or POST request to this URL at regular intervals. If no request arrives within the check interval, the monitor reports DOWN.">
      <CopyableText text={pushUrl} />
    </Form.Item>
  );
}
