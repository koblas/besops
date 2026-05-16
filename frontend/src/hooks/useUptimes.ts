import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export type UptimeMap = Record<string, number>;

export function useUptimes() {
  return useQuery({
    queryKey: ['uptimes'],
    queryFn: async (): Promise<UptimeMap> => {
      const { data } = await api.GET('/monitors/uptimes');
      return data ?? {};
    },
    refetchInterval: 60_000,
  });
}
