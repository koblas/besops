import { useState, useEffect } from 'react';
import { Typography, Form, Input, Button, Card, Space, message, Alert } from 'antd';
import { api } from '../../api/client';

const { Title, Text } = Typography;

export function SecuritySettings() {
  const [passwordForm] = Form.useForm();
  const [twoFAEnabled, setTwoFAEnabled] = useState<boolean | null>(null);
  const [qrUri, setQrUri] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [preparing, setPreparing] = useState(false);
  const [setupForm] = Form.useForm();
  const [disableForm] = Form.useForm();

  useEffect(() => {
    async function loadStatus() {
      const { data } = await api.GET('/auth/2fa');
      if (data) setTwoFAEnabled((data as { enabled: boolean }).enabled);
    }
    loadStatus();
  }, []);

  async function handleChangePassword(values: { currentPassword: string; newPassword: string; confirmPassword: string }) {
    if (values.newPassword !== values.confirmPassword) {
      message.error('Passwords do not match');
      return;
    }
    setSaving(true);
    const { error } = await api.PUT('/auth/password', {
      body: { currentPassword: values.currentPassword, newPassword: values.newPassword },
    });
    setSaving(false);
    if (error) {
      message.error('Failed to change password');
    } else {
      message.success('Password changed');
      passwordForm.resetFields();
    }
  }

  async function handlePrepare2FA(values: { currentPassword: string }) {
    setPreparing(true);
    const { data, error } = await api.POST('/auth/2fa/prepare', { body: { currentPassword: values.currentPassword } });
    setPreparing(false);
    if (error) {
      message.error('Failed to prepare 2FA — check your password');
      return;
    }
    const resp = data as { uri?: string };
    if (resp.uri) setQrUri(resp.uri);
  }

  async function handleEnable2FA(values: { currentPassword: string; token: string }) {
    const { error } = await api.POST('/auth/2fa/enable', { body: { currentPassword: values.currentPassword, token: values.token } });
    if (error) {
      message.error('Invalid token or password');
    } else {
      message.success('2FA enabled');
      setTwoFAEnabled(true);
      setQrUri(null);
      setupForm.resetFields();
    }
  }

  async function handleDisable2FA(values: { currentPassword: string }) {
    const { error } = await api.POST('/auth/2fa/disable', { body: { currentPassword: values.currentPassword } });
    if (error) {
      message.error('Failed to disable 2FA');
    } else {
      message.success('2FA disabled');
      setTwoFAEnabled(false);
      disableForm.resetFields();
    }
  }

  return (
    <div style={{ maxWidth: 500 }}>
      <Title level={4}>Security</Title>

      <Card title="Change Password" size="small" style={{ marginBottom: 24 }}>
        <Form form={passwordForm} layout="vertical" onFinish={handleChangePassword}>
          <Form.Item name="currentPassword" label="Current Password" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="newPassword" label="New Password" rules={[{ required: true, min: 6 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="confirmPassword" label="Confirm New Password" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={saving}>
            Change Password
          </Button>
        </Form>
      </Card>

      <Card title="Two-Factor Authentication" size="small">
        {twoFAEnabled === true && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert type="success" message="2FA is enabled" showIcon />
            <Form form={disableForm} layout="inline" onFinish={handleDisable2FA}>
              <Form.Item name="currentPassword" rules={[{ required: true, message: 'Enter password' }]}>
                <Input.Password placeholder="Current password" />
              </Form.Item>
              <Button danger htmlType="submit">Disable 2FA</Button>
            </Form>
          </Space>
        )}
        {twoFAEnabled === false && !qrUri && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text>2FA is not enabled.</Text>
            <Form layout="inline" onFinish={handlePrepare2FA}>
              <Form.Item name="currentPassword" rules={[{ required: true, message: 'Enter password' }]}>
                <Input.Password placeholder="Current password" />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={preparing}>
                Setup 2FA
              </Button>
            </Form>
          </Space>
        )}
        {qrUri && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text>Scan this URI with your authenticator app:</Text>
            <Input.TextArea value={qrUri} readOnly rows={3} style={{ fontFamily: 'monospace', fontSize: 11 }} />
            <Form form={setupForm} layout="vertical" onFinish={handleEnable2FA}>
              <Form.Item name="currentPassword" label="Current Password" rules={[{ required: true }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item name="token" label="TOTP Code" rules={[{ required: true, message: 'Enter 6-digit code' }]}>
                <Input placeholder="123456" maxLength={6} />
              </Form.Item>
              <Button type="primary" htmlType="submit">Verify & Enable</Button>
            </Form>
          </Space>
        )}
      </Card>
    </div>
  );
}
