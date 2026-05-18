import { useParams, useNavigate } from 'react-router-dom';
import { Typography, Spin, Button, Space, Modal, Row, Col, Card, Table, Result, message } from 'antd';
import {
  EditOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  DeleteOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import {
  useMonitor,
  usePauseMonitor,
  useResumeMonitor,
  useDeleteMonitor,
} from '../../hooks/useMonitors';
import { useHeartbeats, useImportantHeartbeats } from '../../hooks/useHeartbeats';
import { HeartbeatBar } from '../../components/HeartbeatBar';
import { LatencyChart } from '../../components/LatencyChart';
import { StatusBadge } from '../../components/StatusBadge';
import { UptimeBadge } from '../../components/UptimeBadge';
import { formatDateTime, formatLatency } from '../../lib/formatters';
import type { StatusValue } from '../../lib/constants';

const { Title, Text } = Typography;

export function MonitorDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: monitor, isLoading, isError } = useMonitor(id);
  const { data: heartbeats = [] } = useHeartbeats(id);
  const { data: eventsResp } = useImportantHeartbeats(id);
  const events = eventsResp?.data ?? [];
  const pauseMutation = usePauseMonitor();
  const resumeMutation = useResumeMonitor();
  const deleteMutation = useDeleteMonitor();

  if (!id || isLoading) {
    return (
      <div style={{ textAlign: 'center', marginTop: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (isError) {
    return <Result status="error" title="Failed to load monitor" subTitle="Check your connection and try again." extra={<Button onClick={() => navigate('/dashboard')}>Back to Dashboard</Button>} />;
  }

  if (!monitor) {
    return (
      <div style={{ textAlign: 'center', marginTop: 48 }}>
        <Text type="secondary">Monitor not found.</Text>
        <div style={{ marginTop: 16 }}>
          <Button type="primary" onClick={() => navigate('/dashboard')}>Back to Dashboard</Button>
        </div>
      </div>
    );
  }

  const handlePause = () => {
    pauseMutation.mutate(id, {
      onSuccess: () => message.success('Monitor paused'),
      onError: () => message.error('Failed to pause monitor'),
    });
  };

  const handleResume = () => {
    resumeMutation.mutate(id, {
      onSuccess: () => message.success('Monitor resumed'),
      onError: () => message.error('Failed to resume monitor'),
    });
  };

  const handleDelete = () => {
    Modal.confirm({
      title: 'Delete Monitor',
      content: `Are you sure you want to delete "${monitor.name}"? This cannot be undone.`,
      okText: 'Delete',
      okType: 'danger',
      onOk: () => {
        deleteMutation.mutate(id, {
          onSuccess: () => {
            message.success('Monitor deleted');
            navigate('/dashboard');
          },
          onError: () => message.error('Failed to delete monitor'),
        });
      },
    });
  };

  const lastBeat = heartbeats.length > 0 ? heartbeats[heartbeats.length - 1] : null;
  const currentStatus = lastBeat?.status ?? (monitor.active ? 1 : 2);

  const eventColumns = [
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: StatusValue) => <StatusBadge status={status} />,
    },
    {
      title: 'Time',
      dataIndex: 'time',
      key: 'time',
      render: (time: string) => formatDateTime(time),
    },
    {
      title: 'Message',
      dataIndex: 'msg',
      key: 'msg',
      ellipsis: true,
    },
    {
      title: 'Latency',
      dataIndex: 'latency',
      key: 'latency',
      width: 100,
      render: (latency: number | null) => formatLatency(latency),
    },
  ];

  return (
    <div>
      {/* Header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 16,
        }}
      >
        <Space align="center">
          <StatusBadge status={currentStatus as StatusValue} />
          <Title level={4} style={{ margin: 0 }}>
            {monitor.name}
          </Title>
        </Space>
        <Space>
          <Button icon={<EditOutlined />} onClick={() => navigate(`/edit/${id}`)}>
            Edit
          </Button>
          <Button icon={<CopyOutlined />} onClick={() => navigate(`/clone/${id}`)}>
            Clone
          </Button>
          {monitor.active ? (
            <Button
              icon={<PauseCircleOutlined />}
              onClick={handlePause}
              loading={pauseMutation.isPending}
            >
              Pause
            </Button>
          ) : (
            <Button
              icon={<PlayCircleOutlined />}
              onClick={handleResume}
              loading={resumeMutation.isPending}
              type="primary"
            >
              Resume
            </Button>
          )}
          <Button
            icon={<DeleteOutlined />}
            danger
            onClick={handleDelete}
            loading={deleteMutation.isPending}
          >
            Delete
          </Button>
        </Space>
      </div>

      {/* Heartbeat Bar */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <HeartbeatBar heartbeats={heartbeats} />
      </Card>

      {/* Stats Row */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card size="small">
            <UptimeBadge heartbeats={heartbeats} label="Uptime (24h)" />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" style={{ textAlign: 'center' }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Avg Latency
            </Text>
            <div>
              <Text strong style={{ fontSize: 16 }}>
                {formatLatency(
                  heartbeats.length > 0
                    ? heartbeats.reduce((sum, hb) => sum + (hb.latency ?? 0), 0) /
                        heartbeats.filter(hb => hb.latency != null).length
                    : null,
                )}
              </Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" style={{ textAlign: 'center' }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Type
            </Text>
            <div>
              <Text strong style={{ fontSize: 16 }}>
                {monitor.type}
              </Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" style={{ textAlign: 'center' }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Interval
            </Text>
            <div>
              <Text strong style={{ fontSize: 16 }}>
                {monitor.interval}s
              </Text>
            </div>
          </Card>
        </Col>
      </Row>

      {/* Latency Chart */}
      <Card title="Response Time" size="small" style={{ marginBottom: 16 }}>
        <LatencyChart monitorId={id} heartbeats={heartbeats} />
      </Card>

      {/* Events Table */}
      <Card title="Events" size="small">
        <Table
          dataSource={events}
          columns={eventColumns}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  );
}
