'use client';

import { useState, useCallback, useMemo } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { LogList } from '@/components/logs/LogList';
import { LogFilters } from '@/components/logs/LogFilters';
import { ExportMenu } from '@/components/logs/ExportMenu';
import { useActivityLogs } from '@/lib/hooks/useActivityLogs';
import { useLogWebSocket } from '@/lib/hooks/useLogWebSocket';
import { X, AlertCircle } from 'lucide-react';

export interface LogFiltersState {
  employee_id?: string;
  event_type?: string;
  event_category?: string;
  start_date?: string;
  end_date?: string;
  search?: string;
}

const DEFAULT_PER_PAGE = 20;

export function LogsClient() {
  const [filters, setFilters] = useState<LogFiltersState>({});
  const [page, setPage] = useState(1);

  // Fetch logs with filters and pagination
  const { logs, pagination, isLoading, error, refetch } = useActivityLogs({
    ...filters,
    page,
    per_page: DEFAULT_PER_PAGE,
  });

  // WebSocket for real-time updates
  const { connected, newLogs, clearNewLogs } = useLogWebSocket();

  // Track new log IDs for highlighting
  const newLogIds = useMemo(() => new Set(newLogs.map((l) => l.id)), [newLogs]);

  // Merge real-time logs at the beginning of current page (only on page 1)
  const allLogs = useMemo(() => {
    if (page === 1 && newLogs.length > 0) {
      // Prepend new logs, avoiding duplicates
      const existingIds = new Set((logs || []).map((l) => l.id));
      const uniqueNewLogs = newLogs.filter((l) => !existingIds.has(l.id));
      return [...uniqueNewLogs, ...(logs || [])];
    }
    return logs || [];
  }, [logs, newLogs, page]);

  const handleFilterChange = useCallback(
    (newFilters: Partial<LogFiltersState>) => {
      setFilters((prev) => ({ ...prev, ...newFilters }));
      setPage(1); // Reset to first page on filter change
      clearNewLogs(); // Clear real-time logs on filter change
    },
    [clearNewLogs]
  );

  const handleClearFilters = useCallback(() => {
    setFilters({});
    setPage(1);
    clearNewLogs();
  }, [clearNewLogs]);

  const handlePageChange = useCallback(
    (newPage: number) => {
      setPage(newPage);
      // Clear new logs when navigating away from page 1
      if (newPage !== 1) {
        clearNewLogs();
      }
    },
    [clearNewLogs]
  );

  if (error) {
    return (
      <Card className="p-6">
        <div className="flex flex-col items-center gap-4 text-center">
          <AlertCircle className="h-8 w-8 text-red-600" />
          <p className="text-red-600">Failed to load logs. Please try again.</p>
          <Button onClick={() => refetch()} variant="outline">
            Retry
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header with Live indicator and Export */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {connected && (
            <Badge
              variant="outline"
              className="gap-1"
              data-testid="live-indicator"
            >
              <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
              Live
            </Badge>
          )}
          {newLogs.length > 0 && page === 1 && (
            <Badge variant="secondary" className="gap-1">
              {newLogs.length} new
            </Badge>
          )}
        </div>

        <ExportMenu filters={filters} />
      </div>

      {/* Filters */}
      <Card className="p-4">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Filters</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClearFilters}
              disabled={Object.keys(filters).length === 0}
            >
              <X className="h-4 w-4 mr-1" />
              Clear Filters
            </Button>
          </div>

          <LogFilters filters={filters} onChange={handleFilterChange} />
        </div>
      </Card>

      {/* Log List with Pagination */}
      {isLoading ? (
        <div role="status" className="flex justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white" />
          <span className="sr-only">Loading logs...</span>
        </div>
      ) : (
        <LogList
          logs={allLogs}
          pagination={pagination}
          onPageChange={handlePageChange}
          newLogIds={newLogIds}
        />
      )}
    </div>
  );
}
