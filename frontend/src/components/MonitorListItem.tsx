import { useContext } from 'react';
import { Typography } from 'antd';
import { FolderOutlined } from '@ant-design/icons';
import { STATUS_COLORS, STATUS } from '../lib/constants';
import { ThemeContext } from '../contexts/ThemeContext';
import type { Monitor } from '../hooks/useMonitors';
import type { UptimeMap } from '../hooks/useUptimes';

interface MonitorListItemProps {
  monitor: Monitor;
  active: boolean;
  uptimes: UptimeMap;
  onClick: (id: string) => void;
}

export function MonitorListItem({
  monitor,
  active,
  uptimes,
  onClick,
}: MonitorListItemProps) {
  const { isDark } = useContext(ThemeContext);
  const isGroup = monitor.type === 'group';

  const statusColor = monitor.active ? STATUS_COLORS[STATUS.UP] : STATUS_COLORS[STATUS.PENDING];
  const uptime24 = uptimes[monitor.id];

  const hoverBg = isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.03)';
  const activeBg = isDark ? 'rgba(52, 211, 153, 0.1)' : 'rgba(52, 211, 153, 0.08)';

  return (
    <div
      role="button"
      tabIndex={0}
      aria-label={monitor.name}
      onClick={() => onClick(monitor.id)}
      onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onClick(monitor.id); } }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '7px 12px',
        borderRadius: 8,
        cursor: 'pointer',
        background: active ? activeBg : 'transparent',
        transition: 'background 0.15s cubic-bezier(0.4, 0, 0.2, 1)',
      }}
      onMouseEnter={e => {
        if (!active) e.currentTarget.style.background = hoverBg;
      }}
      onMouseLeave={e => {
        if (!active) e.currentTarget.style.background = 'transparent';
      }}
    >
      <UptimePill uptime={uptime24} color={statusColor} />
      <Typography.Text ellipsis style={{ flex: 1, fontWeight: active ? 600 : 400, fontSize: 13 }}>
        {monitor.name}
      </Typography.Text>
      {isGroup && (
        <FolderOutlined style={{ fontSize: 12, color: isDark ? 'rgba(255,255,255,0.45)' : 'rgba(0,0,0,0.35)' }} />
      )}
    </div>
  );
}

function UptimePill({ uptime, color }: { uptime: number | undefined; color: string }) {
  if (uptime === undefined) {
    return (
      <span style={{
        display: 'inline-block',
        minWidth: 44,
        padding: '1px 6px',
        borderRadius: 10,
        fontSize: 11,
        fontWeight: 600,
        textAlign: 'center',
        background: 'rgba(128,128,128,0.15)',
        color: 'rgba(128,128,128,0.7)',
      }}>
        —
      </span>
    );
  }

  const pct = Math.round(uptime * 10000) / 100;
  const display = pct > 100 ? '100%' : pct.toFixed(pct >= 99.9 ? 1 : 0) + '%';

  return (
    <span style={{
      display: 'inline-block',
      minWidth: 44,
      padding: '1px 6px',
      borderRadius: 10,
      fontSize: 11,
      fontWeight: 600,
      textAlign: 'center',
      background: color,
      color: '#fff',
    }}>
      {display}
    </span>
  );
}
