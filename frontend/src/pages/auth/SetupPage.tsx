import { useState } from 'react';
import { Card, Form, Input, Button, Typography, Alert } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

const { Title, Paragraph } = Typography;

export function SetupPage() {
  const { setup, isAuthenticated, needsSetup } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  if (!needsSetup) {
    return <Navigate to="/login" replace />;
  }

  async function handleSubmit(values: {
    username: string;
    password: string;
    confirmPassword: string;
  }) {
    if (values.password !== values.confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    setError(null);
    setLoading(true);
    try {
      await setup(values.username, values.password);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Setup failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card style={{ width: 400, boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
      <Title level={3} style={{ textAlign: 'center', marginBottom: 8 }}>
        Bes Ops
      </Title>
      <Paragraph style={{ textAlign: 'center', marginBottom: 24 }}>
        Create your admin account to get started.
      </Paragraph>
      {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} />}
      <Form onFinish={handleSubmit} layout="vertical" size="large">
        <Form.Item name="username" rules={[{ required: true, message: 'Username is required' }]}>
          <Input prefix={<UserOutlined />} placeholder="Username" autoFocus />
        </Form.Item>
        <Form.Item
          name="password"
          rules={[{ required: true, min: 6, message: 'Minimum 6 characters' }]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="Password" />
        </Form.Item>
        <Form.Item
          name="confirmPassword"
          rules={[{ required: true, message: 'Please confirm your password' }]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="Confirm Password" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>
            Create Account
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}
