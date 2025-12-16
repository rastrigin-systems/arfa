'use client';

import { useState, useCallback, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { RefreshCw, Copy, Check, ChevronDown, ChevronRight } from 'lucide-react';

interface RawActivityLog {
  id: string;
  org_id: string;
  employee_id?: string | null;
  session_id?: string | null;
  agent_id?: string | null;
  event_type: string;
  event_category: string;
  content?: string | null;
  payload: Record<string, unknown>;
  created_at: string;
}

export function DebugLogsClient() {
  const [logs, setLogs] = useState<RawActivityLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [limit, setLimit] = useState(50);
  const [expandedLogs, setExpandedLogs] = useState<Set<string>>(new Set());
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch(`/api/logs?limit=${limit}&offset=0`);

      if (!response.ok) {
        throw new Error(`Failed to fetch logs: ${response.statusText}`);
      }

      const data = await response.json();
      setLogs(data?.logs || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setIsLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const toggleExpanded = (id: string) => {
    setExpandedLogs((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const expandAll = () => {
    setExpandedLogs(new Set(logs.map((log) => log.id)));
  };

  const collapseAll = () => {
    setExpandedLogs(new Set());
  };

  const copyToClipboard = async (log: RawActivityLog) => {
    try {
      await navigator.clipboard.writeText(JSON.stringify(log, null, 2));
      setCopiedId(log.id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch {
      // Clipboard copy failed silently
    }
  };

  const copyAllLogs = async () => {
    try {
      await navigator.clipboard.writeText(JSON.stringify(logs, null, 2));
      setCopiedId('all');
      setTimeout(() => setCopiedId(null), 2000);
    } catch {
      // Clipboard copy failed silently
    }
  };

  if (error) {
    return (
      <Card className="p-6">
        <div className="text-center text-red-600">
          <p>Failed to load logs: {error}</p>
          <Button onClick={fetchLogs} variant="outline" className="mt-4">
            Retry
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Controls */}
      <Card className="p-4">
        <div className="flex flex-wrap items-center gap-4">
          <div className="flex items-center gap-2">
            <label htmlFor="limit" className="text-sm font-medium">
              Limit:
            </label>
            <Input
              id="limit"
              type="number"
              value={limit}
              onChange={(e) => setLimit(parseInt(e.target.value, 10) || 50)}
              className="w-24"
              min={1}
              max={1000}
            />
          </div>

          <Button onClick={fetchLogs} variant="outline" size="sm" disabled={isLoading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>

          <Button onClick={expandAll} variant="outline" size="sm">
            Expand All
          </Button>

          <Button onClick={collapseAll} variant="outline" size="sm">
            Collapse All
          </Button>

          <Button onClick={copyAllLogs} variant="outline" size="sm">
            {copiedId === 'all' ? (
              <>
                <Check className="h-4 w-4 mr-2" />
                Copied!
              </>
            ) : (
              <>
                <Copy className="h-4 w-4 mr-2" />
                Copy All JSON
              </>
            )}
          </Button>

          <Badge variant="secondary">{logs.length} logs</Badge>
        </div>
      </Card>

      {/* Loading state */}
      {isLoading ? (
        <div role="status" className="flex justify-center p-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white" />
          <span className="sr-only">Loading logs...</span>
        </div>
      ) : (
        /* Log entries */
        <div className="space-y-2">
          {logs.length === 0 ? (
            <Card className="p-8 text-center text-muted-foreground">
              No logs found. Logs will appear here as activity is recorded.
            </Card>
          ) : (
            logs.map((log) => (
              <Card key={log.id} className="overflow-hidden">
                {/* Log header - clickable */}
                <button
                  onClick={() => toggleExpanded(log.id)}
                  className="w-full flex items-center gap-3 p-4 text-left hover:bg-muted/50 transition-colors"
                >
                  {expandedLogs.has(log.id) ? (
                    <ChevronDown className="h-4 w-4 flex-shrink-0" />
                  ) : (
                    <ChevronRight className="h-4 w-4 flex-shrink-0" />
                  )}

                  <div className="flex-1 flex flex-wrap items-center gap-2">
                    <Badge variant="outline">{log.event_category}</Badge>
                    <Badge
                      variant={
                        log.event_type.includes('error')
                          ? 'destructive'
                          : log.event_type.includes('start')
                            ? 'default'
                            : 'secondary'
                      }
                    >
                      {log.event_type}
                    </Badge>
                    <span className="text-xs text-muted-foreground font-mono">
                      {log.id.substring(0, 8)}...
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {new Date(log.created_at).toLocaleString()}
                    </span>
                  </div>

                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      copyToClipboard(log);
                    }}
                  >
                    {copiedId === log.id ? (
                      <Check className="h-4 w-4" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </button>

                {/* Expanded JSON view */}
                {expandedLogs.has(log.id) && (
                  <div className="border-t bg-muted/30 p-4">
                    <pre className="text-xs font-mono overflow-x-auto whitespace-pre-wrap break-words">
                      {JSON.stringify(log, null, 2)}
                    </pre>
                  </div>
                )}
              </Card>
            ))
          )}
        </div>
      )}
    </div>
  );
}
