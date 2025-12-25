import { Suspense } from 'react';
import { DebugLogsClient } from './DebugLogsClient';

export const metadata = {
  title: 'Debug Logs | Arfa Enterprise',
  description: 'Raw JSON view of activity logs for debugging',
};

export default function DebugPage() {
  return (
    <div className="space-y-6" data-testid="debug-container">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Debug Logs</h1>
          <p className="text-muted-foreground mt-1">
            Raw JSON view of activity logs from the database
          </p>
        </div>
      </div>

      <Suspense fallback={<DebugLoadingSkeleton />}>
        <DebugLogsClient />
      </Suspense>
    </div>
  );
}

function DebugLoadingSkeleton() {
  return (
    <div role="status" className="animate-pulse space-y-4">
      <div className="h-12 bg-gray-200 rounded dark:bg-gray-700" />
      <div className="h-96 bg-gray-200 rounded dark:bg-gray-700" />
      <span className="sr-only">Loading debug logs...</span>
    </div>
  );
}
