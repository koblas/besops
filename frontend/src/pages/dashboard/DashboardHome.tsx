import { Row, Col, Statistic, Card, Empty, Button, Result } from 'antd';
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  ClockCircleOutlined,
  PlusOutlined,
  DashboardOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { STATUS_COLORS, STATUS } from '../../lib/constants';
import { useMonitors } from '../../hooks/useMonitors';

export function DashboardHome() {
  const { data: monitors = [], isError } = useMonitors();
  const navigate = useNavigate();

  const active = monitors.filter(m => m.active).length;
  const paused = monitors.filter(m => !m.active).length;

  if (isError) {
    return <Result status="error" title="Failed to load monitors" subTitle="Check your connection and try again." />;
  }

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

      {monitors.length === 0 && (
        <Empty description="No monitors yet. Add one to start tracking uptime." style={{ marginTop: 48 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/add')}>
            Add Monitor
          </Button>
        </Empty>
      )}
    </div>
  );
}
