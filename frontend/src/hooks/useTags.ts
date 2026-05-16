import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Tag = components['schemas']['Tag'];
export type TagInput = components['schemas']['TagInput'];
export type MonitorTag = components['schemas']['MonitorTag'];

export function useTags() {
  return useQuery({
    queryKey: ['tags'],
    queryFn: async () => {
      const { data, error } = await api.GET('/tags');
      if (error) throw error;
      return data;
    },
  });
}

export function useCreateTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: TagInput) => {
      const { data, error } = await api.POST('/tags', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
    },
  });
}

export function useUpdateTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: TagInput }) => {
      const { data, error } = await api.PUT('/tags/{tagId}', {
        params: { path: { tagId: id } },
        body: input,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
    },
  });
}

export function useDeleteTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/tags/{tagId}', {
        params: { path: { tagId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
    },
  });
}

export function useAddMonitorTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ monitorId, tagId, value }: { monitorId: string; tagId: string; value?: string }) => {
      const { error } = await api.POST('/monitors/{monitorId}/tags', {
        params: { path: { monitorId } },
        body: { tagId, value },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function useRemoveMonitorTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ monitorId, tagId }: { monitorId: string; tagId: string }) => {
      const { error } = await api.DELETE('/monitors/{monitorId}/tags/{tagId}', {
        params: { path: { monitorId, tagId } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}
