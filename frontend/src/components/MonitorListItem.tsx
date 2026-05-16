import { useContext, useState } from 'react';
import { Typography } from 'antd';
import { RightOutlined } from '@ant-design/icons';
import { STATUS_COLORS, STATUS } from '../lib/constants';
import { ThemeContext } from '../contexts/ThemeContext';
import type { Monitor } from '../hooks/useMonitors';
import type { UptimeMap } from '../hooks/useUptimes';

interface MonitorListItemProps {
  monitor: Monitor;
  active: boolean;
  activeId?: string;
  depth: number;
  childrenMap: Map<string, Monitor[]>;
  matchesSearch: Set<string>;
  uptimes: UptimeMap;
  onClick: (id: string) => void;
}

export function MonitorListItem({
  monitor,
  active,
  activeId,
  depth,
  childrenMap,
  matchesSearch,
  uptimes,
  onClick,
}: MonitorListItemProps) {
  const [collapsed, setCollapsed] = useState(false);
  const { isDark } = useContext(ThemeContext);

  const children = (childrenMap.get(monitor.id) ?? []).filter(m => matchesSearch.has(m.id));
  const hasChildren = children.length > 0;
  const isGroup = monitor.type === 'group';

  const statusColor = monitor.active ? STATUS_COLORS[STATUS.UP] : STATUS_COLORS[STATUS.PENDING];
  const uptime24 = uptimes[monitor.id];

  const hoverBg = isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.03)';
  const activeBg = isDark ? 'rgba(52, 211, 153, 0.1)' : 'rgba(52, 211, 153, 0.08)';

  return (
    <div>
      <div
        role="button"
        tabIndex={0}
        aria-label={monitor.name}
        aria-expanded={(isGroup || hasChildren) ? !collapsed : undefined}
        onClick={() => onClick(monitor.id)}
        onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onClick(monitor.id); } }}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '7px 12px',
          paddingLeft: 12 + depth * 20,
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
        {(isGroup || hasChildren) && (
          <span
            role="button"
            tabIndex={0}
            aria-label={collapsed ? 'Expand group' : 'Collapse group'}
            onClick={e => {
              e.stopPropagation();
              setCollapsed(!collapsed);
            }}
            onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.stopPropagation(); e.preventDefault(); setCollapsed(!collapsed); } }}
            style={{
              cursor: 'pointer',
              fontSize: 9,
              width: 14,
              textAlign: 'center',
              color: isDark ? 'rgba(255,255,255,0.35)' : 'rgba(0,0,0,0.3)',
              transition: 'transform 0.15s',
              transform: collapsed ? 'rotate(0deg)' : 'rotate(90deg)',
            }}
          >
            <RightOutlined />
          </span>
        )}
        {!isGroup && !hasChildren && <span style={{ width: 14 }} />}
        <UptimePill uptime={uptime24} color={statusColor} />
        <Typography.Text ellipsis style={{ flex: 1, fontWeight: active ? 600 : 400, fontSize: 13 }}>
          {monitor.name}
        </Typography.Text>
      </div>
      {hasChildren && !collapsed && (
        <div>
          {children.map(child => (
            <MonitorListItem
              key={child.id}
              monitor={child}
              active={child.id === activeId}
              activeId={activeId}
              depth={depth + 1}
              childrenMap={childrenMap}
              matchesSearch={matchesSearch}
              uptimes={uptimes}
              onClick={onClick}
            />
          ))}
        </div>
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
