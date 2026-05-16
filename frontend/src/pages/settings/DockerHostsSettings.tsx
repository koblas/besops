import { useState } from 'react';
import { Typography, Button, Table, Space, Modal, Form, Input, Select, Empty, Tooltip, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ApiOutlined } from '@ant-design/icons';
import {
  useDockerHosts,
  useCreateDockerHost,
  useUpdateDockerHost,
  useDeleteDockerHost,
  useTestDockerHost,
} from '../../hooks/useDockerHosts';
import type { DockerHost, DockerHostInput } from '../../hooks/useDockerHosts';

const { Title } = Typography;

export function DockerHostsSettings() {
  const { data: hosts = [], isLoading } = useDockerHosts();
  const createMutation = useCreateDockerHost();
  const updateMutation = useUpdateDockerHost();
  const deleteMutation = useDeleteDockerHost();
  const testMutation = useTestDockerHost();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<DockerHost | null>(null);
  const [form] = Form.useForm();

  function openCreate() {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ dockerType: 'socket', dockerDaemon: '/var/run/docker.sock' });
    setModalOpen(true);
  }

  function openEdit(record: DockerHost) {
    setEditing(record);
    form.setFieldsValue(record);
    setModalOpen(true);
  }

  function handleDelete(record: DockerHost) {
    Modal.confirm({
      title: 'Delete Docker Host',
      content: `Delete "${record.name}"?`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => deleteMutation.mutateAsync(record.id).then(() => message.success('Deleted')).catch(() => message.error('Failed to delete')),
    });
  }

  function handleTest(id: string) {
    testMutation.mutate(id, {
      onSuccess: () => message.success('Connection successful'),
      onError: () => message.error('Connection failed'),
    });
  }

  async function handleSubmit() {
    const values = await form.validateFields();
    const input: DockerHostInput = {
      name: values.name,
      dockerType: values.dockerType,
      dockerDaemon: values.dockerDaemon,
    };

    if (editing) {
      updateMutation.mutate({ id: editing.id, input }, {
        onSuccess: () => { message.success('Updated'); setModalOpen(false); },
        onError: () => message.error('Failed to save Docker host'),
      });
    } else {
      createMutation.mutate(input, {
        onSuccess: () => { message.success('Created'); setModalOpen(false); },
        onError: () => message.error('Failed to create Docker host'),
      });
    }
  }

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Type', dataIndex: 'dockerType', key: 'dockerType', width: 100 },
    { title: 'Daemon', dataIndex: 'dockerDaemon', key: 'dockerDaemon' },
    {
      title: 'Actions',
      key: 'actions',
      width: 180,
      render: (_: unknown, record: DockerHost) => (
        <Space size="small">
          <Button size="small" icon={<ApiOutlined />} onClick={() => handleTest(record.id)} loading={testMutation.isPending}>
            Test
          </Button>
          <Tooltip title="Edit"><Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)} /></Tooltip>
          <Tooltip title="Delete"><Button size="small" icon={<DeleteOutlined />} danger onClick={() => handleDelete(record)} /></Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Docker Hosts</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Add Docker Host
        </Button>
      </div>

      <Table
        dataSource={hosts}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={false}
        locale={{ emptyText: <Empty description="No Docker hosts configured. Add one to monitor Docker containers." /> }}
      />

      <Modal
        open={modalOpen}
        title={editing ? 'Edit Docker Host' : 'New Docker Host'}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="Local Docker" />
          </Form.Item>
          <Form.Item name="dockerType" label="Connection Type" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'socket', label: 'Socket' },
                { value: 'tcp', label: 'TCP' },
              ]}
            />
          </Form.Item>
          <Form.Item name="dockerDaemon" label="Docker Daemon">
            <Input placeholder="/var/run/docker.sock" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
