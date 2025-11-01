/**
 * Agent Test Fixtures
 *
 * Mock data for agent-related E2E tests
 */

export const mockAgent = {
  id: '660e8400-e29b-41d4-a716-446655440001',
  name: 'Claude Code',
  slug: 'claude-code',
  description: 'AI-powered coding assistant by Anthropic',
  category: 'coding-assistant',
  official: true,
  config_schema: {
    type: 'object',
    properties: {
      model: {
        type: 'string',
        enum: ['claude-3-5-sonnet-20241022', 'claude-3-opus-20240229'],
        default: 'claude-3-5-sonnet-20241022',
      },
      max_tokens: {
        type: 'number',
        minimum: 1000,
        maximum: 100000,
        default: 4096,
      },
    },
  },
  icon_url: '/icons/claude-code.svg',
  vendor: 'Anthropic',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

export const mockAgents = [
  mockAgent,
  {
    id: '660e8400-e29b-41d4-a716-446655440002',
    name: 'Cursor',
    slug: 'cursor',
    description: 'AI-first code editor',
    category: 'code-editor',
    official: true,
    config_schema: {
      type: 'object',
      properties: {
        model: {
          type: 'string',
          enum: ['gpt-4', 'gpt-3.5-turbo'],
          default: 'gpt-4',
        },
      },
    },
    icon_url: '/icons/cursor.svg',
    vendor: 'Cursor.so',
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  },
  {
    id: '660e8400-e29b-41d4-a716-446655440003',
    name: 'Windsurf',
    slug: 'windsurf',
    description: 'AI pair programming tool',
    category: 'coding-assistant',
    official: true,
    config_schema: {
      type: 'object',
      properties: {
        enabled: {
          type: 'boolean',
          default: true,
        },
      },
    },
    icon_url: '/icons/windsurf.svg',
    vendor: 'Windsurf AI',
    created_at: '2024-01-03T00:00:00Z',
    updated_at: '2024-01-03T00:00:00Z',
  },
];

export const mockAgentConfig = {
  id: '770e8400-e29b-41d4-a716-446655440001',
  agent_id: mockAgent.id,
  agent_name: mockAgent.name,
  org_id: '550e8400-e29b-41d4-a716-446655440030',
  config: {
    model: 'claude-3-5-sonnet-20241022',
    max_tokens: 8192,
  },
  enabled: true,
  created_at: '2024-02-01T10:00:00Z',
  updated_at: '2024-02-01T10:00:00Z',
};

export const mockAgentConfigs = [
  mockAgentConfig,
  {
    id: '770e8400-e29b-41d4-a716-446655440002',
    agent_id: mockAgents[1].id,
    agent_name: mockAgents[1].name,
    org_id: '550e8400-e29b-41d4-a716-446655440030',
    config: {
      model: 'gpt-4',
    },
    enabled: false,
    created_at: '2024-02-02T10:00:00Z',
    updated_at: '2024-02-02T10:00:00Z',
  },
];
