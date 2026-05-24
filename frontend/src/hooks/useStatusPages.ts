import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type StatusPage = components['schemas']['StatusPage'];
export type StatusPageInput = components['schemas']['StatusPageInput'];
export type StatusPageGroup = components['schemas']['StatusPageGroup'];

export function useStatusPages() {
  return useQuery({
    queryKey: ['status-pages'],
    queryFn: async () => {
      const { data, error } = await api.GET('/status-pages');
      if (error) throw error;
      return data;
    },
  });
}

export function useStatusPage(slug: string | undefined) {
  return useQuery({
    queryKey: ['status-pages', slug],
    queryFn: async () => {
      const { data, error } = await api.GET('/status-pages/{slug}', {
        params: { path: { slug: slug! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!slug,
  });
}

export function useCreateStatusPage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: StatusPageInput) => {
      const { data, error } = await api.POST('/status-pages', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status-pages'] });
    },
  });
}

export function useUpdateStatusPage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, input }: { slug: string; input: StatusPageInput }) => {
      const { data, error } = await api.PUT('/status-pages/{slug}', {
        params: { path: { slug } },
        body: input,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, { slug }) => {
      queryClient.invalidateQueries({ queryKey: ['status-pages'] });
      queryClient.invalidateQueries({ queryKey: ['status-pages', slug] });
    },
  });
}

export function useDeleteStatusPage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (slug: string) => {
      const { error } = await api.DELETE('/status-pages/{slug}', {
        params: { path: { slug } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status-pages'] });
    },
  });
}

type Heartbeat = components['schemas']['Heartbeat'];

interface HeartbeatData {
  heartbeatList: Record<string, Heartbeat[]>;
  uptimeList: Record<string, number>;
  monitorNames: Record<string, string>;
}

export function useStatusPageHeartbeats(slug: string | undefined) {
  return useQuery({
    queryKey: ['status-pages', slug, 'heartbeats'],
    queryFn: async (): Promise<HeartbeatData> => {
      const { data, error } = await api.GET('/status-pages/{slug}/heartbeats', {
        params: { path: { slug: slug! } },
      });
      if (error) throw error;
      const heartbeatList: Record<string, Heartbeat[]> = {};
      for (const item of data?.heartbeatList ?? []) {
        heartbeatList[item.monitorId] = item.heartbeats;
      }
      const uptimeList: Record<string, number> = {};
      for (const item of data?.uptimeList ?? []) {
        uptimeList[`${item.monitorId}_24`] = item.uptime;
      }
      const monitorNames: Record<string, string> = {};
      for (const item of data?.monitorNames ?? []) {
        monitorNames[item.monitorId] = item.name;
      }
      return { heartbeatList, uptimeList, monitorNames };
    },
    enabled: !!slug,
  });
}
