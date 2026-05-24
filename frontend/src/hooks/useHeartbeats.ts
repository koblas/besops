import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Heartbeat = components['schemas']['Heartbeat'];
export type ChartPoint = components['schemas']['ChartPoint'];

export function useHeartbeats(monitorId: string | undefined, opts?: { hours?: number; count?: number }) {
  const query = opts ?? { count: 100 };
  return useQuery({
    queryKey: ['heartbeats', monitorId, query],
    queryFn: async () => {
      const { data, error } = await api.GET('/monitors/{monitorId}/heartbeats', {
        params: { path: { monitorId: monitorId! }, query },
      });
      if (error) throw error;
      // API returns newest-first; consumers expect chronological (oldest-first)
      return [...(data ?? [])].reverse();
    },
    enabled: !!monitorId,
  });
}

export function useChartData(monitorId: string | undefined, hours?: number) {
  return useQuery({
    queryKey: ['chart', monitorId, hours],
    queryFn: async () => {
      const { data, error } = await api.GET('/monitors/{monitorId}/heartbeats/chart', {
        params: { path: { monitorId: monitorId! }, query: { hours: hours! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!monitorId && !!hours,
    refetchInterval: 5 * 60 * 1000,
  });
}

export interface ImportantHeartbeatsResponse {
  data: Heartbeat[];
  total: number;
}

export function useImportantHeartbeats(monitorId: string | undefined, limit = 25, offset = 0) {
  return useQuery({
    queryKey: ['events', monitorId, offset, limit],
    queryFn: async () => {
      const { data, error } = await api.GET('/monitors/{monitorId}/events', {
        params: { path: { monitorId: monitorId! }, query: { limit, offset } },
      });
      if (error) throw error;
      return data as unknown as ImportantHeartbeatsResponse;
    },
    enabled: !!monitorId,
  });
}

export function useRecentEvents(limit = 25) {
  return useQuery({
    queryKey: ['events', 'all', limit],
    queryFn: async () => {
      const { data, error } = await api.GET('/events', {
        params: { query: { limit } },
      });
      if (error) throw error;
      return data as unknown as ImportantHeartbeatsResponse;
    },
  });
}
