import { useEffect, useState, useCallback } from 'react';
import { apiClient } from '@/lib/api/client';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];
type LogFilters = {
  session_id?: string;
  employee_id?: string;
  agent_id?: string;
  event_type?: string;
  event_category?: string;
  start_date?: string;
  end_date?: string;
  search?: string;
  limit?: number;
  offset?: number;
};

export interface UseActivityLogsReturn {
  logs: ActivityLog[] | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useActivityLogs(filters: LogFilters = {}): UseActivityLogsReturn {
  const [logs, setLogs] = useState<ActivityLog[] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      type EventType = 'input' | 'output' | 'error' | 'session_start' | 'session_end' | 'agent.installed' | 'mcp.configured' | 'config.synced';
      type EventCategory = 'io' | 'agent' | 'mcp' | 'auth' | 'admin';

      const { data, error: apiError } = await apiClient.GET('/logs', {
        params: {
          query: {
            session_id: filters.session_id,
            employee_id: filters.employee_id,
            agent_id: filters.agent_id,
            event_type: filters.event_type as EventType,
            event_category: filters.event_category as EventCategory,
            start_date: filters.start_date,
            end_date: filters.end_date,
            limit: filters.limit || 100,
            offset: filters.offset || 0,
          },
        },
      });

      if (apiError) {
        throw new Error(apiError.message || 'Failed to fetch logs');
      }

      setLogs(data?.logs || []);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));
    } finally {
      setIsLoading(false);
    }
  }, [
    filters.session_id,
    filters.employee_id,
    filters.agent_id,
    filters.event_type,
    filters.event_category,
    filters.start_date,
    filters.end_date,
    filters.limit,
    filters.offset,
  ]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  return {
    logs,
    isLoading,
    error,
    refetch: fetchLogs,
  };
}
