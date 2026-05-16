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

export function useStatusPageHeartbeats(slug: string | undefined) {
  return useQuery({
    queryKey: ['status-pages', slug, 'heartbeats'],
    queryFn: async () => {
      const { data, error } = await api.GET('/status-pages/{slug}/heartbeats', {
        params: { path: { slug: slug! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!slug,
  });
}
