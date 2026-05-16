import { useParams, Link } from 'react-router-dom';
import { Typography, Spin, Card, Tag, Space, Alert, Result, Button, Tooltip } from 'antd';
import { EditOutlined, DashboardOutlined } from '@ant-design/icons';
import { useStatusPage, useStatusPageHeartbeats } from '../../hooks/useStatusPages';
import { useIncidents } from '../../hooks/useIncidents';
import { useAuth } from '../../hooks/useAuth';
import { HeartbeatBar } from '../../components/HeartbeatBar';
import { StatusBadge } from '../../components/StatusBadge';
import type { StatusValue } from '../../lib/constants';
import type { components } from '../../api/generated/v1';

const { Title, Text, Paragraph } = Typography;

type Heartbeat = components['schemas']['Heartbeat'];

export function StatusPagePublic() {
  const { slug } = useParams<{ slug: string }>();
  const { data: page, isLoading, isError } = useStatusPage(slug);
  const { data: heartbeatData } = useStatusPageHeartbeats(slug);
  const { data: incidentData } = useIncidents(slug);
  const incidents = incidentData?.incidents ?? [];
  const { isAuthenticated } = useAuth();

  const heartbeatMap = heartbeatData?.heartbeatList ?? {};
  const uptimeMap = heartbeatData?.uptimeList ?? {};
  const monitorNames = heartbeatData?.monitorNames ?? {};

  if (isLoading) {
    return (
      <div style={{ maxWidth: 860, margin: '0 auto', padding: 48, textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (isError || !page) {
    return (
      <div style={{ maxWidth: 860, margin: '0 auto', padding: 48 }}>
        <Result status="404" title="Status Page Not Found" subTitle="This status page doesn't exist or isn't published." />
      </div>
    );
  }

  const pinnedIncidents = incidents.filter(i => i.pin && !i.resolved);
  const resolvedIncidents = incidents.filter(i => i.resolved).slice(0, 5);

  const overallStatus = computeOverallStatus(page.groups, heartbeatMap);

  return (
    <div style={{ maxWidth: 860, margin: '0 auto', padding: '32px 24px' }}>
      {page.customCss && <style>{page.customCss}</style>}

      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        {page.icon && <img src={page.icon} alt="" style={{ width: 48, height: 48, marginBottom: 12 }} />}
        <Title level={2} style={{ marginBottom: 4 }}>{page.title}</Title>
        {page.description && <Paragraph type="secondary">{page.description}</Paragraph>}

        {isAuthenticated && (
          <Space style={{ marginTop: 12 }}>
            <Link to="/dashboard">
              <Button icon={<DashboardOutlined />}>Dashboard</Button>
            </Link>
            <Link to={`/manage-status-page`}>
              <Button icon={<EditOutlined />}>Edit Status Page</Button>
            </Link>
          </Space>
        )}
      </div>

      {/* Overall status banner */}
      <Card
        size="small"
        style={{ marginBottom: 24, textAlign: 'center' }}
        styles={{ body: { padding: '16px 24px' } }}
      >
        <Space>
          <StatusBadge status={overallStatus} />
          <Text strong style={{ fontSize: 16 }}>
            {overallStatus === 1 ? 'All Systems Operational' : overallStatus === 0 ? 'System Outage Detected' : 'Checking Status...'}
          </Text>
        </Space>
      </Card>

      {/* Pinned incidents */}
      {pinnedIncidents.map(incident => (
        <Alert
          key={incident.id}
          type={incident.style === 'danger' ? 'error' : incident.style === 'warning' ? 'warning' : 'info'}
          message={incident.title}
          description={incident.content}
          showIcon
          style={{ marginBottom: 12 }}
        />
      ))}

      {/* Monitor groups */}
      {page.groups?.map(group => (
        <Card
          key={group.id ?? group.name}
          title={group.name}
          size="small"
          style={{ marginBottom: 16 }}
        >
          {group.monitorIds?.length ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {group.monitorIds.map(monitorId => {
                const beats = (heartbeatMap[monitorId] ?? []) as Heartbeat[];
                const lastBeat = beats.length > 0 ? beats[beats.length - 1] : null;
                const status = lastBeat?.status ?? 2;
                const name = monitorNames[monitorId] ?? 'Service';
                const uptimeKey = `${monitorId}_24`;
                const uptime = uptimeMap[uptimeKey];
                return (
                  <MonitorRow key={monitorId} name={name} status={status} beats={beats} uptime={uptime} />
                );
              })}
            </div>
          ) : (
            <Text type="secondary">No monitors in this group.</Text>
          )}
        </Card>
      ))}

      {(!page.groups || page.groups.length === 0) && (
        <Card>
          <Text type="secondary">No monitor groups configured for this status page.</Text>
        </Card>
      )}

      {/* Resolved incidents history */}
      {resolvedIncidents.length > 0 && (
        <Card title="Recent Incidents" size="small" style={{ marginTop: 24 }}>
          <Space direction="vertical" style={{ width: '100%' }}>
            {resolvedIncidents.map(incident => (
              <div key={incident.id} style={{ paddingBottom: 8, borderBottom: '1px solid var(--ant-color-border)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Text strong>{incident.title}</Text>
                  <Tag color="green">Resolved</Tag>
                </div>
                {incident.content && <Text type="secondary" style={{ fontSize: 13 }}>{incident.content}</Text>}
                {incident.lastUpdatedDate && (
                  <div><Text type="secondary" style={{ fontSize: 12 }}>{new Date(incident.lastUpdatedDate).toLocaleDateString()}</Text></div>
                )}
              </div>
            ))}
          </Space>
        </Card>
      )}

      {/* Footer */}
      <div style={{ textAlign: 'center', marginTop: 32, paddingTop: 16 }}>
        {page.footerText && <Paragraph type="secondary" style={{ fontSize: 13 }}>{page.footerText}</Paragraph>}
        {page.showPoweredBy !== false && (
          <Text type="secondary" style={{ fontSize: 12 }}>Powered by Bes Ops</Text>
        )}
      </div>
    </div>
  );
}

function MonitorRow({ name, status, beats, uptime }: { name: string; status: number; beats: Heartbeat[]; uptime?: number }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
      <StatusBadge status={status as StatusValue} />
      <Text strong style={{ minWidth: 140, flex: '0 0 auto' }}>{name}</Text>
      <div style={{ flex: 1, minWidth: 0 }}>
        <HeartbeatBar heartbeats={beats} size="small" />
      </div>
      {uptime !== undefined && (
        <Tooltip title="24-hour uptime">
          <Text type="secondary" style={{ minWidth: 55, textAlign: 'right', fontSize: 13 }}>
            {(uptime * 100).toFixed(1)}%
          </Text>
        </Tooltip>
      )}
      <Tag color={status === 1 ? 'green' : status === 0 ? 'red' : 'default'} style={{ margin: 0 }}>
        {status === 1 ? 'Up' : status === 0 ? 'Down' : 'Pending'}
      </Tag>
    </div>
  );
}

function computeOverallStatus(
  groups: components['schemas']['StatusPage']['groups'],
  heartbeatMap: Record<string, Heartbeat[]>,
): StatusValue {
  if (!groups || groups.length === 0) return 2 as StatusValue;

  let hasDown = false;
  let hasUp = false;

  for (const group of groups) {
    for (const mid of group.monitorIds ?? []) {
      const beats = heartbeatMap[mid] ?? [];
      if (beats.length === 0) continue;
      const last = beats[beats.length - 1];
      if (last.status === 1) hasUp = true;
      if (last.status === 0) hasDown = true;
    }
  }

  if (hasDown) return 0 as StatusValue;
  if (hasUp) return 1 as StatusValue;
  return 2 as StatusValue;
}
