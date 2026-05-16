import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Maintenance = components['schemas']['Maintenance'];
export type MaintenanceInput = components['schemas']['MaintenanceInput'];

export function useMaintenanceList() {
  return useQuery({
    queryKey: ['maintenance'],
    queryFn: async () => {
      const { data, error } = await api.GET('/maintenance');
      if (error) throw error;
      return data;
    },
  });
}

export function useMaintenance(id: string | undefined) {
  return useQuery({
    queryKey: ['maintenance', id],
    queryFn: async () => {
      const { data, error } = await api.GET('/maintenance/{maintenanceId}', {
        params: { path: { maintenanceId: id! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateMaintenance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: MaintenanceInput) => {
      const { data, error } = await api.POST('/maintenance', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useUpdateMaintenance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: MaintenanceInput }) => {
      const { data, error } = await api.PUT('/maintenance/{maintenanceId}', {
        params: { path: { maintenanceId: id } },
        body: input,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useDeleteMaintenance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/maintenance/{maintenanceId}', {
        params: { path: { maintenanceId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function usePauseMaintenance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.POST('/maintenance/{maintenanceId}/pause', {
        params: { path: { maintenanceId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useResumeMaintenance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.POST('/maintenance/{maintenanceId}/resume', {
        params: { path: { maintenanceId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance'] });
    },
  });
}

export function useMaintenanceMonitors(id: string | undefined) {
  return useQuery({
    queryKey: ['maintenance', id, 'monitors'],
    queryFn: async () => {
      const { data, error } = await api.GET('/maintenance/{maintenanceId}/monitors', {
        params: { path: { maintenanceId: id! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!id,
  });
}

export function useSetMaintenanceMonitors() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, monitorIds }: { id: string; monitorIds: string[] }) => {
      const { error } = await api.PUT('/maintenance/{maintenanceId}/monitors', {
        params: { path: { maintenanceId: id } },
        body: { monitorIds },
      });
      if (error) throw error;
    },
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['maintenance', id, 'monitors'] });
    },
  });
}

export function useMaintenanceStatusPages(id: string | undefined) {
  return useQuery({
    queryKey: ['maintenance', id, 'status-pages'],
    queryFn: async () => {
      const { data, error } = await api.GET('/maintenance/{maintenanceId}/status-pages', {
        params: { path: { maintenanceId: id! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!id,
  });
}

export function useSetMaintenanceStatusPages() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, statusPageIds }: { id: string; statusPageIds: string[] }) => {
      const { error } = await api.PUT('/maintenance/{maintenanceId}/status-pages', {
        params: { path: { maintenanceId: id } },
        body: { statusPageIds },
      });
      if (error) throw error;
    },
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['maintenance', id, 'status-pages'] });
    },
  });
}
