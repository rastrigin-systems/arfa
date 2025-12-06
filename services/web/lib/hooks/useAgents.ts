import { useQuery } from '@tanstack/react-query';

// Minimal types for Agent
export interface Agent {
  id: string;
  name: string;
  description?: string;
  version?: string;
}

export interface AgentsParams {
  page: number;
  limit: number;
  search?: string;
}

export interface AgentsResponse {
  agents: Agent[];
  total: number;
}

export function useAgents(params: AgentsParams) {
  return useQuery<AgentsResponse>({
    queryKey: ['agents', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      searchParams.set('page', String(params.page));
      searchParams.set('limit', String(params.limit));
      if (params.search) searchParams.set('search', params.search);

      const res = await fetch(`/api/agents?${searchParams.toString()}`);
      if (!res.ok) throw new Error('Failed to fetch agents');
      return res.json();
    },
  });
}
