/**
 * Mock data for E2E tests
 *
 * This file contains all mock entities used in Playwright E2E tests.
 * Mock data is based on the OpenAPI schema types.
 */

import type { components } from '../../../lib/api/schema';

// Type aliases for convenience
type Employee = components['schemas']['Employee'];
type Organization = components['schemas']['Organization'];
type Team = components['schemas']['Team'];
type Role = components['schemas']['Role'];
type Agent = components['schemas']['Agent'];
type OrgAgentConfig = components['schemas']['OrgAgentConfig'];
type TeamAgentConfig = components['schemas']['TeamAgentConfig'];
type EmployeeAgentConfig = components['schemas']['EmployeeAgentConfig'];
type ResolvedAgentConfig = components['schemas']['ResolvedAgentConfig'];

// Mock Organization
export const mockOrganization: Organization = {
  id: '01234567-89ab-cdef-0123-456789abcdef',
  name: 'ACME Corporation',
  slug: 'acme',
  plan: 'professional',
  max_employees: 50,
  max_agents_per_employee: 5,
  has_claude_token: true,
  settings: {
    theme: 'light',
    notifications: true,
  },
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-15T10:30:00Z',
};

// Mock Roles
export const mockRoles: Role[] = [
  {
    id: 'role-member-uuid',
    name: 'Member',
    description: 'Standard employee with basic permissions',
    permissions: ['read:own_profile', 'read:agents', 'sync:configs'],
    created_at: '2025-01-01T00:00:00Z',
    employee_count: 8,
  },
  {
    id: 'role-approver-uuid',
    name: 'Approver',
    description: 'Can approve agent requests and manage team configs',
    permissions: [
      'read:own_profile',
      'read:agents',
      'sync:configs',
      'approve:requests',
      'manage:team_configs',
    ],
    created_at: '2025-01-01T00:00:00Z',
    employee_count: 2,
  },
];

