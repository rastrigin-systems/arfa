'use client';

import { useState } from 'react';
import { AgentCatalog } from '@/components/agents/AgentCatalog';
import { useRouter } from 'next/navigation';

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

type AgentCatalogClientProps = {
  initialAgents: Agent[];
  initialEnabledAgentIds: Set<string>;
};

export function AgentCatalogClient({ initialAgents, initialEnabledAgentIds }: AgentCatalogClientProps) {
  const [agents] = useState(initialAgents);
  const [enabledAgentIds] = useState(initialEnabledAgentIds);
  const router = useRouter();

  const handleEnable = async (agentId: string) => {
    // TODO: Implement enable agent API call
    console.log('Enable agent:', agentId);
    // This will be implemented when we have the API endpoint
    // For now, we'll just navigate to a placeholder
    alert(`Enable agent ${agentId} - API endpoint coming soon`);
  };

  const handleConfigure = (agentId: string) => {
    // Navigate to configuration page
    router.push(`/agents/${agentId}/configure`);
  };

  return (
    <AgentCatalog agents={agents} enabledAgentIds={enabledAgentIds} onEnable={handleEnable} onConfigure={handleConfigure} />
  );
}
