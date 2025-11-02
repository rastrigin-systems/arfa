import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import EmployeeAgentOverridesPage from './page';

// Mock the Next.js navigation
vi.mock('next/navigation', () => ({
  notFound: vi.fn(),
  useRouter: () => ({
    push: vi.fn(),
    back: vi.fn(),
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
  getServerToken: vi.fn(() => Promise.resolve('mock-token')),
}));

describe('EmployeeAgentOverridesPage', () => {
  const mockEmployee = {
    id: 'emp-001',
    org_id: 'org-001',
    team_id: 'team-001',
    team_name: 'Engineering',
    role_id: 'role-001',
    email: 'john.doe@company.com',
    full_name: 'John Doe',
    status: 'active' as const,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  const mockOverrides = [
    {
      id: 'override-001',
      employee_id: 'emp-001',
      agent_id: 'agent-001',
      agent_name: 'Claude Code',
      agent_type: 'claude-code',
      agent_provider: 'anthropic',
      config_override: {
        rate_limit: 200,
        cost_limit: 100,
      },
      override_reason: 'Senior engineer needs higher limits',
      is_enabled: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T14:23:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render employee agent overrides page', async () => {
    const { apiClient } = await import('@/lib/api/client');
    (apiClient.GET as ReturnType<typeof vi.fn>).mockImplementation((path: string) => {
      if (path === '/employees/{employee_id}') {
        return Promise.resolve({ data: mockEmployee, error: null });
      }
      if (path === '/employees/{employee_id}/agent-configs') {
        return Promise.resolve({ data: { configs: mockOverrides }, error: null });
      }
      return Promise.resolve({ data: null, error: null });
    });

    render(await EmployeeAgentOverridesPage({ params: { id: 'emp-001' } }));

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getAllByText(/agent overrides/i).length).toBeGreaterThan(0);
  });

  it('should display employee overrides table', async () => {
    const { apiClient } = await import('@/lib/api/client');
    (apiClient.GET as ReturnType<typeof vi.fn>).mockImplementation((path: string) => {
      if (path === '/employees/{employee_id}') {
        return Promise.resolve({ data: mockEmployee, error: null });
      }
      if (path === '/employees/{employee_id}/agent-configs') {
        return Promise.resolve({ data: { configs: mockOverrides }, error: null });
      }
      return Promise.resolve({ data: null, error: null });
    });

    render(await EmployeeAgentOverridesPage({ params: { id: 'emp-001' } }));

    expect(screen.getByText('Claude Code')).toBeInTheDocument();
    expect(screen.getByText(/senior engineer needs higher limits/i)).toBeInTheDocument();
  });

  it('should show empty state when no overrides exist', async () => {
    const { apiClient } = await import('@/lib/api/client');
    (apiClient.GET as ReturnType<typeof vi.fn>).mockImplementation((path: string) => {
      if (path === '/employees/{employee_id}') {
        return Promise.resolve({ data: mockEmployee, error: null });
      }
      if (path === '/employees/{employee_id}/agent-configs') {
        return Promise.resolve({ data: { configs: [] }, error: null });
      }
      return Promise.resolve({ data: null, error: null });
    });

    render(await EmployeeAgentOverridesPage({ params: { id: 'emp-001' } }));

    expect(screen.getByText(/no employee overrides/i)).toBeInTheDocument();
  });

  it('should handle API errors gracefully', async () => {
    const { apiClient } = await import('@/lib/api/client');
    const { notFound } = await import('next/navigation');

    (apiClient.GET as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: null,
      error: { message: 'Not found' },
    });

    await EmployeeAgentOverridesPage({ params: { id: 'non-existent' } });

    expect(notFound).toHaveBeenCalled();
  });

  it('should display breadcrumb navigation', async () => {
    const { apiClient } = await import('@/lib/api/client');
    (apiClient.GET as ReturnType<typeof vi.fn>).mockImplementation((path: string) => {
      if (path === '/employees/{employee_id}') {
        return Promise.resolve({ data: mockEmployee, error: null });
      }
      if (path === '/employees/{employee_id}/agent-configs') {
        return Promise.resolve({ data: { configs: mockOverrides }, error: null });
      }
      return Promise.resolve({ data: null, error: null });
    });

    render(await EmployeeAgentOverridesPage({ params: { id: 'emp-001' } }));

    expect(screen.getByText('Employees')).toBeInTheDocument();
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Agent Overrides')).toBeInTheDocument();
  });

  it('should have back button to employee detail', async () => {
    const { apiClient } = await import('@/lib/api/client');
    (apiClient.GET as ReturnType<typeof vi.fn>).mockImplementation((path: string) => {
      if (path === '/employees/{employee_id}') {
        return Promise.resolve({ data: mockEmployee, error: null });
      }
      if (path === '/employees/{employee_id}/agent-configs') {
        return Promise.resolve({ data: { configs: mockOverrides }, error: null });
      }
      return Promise.resolve({ data: null, error: null });
    });

    render(await EmployeeAgentOverridesPage({ params: { id: 'emp-001' } }));

    const backLink = screen.getByRole('link', { name: /back to employee/i });
    expect(backLink).toBeInTheDocument();
    expect(backLink).toHaveAttribute('href', '/employees/emp-001');
  });
});