// Mock Teams
export const mockTeams: Team[] = [
  {
    id: 'team-engineering-uuid',
    org_id: mockOrganization.id,
    name: 'Engineering',
    description: 'Software development team',
    member_count: 5,
    agent_config_count: 2,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
  {
    id: 'team-product-uuid',
    org_id: mockOrganization.id,
    name: 'Product',
    description: 'Product management team',
    member_count: 3,
    agent_config_count: 1,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
  {
    id: 'team-design-uuid',
    org_id: mockOrganization.id,
    name: 'Design',
    description: 'UI/UX design team',
    member_count: 2,
    agent_config_count: 1,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
];

// Mock Employees
export const mockEmployees: Employee[] = [
  {
    id: 'employee-alice-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[0].id,
    team_name: mockTeams[0].name,
    role_id: mockRoles[1].id,
    email: 'alice@acme.com',
    full_name: 'Alice Johnson',
    status: 'active',
    has_personal_claude_token: true,
    preferences: {
      theme: 'dark',
      notifications: true,
    },
    last_login_at: '2025-01-15T09:00:00Z',
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T09:00:00Z',
  },
  {
    id: 'employee-bob-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[0].id,
    team_name: mockTeams[0].name,
    role_id: mockRoles[0].id,
    email: 'bob@acme.com',
    full_name: 'Bob Smith',
    status: 'active',
    has_personal_claude_token: false,
    preferences: {
      theme: 'light',
    },
    last_login_at: '2025-01-14T16:30:00Z',
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-14T16:30:00Z',
  },
  {
    id: 'employee-charlie-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[1].id,
    team_name: mockTeams[1].name,
    role_id: mockRoles[0].id,
    email: 'charlie@acme.com',
    full_name: 'Charlie Davis',
    status: 'active',
    has_personal_claude_token: false,
    preferences: {},
    last_login_at: '2025-01-15T08:15:00Z',
    created_at: '2025-01-03T00:00:00Z',
    updated_at: '2025-01-15T08:15:00Z',
  },
  {
    id: 'employee-diana-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[2].id,
    team_name: mockTeams[2].name,
    role_id: mockRoles[0].id,
    email: 'diana@acme.com',
    full_name: 'Diana Martinez',
    status: 'active',
    has_personal_claude_token: false,
    preferences: {
      theme: 'dark',
    },
    last_login_at: '2025-01-13T11:45:00Z',
    created_at: '2025-01-04T00:00:00Z',
    updated_at: '2025-01-13T11:45:00Z',
  },
  {
    id: 'employee-evan-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[0].id,
    team_name: mockTeams[0].name,
    role_id: mockRoles[0].id,
    email: 'evan@acme.com',
    full_name: 'Evan Wilson',
    status: 'suspended',
    has_personal_claude_token: false,
    preferences: {},
    last_login_at: '2025-01-10T14:20:00Z',
    created_at: '2025-01-05T00:00:00Z',
    updated_at: '2025-01-12T10:00:00Z',
  },
];

// Mock Agents (from catalog)
export const mockAgents: Agent[] = [
  {
    id: 'agent-claude-code-uuid',
    name: 'Claude Code',
    type: 'claude-code',
    description: 'AI-powered code assistant with deep codebase understanding',
    provider: 'anthropic',
    llm_provider: 'anthropic',
    llm_model: 'claude-3-5-sonnet-20241022',
    is_public: true,
    default_config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 8192,
      temperature: 0.2,
    },
    capabilities: ['code_generation', 'code_review', 'refactoring', 'debugging'],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 'agent-cursor-uuid',
    name: 'Cursor',
    type: 'cursor',
    description: 'AI code editor with context-aware suggestions',
    provider: 'cursor',
    llm_provider: 'anthropic',
    llm_model: 'claude-3-5-sonnet-20241022',
    is_public: true,
    default_config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 4096,
      autocomplete: true,
    },
    capabilities: ['autocomplete', 'code_generation', 'inline_chat'],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 'agent-windsurf-uuid',
    name: 'Windsurf',
    type: 'windsurf',
    description: 'Multi-model AI coding assistant',
    provider: 'windsurf',
    llm_provider: 'anthropic',
    llm_model: 'claude-3-5-sonnet-20241022',
    is_public: true,
    default_config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 8192,
    },
    capabilities: ['code_generation', 'multi_model_support', 'context_aware'],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 'agent-copilot-uuid',
    name: 'GitHub Copilot',
    type: 'copilot',
    description: 'AI pair programmer by GitHub',
    provider: 'github',
    llm_provider: 'openai',
    llm_model: 'gpt-4',
    is_public: true,
    default_config: {
      model: 'gpt-4',
      max_tokens: 4096,
    },
    capabilities: ['autocomplete', 'code_generation', 'documentation'],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 'agent-codeium-uuid',
    name: 'Codeium',
    type: 'codeium',
    description: 'Free AI code completion tool',
    provider: 'codeium',
    llm_provider: 'codeium',
    llm_model: 'codeium-latest',
    is_public: true,
    default_config: {
      max_tokens: 2048,
    },
    capabilities: ['autocomplete', 'code_search'],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
];

// Mock Org Agent Configs
export const mockOrgAgentConfigs: OrgAgentConfig[] = [
  {
    id: 'org-config-claude-code-uuid',
    org_id: mockOrganization.id,
    agent_id: mockAgents[0].id,
    agent_name: mockAgents[0].name,
    agent_type: mockAgents[0].type,
    agent_provider: mockAgents[0].provider,
    is_enabled: true,
    config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 16384,
      temperature: 0.3,
      system_prompt: 'You are a helpful coding assistant for ACME Corporation.',
    },
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
  {
    id: 'org-config-cursor-uuid',
    org_id: mockOrganization.id,
    agent_id: mockAgents[1].id,
    agent_name: mockAgents[1].name,
    agent_type: mockAgents[1].type,
    agent_provider: mockAgents[1].provider,
    is_enabled: true,
    config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 8192,
      autocomplete: true,
      inline_chat: true,
    },
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-10T14:20:00Z',
  },
  {
    id: 'org-config-copilot-uuid',
    org_id: mockOrganization.id,
    agent_id: mockAgents[3].id,
    agent_name: mockAgents[3].name,
    agent_type: mockAgents[3].type,
    agent_provider: mockAgents[3].provider,
    is_enabled: false,
    config: {
      model: 'gpt-4',
      max_tokens: 4096,
    },
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-08T09:15:00Z',
  },
];

