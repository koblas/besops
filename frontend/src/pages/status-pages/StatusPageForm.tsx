import { useEffect, useRef, useState } from 'react';
import { Typography, Form, Input, Select, Switch, Button, Card, Space, Spin, message, Tooltip, Popconfirm, Alert, Result } from 'antd';
import { PlusOutlined, DeleteOutlined, ArrowUpOutlined, ArrowDownOutlined, EyeOutlined, LinkOutlined } from '@ant-design/icons';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { useStatusPage, useCreateStatusPage, useUpdateStatusPage } from '../../hooks/useStatusPages';
import type { StatusPageInput, StatusPageGroup } from '../../hooks/useStatusPages';
import { useMonitors } from '../../hooks/useMonitors';
import { useTags } from '../../hooks/useTags';

const { Title, Text } = Typography;
const { TextArea } = Input;

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 50);
}

export function StatusPageForm() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const slug = searchParams.get('slug');
  const isEdit = !!slug;

  const { data: existing, isLoading, isError } = useStatusPage(slug ?? undefined);
  const { data: monitors = [] } = useMonitors();
  const { data: tags = [] } = useTags();
  const createMutation = useCreateStatusPage();
  const updateMutation = useUpdateStatusPage();
  const [form] = Form.useForm();
  const [groups, setGroups] = useState<StatusPageGroup[]>([]);
  const didInit = useRef(false);
  const [slugTouched, setSlugTouched] = useState(false);

  const currentSlug = Form.useWatch('slug', form);

  useEffect(() => {
    if (existing && !didInit.current) {
      didInit.current = true;
      setSlugTouched(true);
      form.setFieldsValue({
        title: existing.title,
        slug: existing.slug,
        description: existing.description,
        theme: existing.theme,
        published: existing.published,
        showTags: existing.showTags,
        showPoweredBy: existing.showPoweredBy,
        showCertificateExpiry: existing.showCertificateExpiry,
        customCss: existing.customCss,
        footerText: existing.footerText,
        googleAnalyticsId: existing.googleAnalyticsId,
      });
      setGroups(existing.groups ?? []);
    }
  }, [existing, form]);

  if (slug && isLoading) {
    return <div style={{ padding: 48, textAlign: 'center' }}><Spin size="large" /></div>;
  }

  if (slug && isError) {
    return (
      <div style={{ padding: 24, maxWidth: 700 }}>
        <Result
          status="error"
          title="Could not load status page"
          subTitle="The page may have been deleted, or there was a network error."
          extra={<Button type="primary" onClick={() => navigate('/manage-status-page')}>Back to Status Pages</Button>}
        />
      </div>
    );
  }

  function addGroup() {
    setGroups([...groups, { name: `Group ${groups.length + 1}`, weight: groups.length, monitorIds: [] }]);
  }

  function removeGroup(idx: number) {
    setGroups(groups.filter((_, i) => i !== idx));
  }

  function moveGroup(idx: number, dir: -1 | 1) {
    const arr = [...groups];
    const target = idx + dir;
    if (target < 0 || target >= arr.length) return;
    [arr[idx], arr[target]] = [arr[target], arr[idx]];
    setGroups(arr);
  }

  function updateGroup(idx: number, patch: Partial<StatusPageGroup>) {
    setGroups(groups.map((g, i) => (i === idx ? { ...g, ...patch } : g)));
  }

  function handleTitleChange(e: React.ChangeEvent<HTMLInputElement>) {
    if (!isEdit && !slugTouched) {
      form.setFieldValue('slug', slugify(e.target.value));
    }
  }

  async function handleSubmit(values: Record<string, unknown>) {
    const input: StatusPageInput = {
      title: values.title as string,
      slug: values.slug as string,
      description: values.description as string | undefined,
      theme: values.theme as StatusPageInput['theme'],
      published: (values.published as boolean) ?? true,
      showTags: (values.showTags as boolean) ?? false,
      showPoweredBy: (values.showPoweredBy as boolean) ?? true,
      showCertificateExpiry: (values.showCertificateExpiry as boolean) ?? false,
      customCss: values.customCss as string | undefined,
      footerText: values.footerText as string | undefined,
      googleAnalyticsId: values.googleAnalyticsId as string | undefined,
      groups: groups.map((g, i) => ({ ...g, weight: i })),
    };

    if (isEdit) {
      updateMutation.mutate({ slug: slug!, input }, {
        onSuccess: () => { message.success('Status page updated'); navigate('/manage-status-page'); },
        onError: () => message.error('Could not save — check your connection and try again'),
      });
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Status page created'); navigate('/manage-status-page'); },
        onError: () => message.error('Could not create — the slug may already be in use'),
      });
    }
  }

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <div style={{ padding: 24, maxWidth: 700 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={4} style={{ margin: 0 }}>{isEdit ? 'Edit Status Page' : 'New Status Page'}</Title>
        {isEdit && slug && (
          <Link to={`/status/${slug}`} target="_blank">
            <Button icon={<EyeOutlined />}>View Live Page</Button>
          </Link>
        )}
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{ published: true, showPoweredBy: true, theme: 'auto' }}
      >
        {/* Identity */}
        <Card size="small" style={{ marginBottom: 16 }}>
          <Form.Item
            name="title"
            label="Title"
            extra="Displayed at the top of your public status page"
            rules={[{ required: true, message: 'Give your status page a name' }]}
          >
            <Input placeholder="Acme Platform Status" onChange={handleTitleChange} />
          </Form.Item>

          <Form.Item
            name="slug"
            label="URL Path"
            extra={currentSlug ? <Text type="secondary"><LinkOutlined /> Your page will be at <Text code>/status/{currentSlug}</Text></Text> : 'Letters, numbers, and hyphens only'}
            rules={[
              { required: true, message: 'Required — this becomes the URL' },
              { pattern: /^[a-z0-9-]+$/, message: 'Lowercase letters, numbers, and hyphens only' },
            ]}
          >
            <Input
              placeholder="acme-status"
              disabled={isEdit}
              onChange={() => setSlugTouched(true)}
            />
          </Form.Item>

          <Form.Item
            name="description"
            label="Description"
            extra="Optional subtitle shown below the title"
          >
            <TextArea rows={2} placeholder="Current system status and recent incidents" />
          </Form.Item>
        </Card>

        {/* Monitor Groups */}
        <Card
          title="Monitor Groups"
          size="small"
          style={{ marginBottom: 16 }}
          extra={
            <Text type="secondary" style={{ fontSize: 12 }}>
              {groups.length} {groups.length === 1 ? 'group' : 'groups'}
            </Text>
          }
        >
          {groups.length === 0 && (
            <Alert
              type="info"
              showIcon
              message="No groups yet"
              description="Groups organize your monitors on the public page. Add a group, then select which monitors appear in it."
              style={{ marginBottom: 12 }}
            />
          )}

          {groups.map((group, idx) => (
            <Card
              key={idx}
              size="small"
              style={{ marginBottom: 8 }}
              extra={
                <Space size="small">
                  <Tooltip title="Move up">
                    <Button size="small" icon={<ArrowUpOutlined />} onClick={() => moveGroup(idx, -1)} disabled={idx === 0} />
                  </Tooltip>
                  <Tooltip title="Move down">
                    <Button size="small" icon={<ArrowDownOutlined />} onClick={() => moveGroup(idx, 1)} disabled={idx === groups.length - 1} />
                  </Tooltip>
                  <Popconfirm
                    title="Remove this group?"
                    description={(group.monitorIds?.length ?? 0) > 0 ? `${group.monitorIds!.length} monitor(s) will be removed from the page` : undefined}
                    onConfirm={() => removeGroup(idx)}
                    okText="Remove"
                    okButtonProps={{ danger: true }}
                  >
                    <Tooltip title="Remove group">
                      <Button size="small" icon={<DeleteOutlined />} danger />
                    </Tooltip>
                  </Popconfirm>
                </Space>
              }
            >
              <Input
                value={group.name}
                onChange={e => updateGroup(idx, { name: e.target.value })}
                placeholder="Group name (e.g. Core Services)"
                style={{ marginBottom: 8 }}
              />
              <Select
                mode="multiple"
                placeholder="Select monitors to display in this group"
                value={group.monitorIds}
                onChange={monitorIds => updateGroup(idx, { monitorIds })}
                options={monitors.map(m => ({ value: m.id, label: m.name }))}
                style={{ width: '100%', marginBottom: 8 }}
                notFoundContent="No monitors created yet"
              />
              <Select
                mode="multiple"
                placeholder="Or include monitors by tag"
                value={group.tagIds}
                onChange={tagIds => updateGroup(idx, { tagIds })}
                options={tags.map(t => ({ value: t.id, label: t.name }))}
                style={{ width: '100%' }}
                notFoundContent="No tags created yet"
              />
            </Card>
          ))}

          <Button type="dashed" icon={<PlusOutlined />} onClick={addGroup} block>
            Add Group
          </Button>
        </Card>

        {/* Appearance */}
        <Card title="Appearance" size="small" style={{ marginBottom: 16 }}>
          <Form.Item name="theme" label="Theme" extra="Controls light/dark mode for visitors">
            <Select options={[
              { value: 'auto', label: 'Auto (match visitor\'s system)' },
              { value: 'light', label: 'Light' },
              { value: 'dark', label: 'Dark' },
            ]} />
          </Form.Item>

          <Form.Item name="published" label="Published" valuePropName="checked" extra="Unpublished pages return a 404 to visitors">
            <Switch />
          </Form.Item>

          <Form.Item name="showTags" label="Show Tags" valuePropName="checked" extra="Display monitor tags on the public page">
            <Switch />
          </Form.Item>

          <Form.Item name="showCertificateExpiry" label="Show Certificate Expiry" valuePropName="checked" extra="Display days until TLS certificate expires for each monitor">
            <Switch />
          </Form.Item>

          <Form.Item name="showPoweredBy" label='Show "Powered by Bes Ops"' valuePropName="checked" extra="Small credit line in the page footer">
            <Switch />
          </Form.Item>
        </Card>

        {/* Customization */}
        <Card title="Customization" size="small" style={{ marginBottom: 16 }}>
          <Form.Item name="footerText" label="Footer Text" extra="Custom text shown at the bottom of the page">
            <Input placeholder="© 2026 Acme Inc." />
          </Form.Item>

          <Form.Item name="customCss" label="Custom CSS" extra="Applied to the public page only — use browser DevTools to find class names">
            <TextArea rows={3} style={{ fontFamily: 'monospace', fontSize: 13 }} placeholder=".ant-card { border-radius: 12px; }" />
          </Form.Item>

          <Form.Item name="googleAnalyticsId" label="Google Analytics ID" extra="Measurement ID for tracking page views">
            <Input placeholder="G-XXXXXXXXXX" />
          </Form.Item>
        </Card>

        {/* Actions */}
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              {isEdit ? 'Save Changes' : 'Create Status Page'}
            </Button>
            <Button onClick={() => navigate('/manage-status-page')}>Cancel</Button>
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}
