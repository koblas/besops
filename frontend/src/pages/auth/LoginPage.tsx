import { useState } from 'react';
import { Card, Form, Input, Button, Typography, Alert } from 'antd';
import { UserOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

const { Title } = Typography;

export function LoginPage() {
  const { login, isAuthenticated } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [needs2FA, setNeeds2FA] = useState(false);
  const [form] = Form.useForm();

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  async function handleSubmit(values: { username: string; password: string; token?: string }) {
    setError(null);
    setLoading(true);
    try {
      const result = await login(values.username, values.password, values.token);
      if (result.tokenRequired) {
        setNeeds2FA(true);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card style={{ width: 400, boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
      <Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>
        Bes Ops
      </Title>
      {error && <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} />}
      <Form form={form} onFinish={handleSubmit} layout="vertical" size="large">
        <Form.Item name="username" rules={[{ required: true, message: 'Username is required' }]}>
          <Input prefix={<UserOutlined />} placeholder="Username" autoFocus />
        </Form.Item>
        <Form.Item name="password" rules={[{ required: true, message: 'Password is required' }]}>
          <Input.Password prefix={<LockOutlined />} placeholder="Password" />
        </Form.Item>
        {needs2FA && (
          <Form.Item name="token" rules={[{ required: true, message: '2FA token is required' }]}>
            <Input prefix={<SafetyOutlined />} placeholder="2FA Token" />
          </Form.Item>
        )}
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>
            Log In
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}
