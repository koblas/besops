import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import type { components } from '../api/generated/v1';

export type Notification = components['schemas']['Notification'];
export type NotificationInput = components['schemas']['NotificationInput'];

export function useNotifications() {
  return useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const { data, error } = await api.GET('/notifications');
      if (error) throw error;
      return data;
    },
  });
}

export function useCreateNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: NotificationInput) => {
      const { data, error } = await api.POST('/notifications', { body: input });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}

export function useUpdateNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: NotificationInput }) => {
      const { error } = await api.PUT('/notifications/{notificationId}', {
        params: { path: { notificationId: id } },
        body: input,
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}

export function useDeleteNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE('/notifications/{notificationId}', {
        params: { path: { notificationId: id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}

export function useTestNotification() {
  return useMutation({
    mutationFn: async (id: string) => {
      const { data, error } = await api.POST('/notifications/{notificationId}/test', {
        params: { path: { notificationId: id } },
      });
      if (error) throw error;
      return data;
    },
  });
}
