export const STATUS = {
  DOWN: 0,
  UP: 1,
  PENDING: 2,
  MAINTENANCE: 3,
  DEGRADED: 4,
} as const;

export type StatusValue = (typeof STATUS)[keyof typeof STATUS];

export const STATUS_COLORS: Record<StatusValue, string> = {
  [STATUS.DOWN]: '#f87171',
  [STATUS.UP]: '#34d399',
  [STATUS.PENDING]: '#fbbf24',
  [STATUS.MAINTENANCE]: '#60a5fa',
  [STATUS.DEGRADED]: '#fb923c',
};

export const STATUS_LABELS: Record<StatusValue, string> = {
  [STATUS.DOWN]: 'Down',
  [STATUS.UP]: 'Up',
  [STATUS.PENDING]: 'Pending',
  [STATUS.MAINTENANCE]: 'Maintenance',
  [STATUS.DEGRADED]: 'Degraded',
};

export const MONITOR_TYPE_CATEGORIES = {
  General: ['http', 'port', 'ping', 'dns', 'push', 'group'],
  Network: ['smtp', 'tailscale-ping'],
  Messaging: ['mqtt'],
  Infrastructure: ['grpc-keyword'],
  Database: ['redis'],
} as const;

export const MONITOR_TYPES = Object.values(MONITOR_TYPE_CATEGORIES).flat();
