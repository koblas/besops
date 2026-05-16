import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

export function formatDateTime(date: string | Date): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss');
}

export function formatRelative(date: string | Date): string {
  return dayjs(date).fromNow();
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

export function formatPing(ping: number | null | undefined): string {
  if (ping == null) return '-';
  return `${Math.round(ping)} ms`;
}

export function formatUptime(percentage: number): string {
  return `${(percentage * 100).toFixed(2)}%`;
}
