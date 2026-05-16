import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useAuth } from './useAuth';
import { connectSSE } from '../lib/sse';
import type { Heartbeat } from './useHeartbeats';

export function useSSE() {
  const { token, isAuthenticated } = useAuth();
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!isAuthenticated || !token) return;

    const disconnect = connectSSE('/api/v1/ws/events', token, event => {
      try {
        const payload = JSON.parse(event.data);
        switch (event.event) {
          case 'heartbeat': {
            const hb = payload as Heartbeat;
            queryClient.setQueriesData<Heartbeat[]>(
              { queryKey: ['heartbeats', hb.monitorId] },
              old => old ? [...old.slice(-(500 - 1)), hb] : [hb],
            );
            queryClient.invalidateQueries({ queryKey: ['monitors'] });
            break;
          }
          case 'monitorList':
            queryClient.invalidateQueries({ queryKey: ['monitors'] });
            break;
          case 'notificationList':
            queryClient.invalidateQueries({ queryKey: ['notifications'] });
            break;
          case 'maintenanceList':
            queryClient.invalidateQueries({ queryKey: ['maintenance'] });
            break;
        }
      } catch {
        // ignore malformed events
      }
    });

    return disconnect;
  }, [isAuthenticated, token, queryClient]);
}
