'use client';

import { useState, useCallback } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { LogList } from '@/components/logs/LogList';
import { LogFilters } from '@/components/logs/LogFilters';
import { ExportMenu } from '@/components/logs/ExportMenu';
import { useActivityLogs } from '@/lib/hooks/useActivityLogs';
import { useLogWebSocket } from '@/lib/hooks/useLogWebSocket';
import { X } from 'lucide-react';

export interface LogFiltersState {
  session_id?: string;
  employee_id?: string;
  agent_id?: string;
  event_type?: string;
  event_category?: string;
  start_date?: string;
  end_date?: string;
  search?: string;
}

export function LogsClient() {
  const [filters, setFilters] = useState<LogFiltersState>({});

  // Fetch logs with filters
  const { logs, isLoading, error, refetch } = useActivityLogs(filters);

  // WebSocket for real-time updates
  const { connected, newLogs } = useLogWebSocket();

  // Merge real-time logs with existing logs
  const allLogs = [...(logs || []), ...(newLogs || [])];

  const handleFilterChange = useCallback((newFilters: Partial<LogFiltersState>) => {
    setFilters((prev) => ({ ...prev, ...newFilters }));
  }, []);

  const handleClearFilters = useCallback(() => {
    setFilters({});
  }, []);

  if (error) {
    return (
      <Card className="p-6">
        <div className="text-center text-red-600">
          <p>Failed to load logs. Please try again.</p>
          <Button onClick={() => refetch()} variant="outline" className="mt-4">
            Retry
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4 responsive">
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

      {/* Log List */}
      {isLoading ? (
        <div role="status" className="flex justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white" />
          <span className="sr-only">Loading logs...</span>
        </div>
      ) : (
        <LogList logs={allLogs} />
      )}
    </div>
  );
}
