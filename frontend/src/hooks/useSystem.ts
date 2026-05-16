import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '../api/client';

export function useServerInfo() {
  return useQuery({
    queryKey: ['info'],
    queryFn: async () => {
      const { data, error } = await api.GET('/info');
      if (error) throw error;
      return data;
    },
  });
}

export function useDatabaseSize() {
  return useQuery({
    queryKey: ['database', 'size'],
    queryFn: async () => {
      const { data, error } = await api.GET('/database/size');
      if (error) throw error;
      return data;
    },
  });
}

export function useShrinkDatabase() {
  return useMutation({
    mutationFn: async () => {
      const { data, error } = await api.POST('/database/shrink');
      if (error) throw error;
      return data;
    },
  });
}

export function useClearStatistics() {
  return useMutation({
    mutationFn: async () => {
      const { error } = await api.DELETE('/statistics');
      if (error) throw error;
    },
  });
}
