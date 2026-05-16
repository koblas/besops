import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Proxy = components['schemas']['Proxy'];
export type ProxyInput = components['schemas']['ProxyInput'];

export function useProxies() {
  return useQuery({
    queryKey: ['proxies'],
    queryFn: async () => {
      const { data, error } = await api.GET('/proxies');
      if (error) throw error;
      return data;
    },
  });
}

export function useCreateProxy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: ProxyInput) => {
      const { data, error } = await api.POST('/proxies', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxies'] });
    },
  });
}

export function useUpdateProxy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: ProxyInput }) => {
      const { error } = await api.PUT('/proxies/{proxyId}', {
        params: { path: { proxyId: id } },
        body: input,
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxies'] });
    },
  });
}

export function useDeleteProxy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/proxies/{proxyId}', {
        params: { path: { proxyId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxies'] });
    },
  });
}
