'use client';

import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import type { Agent } from '@/lib/types';

type CompactAgentCardProps = {
  agent: Agent;
  isEnabled: boolean;
  onToggle: (agentId: string, enabled: boolean) => void;
  onConfigure: (agentId: string) => void;
  isLoading?: boolean;
};

export function CompactAgentCard({ agent, isEnabled, onToggle, onConfigure, isLoading }: CompactAgentCardProps) {
  const handleToggle = (checked: boolean) => {
    onToggle(agent.id, checked);
  };

  const handleConfigure = () => {
    onConfigure(agent.id);
  };

  return (
    <article
      className="flex flex-col bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 shadow-sm hover:shadow-md transition-all duration-200 h-full min-h-[200px] max-h-[220px]"
      aria-label={`${agent.name} agent card`}
      data-testid="compact-agent-card"
    >
      {/* Title */}
      <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1.5 truncate" data-testid="agent-name">
        {agent.name}
      </h3>

      {/* Separator */}
      <div className="h-px bg-gray-200 dark:bg-gray-700 mb-2"></div>

      {/* Type Badge */}
      <Badge variant="secondary" className="w-fit text-[10px] font-medium uppercase mb-2 px-2 py-0.5">
        {agent.type.replace(/_/g, ' ')}
      </Badge>

      {/* Description (2 lines max) */}
      <p className="text-xs text-gray-600 dark:text-gray-400 line-clamp-2 mb-3 flex-1 leading-relaxed" data-testid="agent-description">
        {agent.description}
      </p>

      {/* Toggle */}
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-gray-700 dark:text-gray-300">
          {isEnabled ? 'Enabled' : 'Disabled'}
        </span>
        <Switch
          checked={isEnabled}
          onCheckedChange={handleToggle}
          disabled={isLoading}
          className={isEnabled ? 'data-[state=checked]:bg-green-600' : 'dark:bg-gray-600'}
          aria-label={`Toggle ${agent.name} (currently ${isEnabled ? 'enabled' : 'disabled'})`}
        />
      </div>

      {/* Configure Button - Available for all agents */}
      <Button
        onClick={handleConfigure}
        variant="outline"
        size="sm"
        className="w-full text-xs"
        disabled={isLoading}
        aria-label={`Configure ${agent.name} settings`}
      >
        Configure →
      </Button>
    </article>
  );
}
