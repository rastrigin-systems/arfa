import { useEffect, useState, useCallback } from 'react';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];
type PaginationMeta = components['schemas']['PaginationMeta'];

type LogFilters = {
  session_id?: string;
  employee_id?: string;
  agent_id?: string;
  event_type?: string;
  event_category?: string;
  start_date?: string;
  end_date?: string;
  search?: string;
  page?: number;
  per_page?: number;
};

export interface UseActivityLogsReturn {
  logs: ActivityLog[] | null;
  pagination: PaginationMeta | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useActivityLogs(filters: LogFilters = {}): UseActivityLogsReturn {
  const [logs, setLogs] = useState<ActivityLog[] | null>(null);
  const [pagination, setPagination] = useState<PaginationMeta | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      // Build query parameters
      const params = new URLSearchParams();

      if (filters.session_id) params.append('session_id', filters.session_id);
      if (filters.employee_id) params.append('employee_id', filters.employee_id);
      if (filters.agent_id) params.append('agent_id', filters.agent_id);
      if (filters.event_type) params.append('event_type', filters.event_type);
      if (filters.event_category) params.append('event_category', filters.event_category);
      if (filters.start_date) params.append('start_date', filters.start_date);
      if (filters.end_date) params.append('end_date', filters.end_date);
      params.append('page', String(filters.page || 1));
      params.append('per_page', String(filters.per_page || 20));

      // Call Next.js API route instead of backend directly
      const response = await fetch(`/api/logs?${params.toString()}`);

      if (!response.ok) {
        throw new Error(`Failed to fetch logs: ${response.statusText}`);
      }

      const data = await response.json();
      setLogs(data?.logs || []);
      setPagination(data?.pagination || null);
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
    filters.page,
    filters.per_page,
  ]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  return {
    logs,
    pagination,
    isLoading,
    error,
    refetch: fetchLogs,
  };
}
