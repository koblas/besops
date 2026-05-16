export const STATUS = {
  DOWN: 0,
  UP: 1,
  PENDING: 2,
  MAINTENANCE: 3,
} as const;

export type StatusValue = (typeof STATUS)[keyof typeof STATUS];

export const STATUS_COLORS: Record<StatusValue, string> = {
  [STATUS.DOWN]: '#f87171',
  [STATUS.UP]: '#34d399',
  [STATUS.PENDING]: '#fbbf24',
  [STATUS.MAINTENANCE]: '#60a5fa',
};

export const STATUS_LABELS: Record<StatusValue, string> = {
  [STATUS.DOWN]: 'Down',
  [STATUS.UP]: 'Up',
  [STATUS.PENDING]: 'Pending',
  [STATUS.MAINTENANCE]: 'Maintenance',
};

export const MONITOR_TYPE_CATEGORIES = {
  General: ['http', 'keyword', 'json-query', 'port', 'ping', 'dns', 'push', 'group'],
  Specific: ['steam', 'gamedig', 'mqtt', 'snmp', 'tailscale-ping', 'manual'],
  Database: ['postgres', 'mysql', 'mongodb', 'sqlserver', 'redis', 'radius', 'rabbitmq'],
  Docker: ['docker'],
  Browser: ['real-browser'],
} as const;

export const MONITOR_TYPES = Object.values(MONITOR_TYPE_CATEGORIES).flat();
