import { useEffect, useState } from 'react';
import { Form, Button, Card, Space, Spin, Typography, Result, message } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import { useMonitor, useCreateMonitor, useUpdateMonitor } from '../../hooks/useMonitors';
import { useAddMonitorTag, useRemoveMonitorTag } from '../../hooks/useTags';
import type { MonitorInput } from '../../hooks/useMonitors';
import { GeneralFields } from './fields/GeneralFields';
import { HttpFields } from './fields/HttpFields';
import { TcpFields } from './fields/TcpFields';
import { PingFields } from './fields/PingFields';
import { DnsFields } from './fields/DnsFields';
import { SmtpFields } from './fields/SmtpFields';
import { MqttFields } from './fields/MqttFields';
import { GrpcFields } from './fields/GrpcFields';
import { DatabaseFields } from './fields/DatabaseFields';
import { PushFields } from './fields/PushFields';
import { GroupFields } from './fields/GroupFields';
import { TimingFields } from './fields/TimingFields';
import { AlertFields } from './fields/AlertFields';
import { NotificationSelector } from './fields/NotificationSelector';
import { TagSelector } from '../../components/TagSelector';

const { Title, Text } = Typography;

const httpTypes = ['http'];

export function MonitorForm({ mode }: { mode?: 'clone' }) {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const type = Form.useWatch('type', form);

  const isEdit = !!id && mode !== 'clone';
  const { data: existing, isLoading, isError } = useMonitor(id);
  const createMutation = useCreateMonitor();
  const updateMutation = useUpdateMonitor();
  const addTag = useAddMonitorTag();
  const removeTag = useRemoveMonitorTag();

  const [tagIds, setTagIds] = useState<string[]>([]);
  const [originalTagIds, setOriginalTagIds] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (existing) {
      const values = { ...existing } as Record<string, unknown>;
      if (mode === 'clone') {
        delete values.id;
        values.name = `${existing.name} (Copy)`;
      }
      if (values.parentId === null) {
        delete values.parentId;
      }
      form.setFieldsValue(values);

      const ids = (existing.tags ?? []).map(t => t.tagId).filter((id): id is string => !!id);
      setTagIds(ids);
      if (isEdit) {
        setOriginalTagIds(ids);
      }
    }
  }, [existing, form, mode, isEdit]);

  if (id && isLoading) {
    return (
      <div style={{ textAlign: 'center', marginTop: 48 }}>
        <Spin size="large" tip="Loading monitor..." />
      </div>
    );
  }

  if (id && isError) {
    return <Result status="error" title="Failed to load monitor" extra={<Button onClick={() => navigate('/dashboard')}>Back to Dashboard</Button>} />;
  }

  const title = mode === 'clone' ? 'Clone Monitor' : isEdit ? 'Edit Monitor' : 'Add New Monitor';
  const subtitle = isEdit
    ? 'Update the configuration for this monitor.'
    : 'Configure a new monitor to track the availability of your service.';

  async function syncTags(monitorId: string) {
    const toAdd = tagIds.filter(id => !originalTagIds.includes(id));
    const toRemove = originalTagIds.filter(id => !tagIds.includes(id));

    const promises: Promise<void>[] = [];
    for (const tagId of toAdd) {
      promises.push(
        new Promise<void>((resolve, reject) => {
          addTag.mutate({ monitorId, tagId }, { onSuccess: () => resolve(), onError: reject });
        }),
      );
    }
    for (const tagId of toRemove) {
      promises.push(
        new Promise<void>((resolve, reject) => {
          removeTag.mutate({ monitorId, tagId }, { onSuccess: () => resolve(), onError: reject });
        }),
      );
    }
    await Promise.all(promises);
  }

  async function handleSubmit(values: MonitorInput) {
    setSaving(true);
    const input = { ...values, parentId: values.parentId ?? null } as MonitorInput;

    try {
      if (isEdit) {
        await new Promise<void>((resolve, reject) => {
          updateMutation.mutate(
            { id: id!, input },
            { onSuccess: () => resolve(), onError: reject },
          );
        });
        await syncTags(id!);
        message.success('Monitor updated');
        navigate(`/dashboard/${id}`);
      } else {
        const created = await new Promise<{ id: string }>((resolve, reject) => {
          createMutation.mutate(input, {
            onSuccess: (data) => resolve(data as { id: string }),
            onError: reject,
          });
        });
        await syncTags(created.id);
        message.success('Monitor created');
        navigate(`/dashboard/${created.id}`);
      }
    } catch {
      message.error('Failed to save monitor. Check your inputs and try again.');
    } finally {
      setSaving(false);
    }
  }

  const isPending = saving;

  return (
    <div style={{ maxWidth: 680 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginBottom: 4 }}>{title}</Title>
        <Text type="secondary">{subtitle}</Text>
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{ type: 'http', active: true, method: 'GET', interval: 60, maxRetries: 0, timeout: 48, retryInterval: 60, maxRedirects: 10, resendInterval: 0, packetSize: 56 }}
      >
        <Card size="small" style={{ marginBottom: 16 }}>
          <GeneralFields excludeId={isEdit ? id : undefined} />
        </Card>

        <Card title="Tags" size="small" style={{ marginBottom: 16 }}>
          <TagSelector value={tagIds} onChange={setTagIds} />
        </Card>

        {type && type !== 'group' && (
          <Card title="Connection" size="small" style={{ marginBottom: 16 }}>
            <TypeFields type={type} />
            <TimingFields />
          </Card>
        )}

        {type === 'group' && (
          <Card title="Group Settings" size="small" style={{ marginBottom: 16 }}>
            <GroupFields />
          </Card>
        )}

        <Card title="Notifications" size="small" style={{ marginBottom: 16 }}>
          <NotificationSelector />
          {type !== 'group' && <AlertFields />}
        </Card>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              {isEdit ? 'Save Changes' : 'Create Monitor'}
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

function TypeFields({ type }: { type: string }) {
  if (httpTypes.includes(type)) return <HttpFields />;
  if (type === 'port') return <TcpFields />;
  if (type === 'ping' || type === 'tailscale-ping') return <PingFields />;
  if (type === 'dns') return <DnsFields />;
  if (type === 'smtp') return <SmtpFields />;
  if (type === 'mqtt') return <MqttFields />;
  if (type === 'grpc-keyword') return <GrpcFields />;
  if (type === 'redis') return <DatabaseFields />;
  if (type === 'push') return <PushFields />;
  return null;
}
