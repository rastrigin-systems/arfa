'use client';

import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';

type Agent = {
  id: string;
  name: string;
  type: string;
  description: string;
  provider: string;
  llm_provider: string;
  llm_model: string;
  is_public: boolean;
  capabilities?: string[];
  default_config?: Record<string, unknown>;
};

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
      className="flex flex-col bg-white border border-gray-200 rounded-lg p-6 shadow-md hover:shadow-lg transition-shadow"
      style={{ width: '280px', height: '220px' }}
      aria-label={`${agent.name} agent card`}
      data-testid="compact-agent-card"
    >
      {/* Title */}
      <h3 className="text-xl font-semibold text-gray-900 mb-2" data-testid="agent-name">
        {agent.name}
      </h3>

      {/* Separator */}
      <div className="h-px bg-gray-200 mb-3"></div>

      {/* Type Badge */}
      <Badge variant="secondary" className="w-fit text-xs font-medium uppercase mb-2">
        {agent.type}
      </Badge>

      {/* Description (2 lines max) */}
      <p className="text-sm text-gray-600 line-clamp-2 mb-4 flex-1" data-testid="agent-description">
        {agent.description}
      </p>

      {/* Toggle */}
      <div className="flex items-center mb-2">
        <Switch
          checked={isEnabled}
          onCheckedChange={handleToggle}
          disabled={isLoading}
          className={isEnabled ? 'data-[state=checked]:bg-green-600' : ''}
          aria-label={`Toggle ${agent.name} (currently ${isEnabled ? 'enabled' : 'disabled'})`}
        />
        <span className="ml-2 text-sm font-medium text-gray-900">
          {isEnabled ? 'Enabled' : 'Disabled'}
        </span>
      </div>

      {/* Action Button */}
      <Button
        onClick={handleConfigure}
        variant={isEnabled ? 'outline' : 'default'}
        className="w-full"
        disabled={isLoading}
        aria-label={isEnabled ? `Configure ${agent.name} settings` : `Enable ${agent.name} for organization`}
      >
        {isEnabled ? 'Configure →' : 'Enable'}
      </Button>
    </article>
  );
}
