import { useEffect, useState, useRef, useCallback } from 'react';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];

export interface UseLogWebSocketReturn {
  connected: boolean;
  newLogs: ActivityLog[];
  error: Error | null;
  clearNewLogs: () => void;
}

const WS_URL =
  process.env.NEXT_PUBLIC_WS_URL ||
  (typeof window !== 'undefined' && window.location.origin.replace('http', 'ws')) ||
  'ws://localhost:3001';

export function useLogWebSocket(): UseLogWebSocketReturn {
  const [connected, setConnected] = useState(false);
  const [newLogs, setNewLogs] = useState<ActivityLog[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    try {
      // Close existing connection if any
      if (wsRef.current) {
        wsRef.current.close();
      }

      const ws = new WebSocket(`${WS_URL}/api/v1/logs/stream`);

      ws.onopen = () => {
        console.log('[WebSocket] Connected to log stream');
        setConnected(true);
        setError(null);
      };

      ws.onmessage = (event) => {
        try {
          const log: ActivityLog = JSON.parse(event.data);
          setNewLogs((prev) => [...prev, log]);
        } catch (err) {
          console.error('[WebSocket] Failed to parse log message:', err);
        }
      };

      ws.onerror = (event) => {
        console.error('[WebSocket] Error:', event);
        setError(new Error('WebSocket connection error'));
      };

      ws.onclose = () => {
        console.log('[WebSocket] Disconnected from log stream');
        setConnected(false);

        // Attempt to reconnect after 5 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log('[WebSocket] Attempting to reconnect...');
          connect();
        }, 5000);
      };

      wsRef.current = ws;
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to connect to WebSocket'));
    }
  }, []);

  const clearNewLogs = useCallback(() => {
    setNewLogs([]);
  }, []);

  useEffect(() => {
    connect();

    // Cleanup on unmount
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return {
    connected,
    newLogs,
    error,
    clearNewLogs,
  };
}