// Mock Team Agent Configs (overrides)
export const mockTeamAgentConfigs: TeamAgentConfig[] = [
  {
    id: 'team-config-claude-code-uuid',
    team_id: mockTeams[0].id,
    agent_id: mockAgents[0].id,
    agent_name: mockAgents[0].name,
    agent_type: mockAgents[0].type,
    agent_provider: mockAgents[0].provider,
    is_enabled: true,
    config_override: {
      temperature: 0.1, // Override: lower temperature for engineering team
      system_prompt:
        'You are a coding assistant for the Engineering team. Focus on best practices and code quality.',
    },
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-12T15:45:00Z',
  },
];

// Mock Employee Agent Configs (overrides)
export const mockEmployeeAgentConfigs: EmployeeAgentConfig[] = [
  {
    id: 'emp-config-alice-claude-uuid',
    employee_id: mockEmployees[0].id,
    agent_id: mockAgents[0].id,
    agent_name: mockAgents[0].name,
    agent_type: mockAgents[0].type,
    agent_provider: mockAgents[0].provider,
    is_enabled: true,
    config_override: {
      max_tokens: 32768, // Alice gets more tokens
    },
    sync_token: 'sync-token-alice-123',
    last_synced_at: '2025-01-15T09:00:00Z',
    created_at: '2025-01-05T00:00:00Z',
    updated_at: '2025-01-15T09:00:00Z',
  },
];

// Mock Resolved Agent Configs (what employees actually get)
export const mockResolvedAgentConfigs: ResolvedAgentConfig[] = [
  {
    agent_id: mockAgents[0].id,
    agent_name: mockAgents[0].name,
    agent_type: mockAgents[0].type,
    provider: mockAgents[0].provider,
    is_enabled: true,
    config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 32768, // From employee override
      temperature: 0.1, // From team override
      autocomplete: true,
      inline_chat: true,
    },
    system_prompt:
      'You are a helpful coding assistant for ACME Corporation.\n\nYou are a coding assistant for the Engineering team. Focus on best practices and code quality.',
    sync_token: 'sync-token-alice-123',
    last_synced_at: '2025-01-15T09:00:00Z',
  },
  {
    agent_id: mockAgents[1].id,
    agent_name: mockAgents[1].name,
    agent_type: mockAgents[1].type,
    provider: mockAgents[1].provider,
    is_enabled: true,
    config: {
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 8192,
      autocomplete: true,
      inline_chat: true,
    },
    system_prompt: '',
    sync_token: null,
    last_synced_at: null,
  },
];

// Mock MCP Servers (for future use)
export const mockMcpServers = [
  {
    id: 'mcp-filesystem-uuid',
    name: 'Filesystem',
    category: 'file_system',
    description: 'Access local files and directories',
    docker_image: 'ubik/mcp-filesystem:latest',
  },
  {
    id: 'mcp-git-uuid',
    name: 'Git',
    category: 'version_control',
    description: 'Git repository operations',
    docker_image: 'ubik/mcp-git:latest',
  },
  {
    id: 'mcp-postgres-uuid',
    name: 'PostgreSQL',
    category: 'database',
    description: 'PostgreSQL database operations',
    docker_image: 'ubik/mcp-postgres:latest',
  },
];

// Helper function to get mock employee by email
export function getMockEmployeeByEmail(email: string): Employee | undefined {
  return mockEmployees.find((emp) => emp.email === email);
}

// Helper function to filter employees by status
export function filterEmployeesByStatus(
  status: 'active' | 'suspended' | 'inactive'
): Employee[] {
  return mockEmployees.filter((emp) => emp.status === status);
}

// Helper function to search employees
export function searchEmployees(query: string): Employee[] {
  const lowerQuery = query.toLowerCase();
  return mockEmployees.filter(
    (emp) =>
      emp.full_name.toLowerCase().includes(lowerQuery) ||
      emp.email.toLowerCase().includes(lowerQuery)
  );
}
