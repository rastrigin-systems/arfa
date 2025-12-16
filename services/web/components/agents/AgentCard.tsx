'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { Agent } from '@/lib/types';

type AgentCardProps = {
  agent: Agent;
  isEnabled: boolean;
  onEnable?: (agentId: string) => void;
  onConfigure?: (agentId: string) => void;
};

export function AgentCard({ agent, isEnabled, onEnable, onConfigure }: AgentCardProps) {
  const handleAction = () => {
    if (isEnabled) {
      onConfigure?.(agent.id);
    } else {
      onEnable?.(agent.id);
    }
  };

  return (
    <article
      className="flex flex-col gap-4 rounded-lg border bg-card p-6 text-card-foreground shadow-sm transition-shadow hover:shadow-md"
      aria-label={`${agent.name} agent card`}
      data-testid="agent-card"
    >
      {/* Header with name and provider */}
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <h3 className="text-xl font-semibold" data-testid="agent-name">
            {agent.name}
          </h3>
          <p className="text-sm text-muted-foreground" data-testid="agent-provider">
            {agent.provider}
          </p>
        </div>
      </div>

      {/* Description */}
      <p className="text-sm text-muted-foreground" data-testid="agent-description">
        {agent.description}
      </p>

      {/* Capabilities */}
      {agent.capabilities && agent.capabilities.length > 0 && (
        <div className="flex flex-wrap gap-2" aria-label="Agent capabilities">
          {agent.capabilities.map((capability) => (
            <Badge key={capability} variant="secondary" className="text-xs">
              {capability}
            </Badge>
          ))}
        </div>
      )}

      {/* Action button */}
      <div className="mt-auto flex justify-end">
        <Button
          onClick={handleAction}
          variant={isEnabled ? 'outline' : 'default'}
          aria-label={isEnabled ? `Configure ${agent.name}` : `Enable ${agent.name} for organization`}
        >
          {isEnabled ? 'Configure' : 'Enable for Org'}
        </Button>
      </div>
    </article>
  );
}
