'use client';

import { useState, Fragment } from 'react';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Pagination } from '@/components/ui/pagination';
import { ChevronDown, ChevronRight, Sparkles } from 'lucide-react';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];
type PaginationMeta = components['schemas']['PaginationMeta'];

interface LogListProps {
  logs: ActivityLog[];
  pagination: PaginationMeta | null;
  onPageChange: (page: number) => void;
  newLogIds?: Set<string>;
}

const EVENT_TYPE_COLORS: Record<string, string> = {
  tool_call: 'bg-orange-500',
  api_request: 'bg-indigo-500',
  api_response: 'bg-indigo-400',
};

export function LogList({ logs, pagination, onPageChange, newLogIds }: LogListProps) {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  const toggleRow = (logId: string) => {
    setExpandedRows((prev) => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
      } else {
        next.add(logId);
      }
      return next;
    });
  };

  if (logs.length === 0) {
    return (
      <Card className="p-12 text-center border-dashed">
        <div className="flex flex-col items-center gap-2">
          <Sparkles className="h-8 w-8 text-muted-foreground" />
          <p className="text-muted-foreground">No logs found</p>
          <p className="text-sm text-muted-foreground">
            Try adjusting your filters or wait for new activity
          </p>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <Card className="overflow-hidden">
        <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10"></TableHead>
                <TableHead className="w-44">Time</TableHead>
                <TableHead className="w-28">Agent</TableHead>
                <TableHead className="w-32">Event Type</TableHead>
                <TableHead className="w-28">Category</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <LogRow
                  key={log.id}
                  log={log}
                  expanded={expandedRows.has(log.id)}
                  onToggle={() => toggleRow(log.id)}
                  isNew={newLogIds?.has(log.id)}
                />
              ))}
            </TableBody>
          </Table>
      </Card>

      {pagination && pagination.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(pagination.page - 1) * pagination.per_page + 1} -{' '}
            {Math.min(pagination.page * pagination.per_page, pagination.total)} of{' '}
            {pagination.total} logs
          </p>
          <Pagination
            currentPage={pagination.page}
            totalPages={pagination.total_pages}
            onPageChange={onPageChange}
          />
        </div>
      )}
    </div>
  );
}

interface LogRowProps {
  log: ActivityLog;
  expanded: boolean;
  onToggle: () => void;
  isNew?: boolean;
}

function LogRow({ log, expanded, onToggle, isNew }: LogRowProps) {
  const hasExpandableContent =
    (log.content && log.content.length > 50) ||
    (log.payload && Object.keys(log.payload).length > 0);

  const eventTypeColor = EVENT_TYPE_COLORS[log.event_type] || 'bg-gray-400';

  return (
    <Fragment>
      <TableRow
        className={`cursor-pointer ${isNew ? 'bg-green-50 dark:bg-green-950/20' : ''}`}
        onClick={hasExpandableContent ? onToggle : undefined}
      >
        <TableCell className="p-2">
          {hasExpandableContent && (
            <Button variant="ghost" size="icon" className="h-6 w-6">
              {expanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </Button>
          )}
        </TableCell>

        <TableCell className="font-mono text-sm">
          {formatDateTime(log.created_at)}
        </TableCell>

        <TableCell>
          {log.client_name ? (
            <span className="text-sm text-muted-foreground">{log.client_name}</span>
          ) : (
            <span className="text-muted-foreground">-</span>
          )}
        </TableCell>

        <TableCell>
          <div className="flex items-center gap-2">
            <div className={`h-2 w-2 rounded-full ${eventTypeColor}`} />
            <Badge variant="outline" className="text-xs">
              {log.event_type}
            </Badge>
          </div>
        </TableCell>

        <TableCell>
          {log.event_category && (
            <Badge variant="secondary" className="text-xs">
              {log.event_category}
            </Badge>
          )}
        </TableCell>
      </TableRow>

      {expanded && hasExpandableContent && (
        <TableRow className="bg-muted/30 hover:bg-muted/30">
          <TableCell colSpan={5} className="p-4">
            <ExpandedLogContent log={log} />
          </TableCell>
        </TableRow>
      )}
    </Fragment>
  );
}

function ExpandedLogContent({ log }: { log: ActivityLog }) {
  const [showFullContent, setShowFullContent] = useState(false);

  return (
    <div className="space-y-4">
      {log.content && (
        <div>
          <h4 className="text-sm font-medium mb-2">Content</h4>
          <div
            className={`bg-background rounded-md p-3 ${
              !showFullContent && log.content.length > 500
                ? 'max-h-40 overflow-hidden'
                : ''
            }`}
          >
            <pre className="text-sm whitespace-pre-wrap break-words font-mono">
              {log.content}
            </pre>
          </div>
          {log.content.length > 500 && (
            <Button
              variant="link"
              size="sm"
              className="mt-1 h-auto p-0"
              onClick={(e) => {
                e.stopPropagation();
                setShowFullContent(!showFullContent);
              }}
            >
              {showFullContent ? 'Show less' : 'Show more'}
            </Button>
          )}
        </div>
      )}

      {log.payload && Object.keys(log.payload).length > 0 && (
        <div>
          <h4 className="text-sm font-medium mb-2">Payload</h4>
          <pre className="text-xs bg-background p-3 rounded-md font-mono whitespace-pre-wrap break-all">
            {JSON.stringify(log.payload, null, 2)}
          </pre>
        </div>
      )}

      <div className="flex gap-4 text-xs text-muted-foreground">
        {log.employee_id && <span>Employee: {log.employee_id.slice(0, 8)}</span>}
      </div>
    </div>
  );
}

function formatDateTime(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}
