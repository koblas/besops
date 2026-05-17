import { useEffect } from 'react';
import { Typography, Form, Input, Select, Switch, Button, Card, InputNumber, DatePicker, TimePicker, Space, Spin, Checkbox, Result, message, Alert } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import {
  useMaintenance,
  useCreateMaintenance,
  useUpdateMaintenance,
} from '../../hooks/useMaintenance';
import type { MaintenanceInput } from '../../hooks/useMaintenance';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;

const strategies = [
  { value: 'manual', label: 'Manual', description: 'Activate and deactivate manually' },
  { value: 'single', label: 'Single Window', description: 'One-time window with start and end' },
  { value: 'recurring-weekday', label: 'Recurring (Weekday)', description: 'Repeats on selected days of the week' },
  { value: 'recurring-day-of-month', label: 'Recurring (Day of Month)', description: 'Repeats on selected calendar days' },
  { value: 'recurring-interval', label: 'Recurring (Interval)', description: 'Repeats every N days' },
  { value: 'cron', label: 'Cron Expression', description: 'Custom schedule using cron syntax' },
];

const weekdayOptions = [
  { label: 'Sun', value: 0 },
  { label: 'Mon', value: 1 },
  { label: 'Tue', value: 2 },
  { label: 'Wed', value: 3 },
  { label: 'Thu', value: 4 },
  { label: 'Fri', value: 5 },
  { label: 'Sat', value: 6 },
];

