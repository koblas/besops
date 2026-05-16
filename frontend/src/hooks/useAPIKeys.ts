import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type APIKey = components['schemas']['APIKey'];
export type APIKeyInput = components['schemas']['APIKeyInput'];

export function useAPIKeys() {
  return useQuery({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const { data, error } = await api.GET('/api-keys');
      if (error) throw error;
      return data;
    },
  });
}

export function useCreateAPIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: APIKeyInput) => {
      const { data, error } = await api.POST('/api-keys', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}

export function useDeleteAPIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/api-keys/{keyId}', {
        params: { path: { keyId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}

export function useEnableAPIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.POST('/api-keys/{keyId}/enable', {
        params: { path: { keyId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}

export function useDisableAPIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.POST('/api-keys/{keyId}/disable', {
        params: { path: { keyId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}
