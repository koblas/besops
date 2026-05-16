import { Form, Input, Select, Typography } from 'antd';
import { Link } from 'react-router-dom';
import { useDockerHosts } from '../../../hooks/useDockerHosts';

const { Text } = Typography;

export function DockerFields() {
  const { data: hosts = [] } = useDockerHosts();

  return (
    <>
      <Form.Item
        name="dockerHost"
        label="Docker Host"
        rules={[{ required: true, message: 'Please select a Docker host' }]}
        extra={
          hosts.length === 0
            ? <Text type="secondary">No Docker hosts configured. <Link to="/settings/docker-hosts">Add one in Settings</Link> first.</Text>
            : 'The Docker daemon that manages this container.'
        }
      >
        <Select
          placeholder={hosts.length === 0 ? 'No hosts available' : 'Select Docker host'}
          options={hosts.map(h => ({ value: h.id, label: h.name }))}
          disabled={hosts.length === 0}
        />
      </Form.Item>

      <Form.Item
        name="dockerContainer"
        label="Container Name or ID"
        rules={[{ required: true, message: 'Please enter a container name or ID' }]}
        extra="The container to check. Must be running for the monitor to report UP."
      >
        <Input placeholder="my-app-container" />
      </Form.Item>
    </>
  );
}