export function MaintenanceForm({ mode }: { mode?: 'clone' }) {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const strategy = Form.useWatch('strategy', form);

  const isEdit = !!id && mode !== 'clone';
  const { data: existing, isLoading, isError } = useMaintenance(id);
  const createMutation = useCreateMaintenance();
  const updateMutation = useUpdateMaintenance();

  useEffect(() => {
    if (existing) {
      const values = { ...existing } as Record<string, unknown>;
      if (mode === 'clone') {
        delete values.id;
        values.title = `${values.title} (Copy)`;
      }
      if (values.startDate) values.startDate = dayjs(values.startDate as string);
      if (values.endDate) values.endDate = dayjs(values.endDate as string);
      if (values.startTime) values.startTime = dayjs(values.startTime as string, 'HH:mm');
      if (values.endTime) values.endTime = dayjs(values.endTime as string, 'HH:mm');
      form.setFieldsValue(values);
    }
  }, [existing, form, mode]);

  if (id && isLoading) {
    return (
      <div style={{ textAlign: 'center', marginTop: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (id && isError) {
    return <Result status="error" title="Failed to load maintenance window" extra={<Button onClick={() => navigate('/maintenance')}>Back to Maintenance</Button>} />;
  }

  const title = mode === 'clone' ? 'Clone Maintenance' : isEdit ? 'Edit Maintenance' : 'New Maintenance Window';
  const subtitle = isEdit
    ? 'Update the schedule for this maintenance window.'
    : 'Define when notifications should be suppressed for planned work.';

  async function handleSubmit(values: Record<string, unknown>) {
    const input: MaintenanceInput = {
      title: values.title as string,
      description: values.description as string | undefined,
      active: (values.active as boolean) ?? true,
      strategy: values.strategy as MaintenanceInput['strategy'],
      startDate: values.startDate ? (values.startDate as dayjs.Dayjs).toISOString() : undefined,
      endDate: values.endDate ? (values.endDate as dayjs.Dayjs).toISOString() : undefined,
      startTime: values.startTime ? (values.startTime as dayjs.Dayjs).format('HH:mm') : undefined,
      endTime: values.endTime ? (values.endTime as dayjs.Dayjs).format('HH:mm') : undefined,
      weekdays: values.weekdays as number[] | undefined,
      daysOfMonth: values.daysOfMonth as number[] | undefined,
      intervalDay: values.intervalDay as number | undefined,
      cron: values.cron as string | undefined,
      durationMinutes: values.durationMinutes as number | undefined,
      timezoneOption: values.timezoneOption as string | undefined,
    };

    if (isEdit) {
      updateMutation.mutate({ id: id!, input }, {
        onSuccess: () => { message.success('Maintenance window updated'); navigate('/maintenance'); },
        onError: () => message.error('Failed to save. Check your inputs and try again.'),
      });
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Maintenance window created'); navigate('/maintenance'); },
        onError: () => message.error('Failed to create. Check your inputs and try again.'),
      });
    }
  }

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <div style={{ maxWidth: 680, margin: '0 auto', padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginBottom: 4 }}>{title}</Title>
        <Text type="secondary">{subtitle}</Text>
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{ strategy: 'manual', active: true }}
      >
        <Card size="small" style={{ marginBottom: 16 }}>
          <Form.Item name="title" label="Title" rules={[{ required: true, message: 'Give this maintenance window a name' }]}>
            <Input placeholder="e.g. Weekly Database Backup" />
          </Form.Item>
          <Form.Item name="description" label="Description" extra="Visible on status pages during maintenance">
            <TextArea rows={2} placeholder="What's happening during this window?" />
          </Form.Item>
          <Form.Item name="active" label="Active" valuePropName="checked" extra="Inactive windows won't suppress notifications">
            <Switch />
          </Form.Item>
        </Card>

        <Card title="Schedule" size="small" style={{ marginBottom: 16 }}>
          <Form.Item name="strategy" label="Strategy" rules={[{ required: true }]}>
            <Select
              options={strategies}
              optionRender={(option) => (
                <div>
                  <div>{option.label}</div>
                  <Text type="secondary" style={{ fontSize: 12 }}>{(option.data as { description?: string }).description}</Text>
                </div>
              )}
            />
          </Form.Item>

          {strategy === 'manual' && (
            <Alert type="info" message="This window must be started and stopped manually from the maintenance list." showIcon />
          )}

          {strategy === 'single' && <SingleFields />}
          {strategy === 'recurring-weekday' && <RecurringWeekdayFields />}
          {strategy === 'recurring-day-of-month' && <RecurringDayOfMonthFields />}
          {strategy === 'recurring-interval' && <RecurringIntervalFields />}
          {strategy === 'cron' && <CronFields />}

          {strategy && strategy !== 'manual' && (
            <Form.Item name="timezoneOption" label="Timezone" extra="IANA timezone identifier">
              <Input placeholder="UTC" />
            </Form.Item>
          )}
        </Card>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              {isEdit ? 'Save Changes' : 'Create Window'}
            </Button>
            <Button onClick={() => navigate('/maintenance')}>
              Cancel
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}

function SingleFields() {
  return (
    <>
      <Form.Item name="startDate" label="Start" rules={[{ required: true, message: 'Start time is required' }]}>
        <DatePicker showTime format="YYYY-MM-DD HH:mm" style={{ width: '100%' }} placeholder="Select start date and time" />
      </Form.Item>
      <Form.Item name="endDate" label="End" rules={[{ required: true, message: 'End time is required' }]}>
        <DatePicker showTime format="YYYY-MM-DD HH:mm" style={{ width: '100%' }} placeholder="Select end date and time" />
      </Form.Item>
    </>
  );
}

function RecurringWeekdayFields() {
  return (
    <>
      <Form.Item name="weekdays" label="Days" rules={[{ required: true, message: 'Select at least one day' }]}>
        <Checkbox.Group options={weekdayOptions} />
      </Form.Item>
      <Space.Compact style={{ width: '100%', marginBottom: 16 }}>
        <Form.Item name="startTime" label="Window Start" style={{ flex: 1, marginBottom: 0 }} rules={[{ required: true }]}>
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="endTime" label="Window End" style={{ flex: 1, marginBottom: 0 }} rules={[{ required: true }]}>
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Space.Compact>
      <Space.Compact style={{ width: '100%' }}>
        <Form.Item name="startDate" label="Effective From" style={{ flex: 1 }} extra="Leave blank for no start limit">
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="endDate" label="Effective Until" style={{ flex: 1 }} extra="Leave blank for no end limit">
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>
      </Space.Compact>
    </>
  );
}

function RecurringDayOfMonthFields() {
  return (
    <>
      <Form.Item name="daysOfMonth" label="Days of Month" rules={[{ required: true, message: 'Select at least one day' }]}>
        <Select
          mode="multiple"
          placeholder="Select days (e.g. 1, 15)"
          options={Array.from({ length: 31 }, (_, i) => ({ value: i + 1, label: `${i + 1}` }))}
        />
      </Form.Item>
      <Space.Compact style={{ width: '100%' }}>
        <Form.Item name="startTime" label="Window Start" style={{ flex: 1 }} rules={[{ required: true }]}>
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="endTime" label="Window End" style={{ flex: 1 }} rules={[{ required: true }]}>
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Space.Compact>
    </>
  );
}

function RecurringIntervalFields() {
  return (
    <>
      <Form.Item name="intervalDay" label="Repeat Every (days)" rules={[{ required: true }]}>
        <InputNumber min={1} style={{ width: '100%' }} placeholder="e.g. 7 for weekly" />
      </Form.Item>
      <Space.Compact style={{ width: '100%' }}>
        <Form.Item name="startTime" label="Window Start" style={{ flex: 1 }} rules={[{ required: true }]}>
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="endTime" label="Window End" style={{ flex: 1 }} rules={[{ required: true }]}>
          <TimePicker format="HH:mm" style={{ width: '100%' }} />
        </Form.Item>
      </Space.Compact>
      <Form.Item name="startDate" label="First Occurrence" rules={[{ required: true }]} extra="The date of the first maintenance window">
        <DatePicker style={{ width: '100%' }} />
      </Form.Item>
    </>
  );
}

function CronFields() {
  return (
    <>
      <Form.Item name="cron" label="Cron Expression" rules={[{ required: true, message: 'Cron expression is required' }]} extra="Standard 5-field format: minute hour day month weekday">
        <Input placeholder="0 2 * * 0" style={{ fontFamily: 'monospace' }} />
      </Form.Item>
      <Form.Item name="durationMinutes" label="Duration (minutes)" rules={[{ required: true, message: 'Duration is required for cron schedules' }]} extra="How long each window lasts">
        <InputNumber min={1} style={{ width: '100%' }} placeholder="e.g. 60" />
      </Form.Item>
    </>
  );
}
