import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import EmployeeAgentConfigsPage from './page';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    refresh: vi.fn(),
  }),
  useParams: () => ({
    id: 'emp-123',
  }),
}));

// Mock the API client
vi.mock('@/lib/api/client', () => ({
  apiClient: {
    GET: vi.fn(),
    POST: vi.fn(),
    PATCH: vi.fn(),
    DELETE: vi.fn(),
  },
}));

// Mock auth
vi.mock('@/lib/auth', () => ({
  getServerSession: vi.fn(() =>
    Promise.resolve({
      user: { id: 'user-123', email: 'test@example.com' },
      accessToken: 'mock-token',
    })
  ),
}));

describe('EmployeeAgentConfigsPage', () => {
  const mockEmployee = {
    id: 'emp-123',
    email: 'employee@example.com',
    full_name: 'Test Employee',
    status: 'active',
    org_id: 'org-123',
    role_id: 'role-123',
  };

  const mockAgentConfigs = [
    {
      id: 'config-1',
      employee_id: 'emp-123',
      agent_id: 'agent-1',
      agent_name: 'Claude Code',
      agent_type: 'ai-assistant',
      agent_provider: 'Anthropic',
      config_override: { temperature: 0.5 },
      is_enabled: true,
      sync_token: 'token-123',
      last_synced_at: '2024-01-01T00:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'config-2',
      employee_id: 'emp-123',
      agent_id: 'agent-2',
      agent_name: 'Cursor',
      agent_type: 'ai-assistant',
      agent_provider: 'Cursor Inc',
      config_override: {},
      is_enabled: false,
      sync_token: null,
      last_synced_at: null,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ];

  const mockAgents = [
    {
      id: 'agent-1',
      name: 'Claude Code',
      type: 'ai-assistant',
      description: 'AI coding assistant',
      provider: 'Anthropic',
      llm_provider: 'anthropic',
      llm_model: 'claude-3-opus',
      is_public: true,
      capabilities: ['code', 'chat'],
      default_config: { temperature: 0.7 },
    },
    {
      id: 'agent-3',
      name: 'GitHub Copilot',
      type: 'ai-assistant',
      description: 'AI pair programmer',
      provider: 'GitHub',
      llm_provider: 'openai',
      llm_model: 'gpt-4',
      is_public: true,
      capabilities: ['code'],
      default_config: {},
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render employee agent configurations page', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    expect(screen.getByText(/agent configurations/i)).toBeInTheDocument();
  });

  it('should display list of configured agents', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    await waitFor(() => {
      expect(screen.getByText('Claude Code')).toBeInTheDocument();
      expect(screen.getByText('Cursor')).toBeInTheDocument();
    });
  });

  it('should show enabled/disabled status for each agent', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    await waitFor(() => {
      const enabledBadges = screen.getAllByText(/enabled/i);
      const disabledBadges = screen.getAllByText(/disabled/i);

      expect(enabledBadges.length).toBeGreaterThan(0);
      expect(disabledBadges.length).toBeGreaterThan(0);
    });
  });

  it('should have an "Add Configuration" button', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    const addButton = screen.getByRole('button', { name: /add configuration/i });
    expect(addButton).toBeInTheDocument();
  });

  it('should display sync token and last sync time for configured agents', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    await waitFor(() => {
      expect(screen.getByText(/token-123/i)).toBeInTheDocument();
      expect(screen.getByText(/2 hours ago/i)).toBeInTheDocument();
    });
  });

  it('should show configuration override JSON for each agent', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    await waitFor(() => {
      expect(screen.getByText(/"temperature": 0.5/i)).toBeInTheDocument();
    });
  });

  it('should have edit and delete actions for each configuration', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    await waitFor(() => {
      const editButtons = screen.getAllByRole('button', { name: /edit/i });
      const deleteButtons = screen.getAllByRole('button', { name: /delete/i });

      expect(editButtons.length).toBeGreaterThan(0);
      expect(deleteButtons.length).toBeGreaterThan(0);
    });
  });

  it('should display available agents in catalog view', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    const availableTab = screen.getByRole('tab', { name: /available agents/i });
    await userEvent.click(availableTab);

    await waitFor(() => {
      expect(screen.getByText('GitHub Copilot')).toBeInTheDocument();
    });
  });

  it('should show breadcrumb navigation', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'emp-123' } });
    render(page);

    expect(screen.getByText(/employees/i)).toBeInTheDocument();
    expect(screen.getByText(/test employee/i)).toBeInTheDocument();
    expect(screen.getByText(/agents/i)).toBeInTheDocument();
  });

  it('should handle missing employee gracefully', async () => {
    const page = await EmployeeAgentConfigsPage({ params: { id: 'invalid-id' } });
    render(page);

    await waitFor(() => {
      expect(screen.getByText(/employee not found/i)).toBeInTheDocument();
    });
  });
});
