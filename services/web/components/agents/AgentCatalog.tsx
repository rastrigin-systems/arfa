'use client';

import { AgentCard } from './AgentCard';
import { Button } from '@/components/ui/button';
import { AlertCircle } from 'lucide-react';
import type { Agent } from '@/lib/types';

type AgentCatalogProps = {
  agents: Agent[];
  enabledAgentIds: Set<string>;
  isLoading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
  onEnable?: (agentId: string) => void;
  onConfigure?: (agentId: string) => void;
};

function LoadingSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {[...Array(6)].map((_, i) => (
        <div
          key={i}
          className="h-64 animate-pulse rounded-lg border bg-muted"
          role="status"
          aria-label="Loading agent card"
        />
      ))}
    </div>
  );
}

function ErrorState({ error, onRetry }: { error: Error; onRetry?: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-destructive/50 bg-destructive/10 p-12 text-center">
      <AlertCircle className="h-12 w-12 text-destructive" />
      <div>
        <h3 className="text-lg font-semibold">Failed to load agents</h3>
        <p className="text-sm text-muted-foreground">{error.message}</p>
      </div>
      {onRetry && (
        <Button onClick={onRetry} variant="outline">
          Retry
        </Button>
      )}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed p-12 text-center">
      <div>
        <h3 className="text-lg font-semibold">No agents available</h3>
        <p className="text-sm text-muted-foreground">There are no AI agents available at this time.</p>
      </div>
    </div>
  );
}

export function AgentCatalog({ agents, enabledAgentIds, isLoading, error, onRetry, onEnable, onConfigure }: AgentCatalogProps) {
  // Loading state
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  // Error state
  if (error) {
    return <ErrorState error={error} onRetry={onRetry} />;
  }

  // Empty state
  if (agents.length === 0) {
    return <EmptyState />;
  }

  // Normal state - render grid
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {agents.map((agent) => (
        <AgentCard
          key={agent.id}
          agent={agent}
          isEnabled={enabledAgentIds.has(agent.id)}
          onEnable={onEnable}
          onConfigure={onConfigure}
        />
      ))}
    </div>
  );
}
