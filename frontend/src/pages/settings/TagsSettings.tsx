import { useState } from 'react';
import { Typography, Button, Table, Space, ColorPicker, Modal, Form, Input, Empty, Tooltip, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTags, useCreateTag, useUpdateTag, useDeleteTag } from '../../hooks/useTags';
import type { Tag, TagInput } from '../../hooks/useTags';
import { TagBadge } from '../../components/TagBadge';

const { Title } = Typography;

export function TagsSettings() {
  const { data: tags = [], isLoading } = useTags();
  const createMutation = useCreateTag();
  const updateMutation = useUpdateTag();
  const deleteMutation = useDeleteTag();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Tag | null>(null);
  const [form] = Form.useForm();

  function openCreate() {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ color: '#2196F3' });
    setModalOpen(true);
  }

  function openEdit(record: Tag) {
    setEditing(record);
    form.setFieldsValue({ name: record.name, color: record.color });
    setModalOpen(true);
  }

  function handleDelete(record: Tag) {
    Modal.confirm({
      title: 'Delete Tag',
      content: `Delete "${record.name}"? It will be removed from all monitors.`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.id).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  async function handleSubmit() {
    const values = await form.validateFields();
    const color = typeof values.color === 'string' ? values.color : values.color?.toHexString?.() ?? '#2196F3';
    const input: TagInput = { name: values.name, color };

    if (editing) {
      updateMutation.mutate(
        { id: editing.id, input },
        {
          onSuccess: () => { message.success('Updated'); setModalOpen(false); },
          onError: () => message.error('Failed to save tag'),
        },
      );
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Created'); setModalOpen(false); },
        onError: () => message.error('Failed to create tag'),
      });
    }
  }

  const columns = [
    {
      title: 'Preview',
      key: 'preview',
      width: 150,
      render: (_: unknown, record: Tag) => <TagBadge name={record.name} color={record.color} />,
    },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'Color',
      dataIndex: 'color',
      key: 'color',
      width: 100,
      render: (color: string) => (
        <div style={{ width: 24, height: 24, borderRadius: 4, background: color }} />
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: unknown, record: Tag) => (
        <Space size="small">
          <Tooltip title="Edit"><Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)} /></Tooltip>
          <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Tags</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Add Tag
        </Button>
      </div>

      <Table
        dataSource={tags}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={false}
        locale={{ emptyText: <Empty description="No tags yet. Tags help you organize and filter monitors." /> }}
      />

      <Modal
        open={modalOpen}
        title={editing ? 'Edit Tag' : 'New Tag'}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="production" />
          </Form.Item>
          <Form.Item name="color" label="Color">
            <ColorPicker showText />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
