import { useEffect } from 'react';
import { Typography, Form, Input, Select, Switch, Button, Result, message, Spin } from 'antd';
import { useSettings, useUpdateSettings } from '../../hooks/useSettings';

const { Title } = Typography;

export function GeneralSettings() {
  const { data: settings, isLoading, isError } = useSettings();
  const updateMutation = useUpdateSettings();
  const [form] = Form.useForm();

  useEffect(() => {
    if (settings) {
      form.setFieldsValue(settings);
    }
  }, [settings, form]);

  if (isLoading) return <Spin />;
  if (isError) return <Result status="error" title="Failed to load settings" subTitle="Check your connection and try again." />;

  async function handleSubmit(values: Record<string, unknown>) {
    updateMutation.mutate(values, {
      onSuccess: () => message.success('Settings saved'),
      onError: () => message.error('Failed to save settings'),
    });
  }

  return (
    <div>
      <Title level={4}>General</Title>
      <Form form={form} layout="vertical" onFinish={handleSubmit} style={{ maxWidth: 500 }}>
        <Form.Item name="primaryBaseURL" label="Primary Base URL">
          <Input placeholder="https://status.example.com" />
        </Form.Item>

        <Form.Item name="serverTimezone" label="Server Timezone">
          <Input placeholder="America/New_York" />
        </Form.Item>

        <Form.Item name="entryPage" label="Entry Page">
          <Select
            options={[
              { value: 'dashboard', label: 'Dashboard' },
              { value: 'statusPage', label: 'Status Page' },
            ]}
          />
        </Form.Item>

        <Form.Item name="statusPageSlug" label="Default Status Page Slug">
          <Input placeholder="default" />
        </Form.Item>

        <Form.Item name="trustProxy" label="Trust Proxy" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item name="keepDataPeriodDays" label="Keep Data Period (days)">
          <Input type="number" />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>
            Save
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
}
