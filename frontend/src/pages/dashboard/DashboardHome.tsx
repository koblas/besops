import { Row, Col, Statistic, Card, Empty, Button, Result, Table, Typography } from 'antd';
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  ClockCircleOutlined,
  PlusOutlined,
  DashboardOutlined,
} from '@ant-design/icons';
import { useNavigate, Link } from 'react-router-dom';
import { STATUS_COLORS, STATUS, type StatusValue } from '../../lib/constants';
import { useMonitors } from '../../hooks/useMonitors';
import { useRecentEvents } from '../../hooks/useHeartbeats';
import { StatusBadge } from '../../components/StatusBadge';
import { formatRelative } from '../../lib/formatters';

const { Text } = Typography;

export function DashboardHome() {
  const { data: monitors = [], isError } = useMonitors();
  const { data: eventsResp, isLoading: eventsLoading } = useRecentEvents();
  const events = eventsResp?.data ?? [];
  const navigate = useNavigate();

  const active = monitors.filter(m => m.active).length;
  const paused = monitors.filter(m => !m.active).length;

  const monitorMap = new Map(monitors.map(m => [m.id, m]));

  if (isError) {
    return <Result status="error" title="Failed to load monitors" subTitle="Check your connection and try again." />;
  }

  const eventColumns = [
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: StatusValue) => <StatusBadge status={status} />,
    },
    {
      title: 'Monitor',
      dataIndex: 'monitorId',
      key: 'monitor',
      render: (monitorId: string) => {
        const mon = monitorMap.get(monitorId);
        if (!mon) return <Text type="secondary">Unknown</Text>;
        return <Link to={`/dashboard/${monitorId}`}>{mon.name}</Link>;
      },
    },
    {
      title: 'Message',
      dataIndex: 'msg',
      key: 'msg',
      ellipsis: true,
    },
    {
      title: 'Time',
      dataIndex: 'time',
      key: 'time',
      width: 140,
      render: (time: string) => <Text type="secondary">{formatRelative(time)}</Text>,
    },
  ];

  return (
    <div>
      <Row gutter={[12, 12]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="Active"
              value={active}
              styles={{ content: { color: STATUS_COLORS[STATUS.UP], fontSize: 28, fontWeight: 600 } }}
              prefix={<ArrowUpOutlined style={{ fontSize: 16 }} />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="Down"
              value={0}
              styles={{ content: { color: STATUS_COLORS[STATUS.DOWN], fontSize: 28, fontWeight: 600 } }}
              prefix={<ArrowDownOutlined style={{ fontSize: 16 }} />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="Paused"
              value={paused}
              styles={{ content: { color: STATUS_COLORS[STATUS.PENDING], fontSize: 28, fontWeight: 600 } }}
              prefix={<ClockCircleOutlined style={{ fontSize: 16 }} />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="Total"
              value={monitors.length}
              styles={{ content: { fontSize: 28, fontWeight: 600 } }}
              prefix={<DashboardOutlined style={{ fontSize: 16 }} />}
            />
          </Card>
        </Col>
      </Row>

      {monitors.length === 0 ? (
        <Empty description="No monitors yet. Add one to start tracking uptime." style={{ marginTop: 48 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/add')}>
            Add Monitor
          </Button>
        </Empty>
      ) : (
        <Card title="Recent Events" size="small">
          <Table
            dataSource={events}
            columns={eventColumns}
            rowKey="id"
            size="small"
            loading={eventsLoading}
            pagination={false}
            locale={{ emptyText: 'No events yet' }}
          />
        </Card>
      )}
    </div>
  );
}
