import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Incident = components['schemas']['Incident'];
export type IncidentInput = components['schemas']['IncidentInput'];

export function useIncidents(slug: string | undefined) {
  return useQuery({
    queryKey: ['status-pages', slug, 'incidents'],
    queryFn: async () => {
      const { data, error } = await api.GET('/status-pages/{slug}/incidents', {
        params: { path: { slug: slug! } },
      });
      if (error) throw error;
      return data;
    },
    enabled: !!slug,
  });
}

export function useCreateIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, input }: { slug: string; input: IncidentInput }) => {
      const { data, error } = await api.POST('/status-pages/{slug}/incidents', {
        params: { path: { slug } },
        body: input,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, { slug }) => {
      queryClient.invalidateQueries({ queryKey: ['status-pages', slug, 'incidents'] });
    },
  });
}

export function useUpdateIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, id, input }: { slug: string; id: string; input: IncidentInput }) => {
      const { data, error } = await api.PUT('/status-pages/{slug}/incidents/{incidentId}', {
        params: { path: { slug, incidentId: id } },
        body: input,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, { slug }) => {
      queryClient.invalidateQueries({ queryKey: ['status-pages', slug, 'incidents'] });
    },
  });
}

export function useDeleteIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, id }: { slug: string; id: string }) => {
      const { error } = await api.DELETE('/status-pages/{slug}/incidents/{incidentId}', {
        params: { path: { slug, incidentId: id } },
      });
      if (error) throw error;
    },
    onSuccess: (_data, { slug }) => {
      queryClient.invalidateQueries({ queryKey: ['status-pages', slug, 'incidents'] });
    },
  });
}

export function useResolveIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, id }: { slug: string; id: string }) => {
      const { data, error } = await api.POST('/status-pages/{slug}/incidents/{incidentId}/resolve', {
        params: { path: { slug, incidentId: id } },
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, { slug }) => {
      queryClient.invalidateQueries({ queryKey: ['status-pages', slug, 'incidents'] });
    },
  });
}

export function useUnpinIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ slug, id }: { slug: string; id: string }) => {
      const { error } = await api.POST('/status-pages/{slug}/incidents/{incidentId}/unpin', {
        params: { path: { slug, incidentId: id } },
      });
      if (error) throw error;
    },
    onSuccess: (_data, { slug }) => {
      queryClient.invalidateQueries({ queryKey: ['status-pages', slug, 'incidents'] });
    },
  });
}
