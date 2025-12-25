import { Suspense } from 'react';
import { LogsClient } from './LogsClient';

export const metadata = {
  title: 'Activity Logs | Arfa Enterprise',
  description: 'View and search agent activity logs in real-time',
};

export default function LogsPage() {
  return (
    <div className="space-y-6" data-testid="logs-container">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Activity Logs</h1>
      </div>

      <Suspense fallback={<LogsLoadingSkeleton />}>
        <LogsClient />
      </Suspense>
    </div>
  );
}

function LogsLoadingSkeleton() {
  return (
    <div role="status" className="animate-pulse space-y-4">
      <div className="h-12 bg-gray-200 rounded dark:bg-gray-700" />
      <div className="h-96 bg-gray-200 rounded dark:bg-gray-700" />
      <span className="sr-only">Loading logs...</span>
    </div>
  );
}
