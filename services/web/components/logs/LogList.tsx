'use client';

import { useState, useMemo } from 'react';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ChevronDown, ChevronRight, Clock, User, Bot } from 'lucide-react';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];

interface LogListProps {
  logs: ActivityLog[];
}

interface SessionGroup {
  session_id: string;
  logs: ActivityLog[];
  start_time: string;
  end_time?: string;
  duration?: number;
  employee_name?: string;
  agent_name?: string;
}

export function LogList({ logs }: LogListProps) {
  const [expandedSessions, setExpandedSessions] = useState<Set<string>>(new Set());

  // Group logs by session
  const sessionGroups = useMemo(() => {
    const groups = new Map<string, SessionGroup>();

    logs.forEach((log) => {
      if (!log.session_id) return;

      if (!groups.has(log.session_id)) {
        groups.set(log.session_id, {
          session_id: log.session_id,
          logs: [],
          start_time: log.created_at,
          employee_name: log.employee_name,
          agent_name: log.agent_name,
        });
      }

      const group = groups.get(log.session_id)!;
      group.logs.push(log);

      // Update end time and duration
      if (log.event_type === 'session_end') {
        group.end_time = log.created_at;
        const start = new Date(group.start_time).getTime();
        const end = new Date(log.created_at).getTime();
        group.duration = Math.floor((end - start) / 1000); // seconds
      }
    });

    // Sort by most recent first
    return Array.from(groups.values()).sort(
      (a, b) => new Date(b.start_time).getTime() - new Date(a.start_time).getTime()
    );
  }, [logs]);

  const toggleSession = (sessionId: string) => {
    setExpandedSessions((prev) => {
      const next = new Set(prev);
      if (next.has(sessionId)) {
        next.delete(sessionId);
      } else {
        next.add(sessionId);
      }
      return next;
    });
  };

  if (logs.length === 0) {
    return (
      <Card className="p-12 text-center">
        <p className="text-muted-foreground">No logs found</p>
        <p className="text-sm text-muted-foreground mt-1">
          Try adjusting your filters or wait for new activity
        </p>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {sessionGroups.map((session) => (
        <SessionCard
          key={session.session_id}
          session={session}
          expanded={expandedSessions.has(session.session_id)}
          onToggle={() => toggleSession(session.session_id)}
        />
      ))}
    </div>
  );
}

interface SessionCardProps {
  session: SessionGroup;
  expanded: boolean;
  onToggle: () => void;
}

function SessionCard({ session, expanded, onToggle }: SessionCardProps) {
  return (
    <Card className="overflow-hidden">
      {/* Session Header */}
      <div
        className="p-4 bg-muted/50 flex items-center justify-between cursor-pointer hover:bg-muted"
        onClick={onToggle}
      >
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" className="h-6 w-6">
            {expanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </Button>

          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <span className="font-mono text-sm">
                Session: {session.session_id.slice(0, 8)}
              </span>
              {session.duration && (
                <Badge variant="outline">
                  <Clock className="h-3 w-3 mr-1" />
                  {formatDuration(session.duration)}
                </Badge>
              )}
            </div>

            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              {session.employee_name && (
                <span className="flex items-center gap-1">
                  <User className="h-3 w-3" />
                  {session.employee_name}
                </span>
              )}
              {session.agent_name && (
                <span className="flex items-center gap-1">
                  <Bot className="h-3 w-3" />
                  {session.agent_name}
                </span>
              )}
              <span>{new Date(session.start_time).toLocaleString()}</span>
            </div>
          </div>
        </div>

        <Badge>{session.logs.length} events</Badge>
      </div>

      {/* Session Logs */}
      {expanded && (
        <div className="divide-y">
          {session.logs.map((log) => (
            <LogEntry key={log.id} log={log} />
          ))}
        </div>
      )}
    </Card>
  );
}

interface LogEntryProps {
  log: ActivityLog;
}

function LogEntry({ log }: LogEntryProps) {
  const [expanded, setExpanded] = useState(false);

  const eventTypeColor = {
    input: 'bg-blue-500',
    output: 'bg-green-500',
    error: 'bg-red-500',
    session_start: 'bg-purple-500',
    session_end: 'bg-gray-500',
  }[log.event_type as string] || 'bg-gray-400';

  return (
    <div className="p-4 hover:bg-muted/50 transition-colors">
      <div className="flex items-start gap-3">
        <div className={`mt-1 h-2 w-2 rounded-full ${eventTypeColor}`} />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-medium">{formatTime(log.created_at)}</span>
            <Badge variant="outline" className="text-xs">
              {log.event_type}
            </Badge>
            {log.event_category && (
              <Badge variant="secondary" className="text-xs">
                {log.event_category}
              </Badge>
            )}
          </div>

          {log.content && (
            <>
              <div
                className={`text-sm ${
                  expanded ? '' : 'line-clamp-2'
                } cursor-pointer`}
                onClick={() => setExpanded(!expanded)}
              >
                <pre className="whitespace-pre-wrap break-words font-mono text-xs">
                  {log.content}
                </pre>
              </div>
              {log.content.length > 100 && (
                <Button
                  variant="link"
                  size="sm"
                  className="h-auto p-0 text-xs"
                  onClick={() => setExpanded(!expanded)}
                >
                  {expanded ? 'Show less' : 'Show more'}
                </Button>
              )}
            </>
          )}

          {log.metadata && (
            <details className="mt-2">
              <summary className="text-xs text-muted-foreground cursor-pointer">
                Metadata
              </summary>
              <pre className="mt-1 text-xs bg-muted p-2 rounded overflow-x-auto">
                {JSON.stringify(log.metadata, null, 2)}
              </pre>
            </details>
          )}
        </div>
      </div>
    </div>
  );
}

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}m ${secs}s`;
}

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString();
}
