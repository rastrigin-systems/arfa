import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import OrgAgentConfigsPage from './page';
import * as authModule from '@/lib/auth';

// Mock modules
vi.mock('@/lib/auth');
vi.mock('@/lib/api/client', () => ({
  apiClient: {
    GET: vi.fn(),
  },
}));

// Mock child component
vi.mock('./OrgAgentConfigsClient', () => ({
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  OrgAgentConfigsClient: ({ initialConfigs, initialAgents }: any) => (
    <div data-testid="org-agent-configs-client">
      <div data-testid="configs-count">{initialConfigs.length}</div>
      <div data-testid="agents-count">{initialAgents.length}</div>
    </div>
  ),
}));

describe('OrgAgentConfigsPage', () => {
  const mockToken = 'mock-jwt-token';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch and pass organization agent configs to client component', async () => {
    // Arrange
    vi.mocked(authModule.getServerToken).mockResolvedValue(mockToken);

    const mockConfigs = [
      {
        id: '123e4567-e89b-12d3-a456-426614174000',
        org_id: '123e4567-e89b-12d3-a456-426614174999',
        agent_id: '123e4567-e89b-12d3-a456-426614174001',
        agent_name: 'Claude Code',
        agent_type: 'claude-code',
        agent_provider: 'anthropic',
        config: { model: 'claude-3-5-sonnet-20241022', temperature: 0.2 },
        is_enabled: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ];

    const mockAgents = [
      {
        id: '123e4567-e89b-12d3-a456-426614174001',
        name: 'Claude Code',
        type: 'claude-code',
        description: 'AI-powered code assistant',
        provider: 'anthropic',
        llm_provider: 'anthropic',
        llm_model: 'claude-3-5-sonnet-20241022',
        is_public: true,
      },
    ];

    const { apiClient } = await import('@/lib/api/client');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(apiClient.GET).mockImplementation((path: string): any => {
      if (path === '/organizations/current/agent-configs') {
        return Promise.resolve({
          data: { configs: mockConfigs },
          error: undefined,
          response: new Response(),
        });
      }
      if (path === '/agents') {
        return Promise.resolve({
          data: { agents: mockAgents },
          error: undefined,
          response: new Response(),
        });
      }
      return Promise.resolve({ data: null, error: undefined, response: new Response() });
    });

    // Act
    render(await OrgAgentConfigsPage());

    // Assert
    await waitFor(() => {
      expect(screen.getByTestId('org-agent-configs-client')).toBeInTheDocument();
    });

    expect(screen.getByTestId('configs-count')).toHaveTextContent('1');
    expect(screen.getByTestId('agents-count')).toHaveTextContent('1');
  });

  it('should throw error when user is not authenticated', async () => {
    // Arrange
    vi.mocked(authModule.getServerToken).mockResolvedValue(null);

    // Act & Assert
    await expect(OrgAgentConfigsPage()).rejects.toThrow('Unauthorized');
  });

  it('should throw error when API call fails', async () => {
    // Arrange
    vi.mocked(authModule.getServerToken).mockResolvedValue(mockToken);

    const { apiClient } = await import('@/lib/api/client');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: null,
      error: { message: 'Failed to load configs' },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    // Act & Assert
    await expect(OrgAgentConfigsPage()).rejects.toThrow('Failed to load organization agent configurations');
  });

  it('should render page header with title and description', async () => {
    // Arrange
    vi.mocked(authModule.getServerToken).mockResolvedValue(mockToken);

    const { apiClient } = await import('@/lib/api/client');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { configs: [], agents: [] },
      error: undefined,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    // Act
    render(await OrgAgentConfigsPage());

    // Assert
    expect(screen.getByText('Organization Agent Configurations')).toBeInTheDocument();
    expect(
      screen.getByText(/Configure AI agents at the organization level to make them available to your teams/i)
    ).toBeInTheDocument();
  });

  it('should handle empty configs list', async () => {
    // Arrange
    vi.mocked(authModule.getServerToken).mockResolvedValue(mockToken);

    const { apiClient } = await import('@/lib/api/client');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(apiClient.GET).mockImplementation((path: string): any => {
      if (path === '/organizations/current/agent-configs') {
        return Promise.resolve({
          data: { configs: [] },
          error: undefined,
          response: new Response(),
        });
      }
      if (path === '/agents') {
        return Promise.resolve({
          data: { agents: [] },
          error: undefined,
          response: new Response(),
        });
      }
      return Promise.resolve({ data: null, error: undefined, response: new Response() });
    });

    // Act
    render(await OrgAgentConfigsPage());

    // Assert
    await waitFor(() => {
      expect(screen.getByTestId('configs-count')).toHaveTextContent('0');
      expect(screen.getByTestId('agents-count')).toHaveTextContent('0');
    });
  });
});
