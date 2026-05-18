import { useEffect, useState } from 'react';
import { Form, Button, Card, Space, Spin, Typography, Result, Input, Select, message } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import { useMonitor, useCreateMonitor, useUpdateMonitor } from '../../hooks/useMonitors';
import { useAddMonitorTag, useRemoveMonitorTag, useTags } from '../../hooks/useTags';
import type { MonitorInput } from '../../hooks/useMonitors';

const { Title, Text } = Typography;

export function GroupForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const isEdit = !!id;

  const { data: existing, isLoading, isError } = useMonitor(id);
  const { data: tags = [] } = useTags();
  const createMutation = useCreateMonitor();
  const updateMutation = useUpdateMonitor();
  const addTag = useAddMonitorTag();
  const removeTag = useRemoveMonitorTag();

  const [tagIds, setTagIds] = useState<string[]>([]);
  const [originalTagIds, setOriginalTagIds] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (existing) {
      const config = existing.config as Record<string, unknown> | undefined;
      form.setFieldsValue({
        name: existing.name,
        groupTagIds: config && 'tagIds' in config ? config.tagIds : [],
      });

      const ids = (existing.tags ?? []).map(t => t.tagId).filter((id): id is string => !!id);
      setTagIds(ids);
      if (isEdit) {
        setOriginalTagIds(ids);
      }
    }
  }, [existing, form, isEdit]);

  if (id && isLoading) {
    return <div style={{ textAlign: 'center', marginTop: 48 }}><Spin size="large" /></div>;
  }

  if (id && isError) {
    return <Result status="error" title="Failed to load group" extra={<Button onClick={() => navigate('/dashboard')}>Back to Dashboard</Button>} />;
  }

  const tagOptions = tags.map(t => ({ value: t.id, label: t.name }));

  async function syncTags(monitorId: string) {
    const toAdd = tagIds.filter(id => !originalTagIds.includes(id));
    const toRemove = originalTagIds.filter(id => !tagIds.includes(id));

    const promises: Promise<void>[] = [];
    for (const tId of toAdd) {
      promises.push(new Promise<void>((resolve, reject) => {
        addTag.mutate({ monitorId, tagId: tId }, { onSuccess: () => resolve(), onError: reject });
      }));
    }
    for (const tId of toRemove) {
      promises.push(new Promise<void>((resolve, reject) => {
        removeTag.mutate({ monitorId, tagId: tId }, { onSuccess: () => resolve(), onError: reject });
      }));
    }
    await Promise.all(promises);
  }

  async function handleSubmit(values: { name: string; groupTagIds: string[] }) {
    setSaving(true);

    const input: MonitorInput = {
      name: values.name,
      type: 'group',
      active: true,
      interval: 60,
      timeout: 48,
      maxRetries: 0,
      retryInterval: 60,
      resendInterval: 0,
      config: { kind: 'group' as const, ...(values.groupTagIds?.length ? { tagIds: values.groupTagIds } : {}) },
    };

    try {
      if (isEdit) {
        await new Promise<void>((resolve, reject) => {
          updateMutation.mutate({ id: id!, input }, { onSuccess: () => resolve(), onError: reject });
        });
        await syncTags(id!);
        message.success('Group updated');
        navigate(`/dashboard/${id}`);
      } else {
        const created = await new Promise<{ id: string }>((resolve, reject) => {
          createMutation.mutate(input, {
            onSuccess: (data) => resolve(data as { id: string }),
            onError: reject,
          });
        });
        await syncTags(created.id);
        message.success('Group created');
        navigate(`/dashboard/${created.id}`);
      }
    } catch {
      message.error('Failed to save group. Check your inputs and try again.');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div style={{ maxWidth: 520 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginBottom: 4 }}>{isEdit ? 'Edit Group' : 'New Group'}</Title>
        <Text type="secondary">Groups aggregate the status of monitors that share the selected tags.</Text>
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
      >
        <Card size="small" style={{ marginBottom: 16 }}>
          <Form.Item
            name="name"
            label="Group Name"
            rules={[{ required: true, message: 'Give your group a name' }]}
          >
            <Input placeholder="Production Services" />
          </Form.Item>

          <Form.Item
            name="groupTagIds"
            label="Member Tags"
            extra="Any monitor with at least one of these tags is a member of this group."
          >
            <Select
              mode="multiple"
              allowClear
              placeholder="Select tags..."
              options={tagOptions}
              optionFilterProp="label"
            />
          </Form.Item>

          <Text type="secondary" style={{ fontSize: 12 }}>
            Group status is re-evaluated every 60 seconds.
          </Text>
        </Card>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>
              {isEdit ? 'Save Group' : 'Create Group'}
            </Button>
            <Button onClick={() => navigate(isEdit ? `/dashboard/${id}` : '/dashboard')}>
              Cancel
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}
