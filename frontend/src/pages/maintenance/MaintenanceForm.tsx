import { useEffect } from 'react';
import { Typography, Form, Input, Select, Switch, Button, InputNumber, DatePicker, TimePicker, Space, Spin, Checkbox, message } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import {
  useMaintenance,
  useCreateMaintenance,
  useUpdateMaintenance,
} from '../../hooks/useMaintenance';
import type { MaintenanceInput } from '../../hooks/useMaintenance';
import dayjs from 'dayjs';

const { Title } = Typography;
const { TextArea } = Input;

const strategies = [
  { value: 'manual', label: 'Manual' },
  { value: 'single', label: 'Single Window' },
  { value: 'recurring-weekday', label: 'Recurring (Weekday)' },
  { value: 'recurring-day-of-month', label: 'Recurring (Day of Month)' },
  { value: 'recurring-interval', label: 'Recurring (Interval)' },
  { value: 'cron', label: 'Cron' },
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
  const { data: existing, isLoading } = useMaintenance(id);
  const createMutation = useCreateMaintenance();
  const updateMutation = useUpdateMaintenance();

  useEffect(() => {
    if (existing) {
      const values = { ...existing } as Record<string, unknown>;
      if (mode === 'clone') {
        delete values.id;
        values.title = `${values.title} (Copy)`;
      }
      form.setFieldsValue(values);
    }
  }, [existing, form, mode]);

  if (id && isLoading) {
    return <div style={{ padding: 24, textAlign: 'center' }}><Spin size="large" /></div>;
  }

  const title = mode === 'clone' ? 'Clone Maintenance' : isEdit ? 'Edit Maintenance' : 'Add Maintenance';

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
        onSuccess: () => { message.success('Updated'); navigate('/maintenance'); },
        onError: () => message.error('Failed to save. Check your inputs and try again.'),
      });
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Created'); navigate('/maintenance'); },
        onError: () => message.error('Failed to create. Check your inputs and try again.'),
      });
    }
  }

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <div style={{ padding: 24, maxWidth: 600 }}>
      <Title level={4}>{title}</Title>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{ strategy: 'manual', active: true }}
      >
        <Form.Item name="title" label="Title" rules={[{ required: true }]}>
          <Input placeholder="Scheduled Maintenance" />
        </Form.Item>

        <Form.Item name="description" label="Description">
          <TextArea rows={3} placeholder="Optional description" />
        </Form.Item>

        <Form.Item name="active" label="Active" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item name="strategy" label="Strategy" rules={[{ required: true }]}>
          <Select options={strategies} />
        </Form.Item>

        {strategy === 'single' && (
          <>
            <Form.Item name="startDate" label="Start Date/Time">
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="endDate" label="End Date/Time">
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
          </>
        )}

        {strategy === 'recurring-weekday' && (
          <>
            <Form.Item name="weekdays" label="Days of Week">
              <Checkbox.Group options={weekdayOptions} />
            </Form.Item>
            <Form.Item name="startTime" label="Start Time">
              <TimePicker format="HH:mm" style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="endTime" label="End Time">
              <TimePicker format="HH:mm" style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="startDate" label="Effective From">
              <DatePicker style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="endDate" label="Effective Until">
              <DatePicker style={{ width: '100%' }} />
            </Form.Item>
          </>
        )}

        {strategy === 'recurring-day-of-month' && (
          <>
            <Form.Item name="daysOfMonth" label="Days of Month">
              <Select
                mode="multiple"
                placeholder="Select days"
                options={Array.from({ length: 31 }, (_, i) => ({ value: i + 1, label: `${i + 1}` }))}
              />
            </Form.Item>
            <Form.Item name="startTime" label="Start Time">
              <TimePicker format="HH:mm" style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="endTime" label="End Time">
              <TimePicker format="HH:mm" style={{ width: '100%' }} />
            </Form.Item>
          </>
        )}

        {strategy === 'recurring-interval' && (
          <>
            <Form.Item name="intervalDay" label="Interval (days)">
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="startTime" label="Start Time">
              <TimePicker format="HH:mm" style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="endTime" label="End Time">
              <TimePicker format="HH:mm" style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="startDate" label="First Occurrence">
              <DatePicker style={{ width: '100%' }} />
            </Form.Item>
          </>
        )}

        {strategy === 'cron' && (
          <>
            <Form.Item name="cron" label="Cron Expression" rules={[{ required: true }]}>
              <Input placeholder="0 2 * * *" style={{ fontFamily: 'monospace' }} />
            </Form.Item>
            <Form.Item name="durationMinutes" label="Duration (minutes)">
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
          </>
        )}

        {strategy !== 'manual' && (
          <Form.Item name="timezoneOption" label="Timezone">
            <Input placeholder="America/New_York" />
          </Form.Item>
        )}

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              {isEdit ? 'Save' : 'Create'}
            </Button>
            <Button onClick={() => navigate('/maintenance')}>Cancel</Button>
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}
