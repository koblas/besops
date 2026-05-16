import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Monitor = components['schemas']['Monitor'];
export type MonitorInput = components['schemas']['MonitorInput'];

export function useMonitors() {
  return useQuery({
    queryKey: ['monitors'],
    queryFn: async () => {
      const { data, error } = await api.GET('/monitors');
      if (error) throw error;
      return data;
    },
  });
}

export function useMonitor(id: string | undefined) {
  return useQuery({
    queryKey: ['monitors', id],
    queryFn: async () => {
      const { data, error } = await api.GET('/monitors/{monitorId}', {
        params: { path: { monitorId: id! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!id,
  });
}

export function usePauseMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.POST('/monitors/{monitorId}/pause', {
        params: { path: { monitorId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function useResumeMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.POST('/monitors/{monitorId}/resume', {
        params: { path: { monitorId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function useCreateMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: MonitorInput) => {
      const { data, error } = await api.POST('/monitors', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function useUpdateMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: MonitorInput }) => {
      const { data, error } = await api.PUT('/monitors/{monitorId}', {
        params: { path: { monitorId: id } },
        body: input,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
      queryClient.invalidateQueries({ queryKey: ['monitors', id] });
    },
  });
}

export function useDeleteMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/monitors/{monitorId}', {
        params: { path: { monitorId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}
