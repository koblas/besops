import { Typography } from 'antd';
import { STATUS_COLORS, STATUS } from '../lib/constants';
import type { Heartbeat } from '../hooks/useHeartbeats';

interface UptimeBadgeProps {
  heartbeats: Heartbeat[];
  label?: string;
}

export function UptimeBadge({ heartbeats, label }: UptimeBadgeProps) {
  const total = heartbeats.length;
  const up = heartbeats.filter(hb => hb.status === STATUS.UP).length;
  const percentage = total > 0 ? (up / total) * 100 : 0;
  const color =
    percentage >= 99
      ? STATUS_COLORS[STATUS.UP]
      : percentage >= 95
        ? '#f0ad4e'
        : STATUS_COLORS[STATUS.DOWN];

  return (
    <div style={{ textAlign: 'center' }}>
      {label && (
        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
          {label}
        </Typography.Text>
      )}
      <div>
        <Typography.Text strong style={{ color, fontSize: 16 }}>
          {percentage.toFixed(2)}%
        </Typography.Text>
      </div>
    </div>
  );
}
