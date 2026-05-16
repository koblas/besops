import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type DockerHost = components['schemas']['DockerHost'];
export type DockerHostInput = components['schemas']['DockerHostInput'];

export function useDockerHosts() {
  return useQuery({
    queryKey: ['docker-hosts'],
    queryFn: async () => {
      const { data, error } = await api.GET('/docker-hosts');
      if (error) throw error;
      return data;
    },
  });
}

export function useCreateDockerHost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: DockerHostInput) => {
      const { data, error } = await api.POST('/docker-hosts', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docker-hosts'] });
    },
  });
}

export function useUpdateDockerHost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: DockerHostInput }) => {
      const { error } = await api.PUT('/docker-hosts/{dockerHostId}', {
        params: { path: { dockerHostId: id } },
        body: input,
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docker-hosts'] });
    },
  });
}

export function useDeleteDockerHost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/docker-hosts/{dockerHostId}', {
        params: { path: { dockerHostId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docker-hosts'] });
    },
  });
}

export function useTestDockerHost() {
  return useMutation({
    mutationFn: async (id: string) => {
      const { data, error } = await api.POST('/docker-hosts/{dockerHostId}/test', {
        params: { path: { dockerHostId: id } },
      });
      if (error) throw error;
      return data;
    },
  });
}
