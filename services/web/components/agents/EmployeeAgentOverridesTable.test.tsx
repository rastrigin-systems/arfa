import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { EmployeeAgentOverridesTable } from './EmployeeAgentOverridesTable';

describe('EmployeeAgentOverridesTable', () => {
  const mockOverrides = [
    {
      id: '123e4567-e89b-12d3-a456-426614174000',
      employee_id: 'emp-001',
      agent_id: 'agent-001',
      agent_name: 'Claude Code',
      agent_type: 'claude-code',
      agent_provider: 'anthropic',
      config_override: {
        rate_limit: 200,
        cost_limit: 100,
        model: 'claude-sonnet-4.5'
      },
      override_reason: 'Senior engineer needs higher limits',
      is_enabled: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T14:23:00Z',
      updated_by: 'manager@company.com',
      org_config: {
        rate_limit: 100,
        cost_limit: 50,
        model: 'claude-sonnet-4.5'
      }
    },
    {
      id: '123e4567-e89b-12d3-a456-426614174002',
      employee_id: 'emp-001',
      agent_id: 'agent-002',
      agent_name: 'Cursor',
      agent_type: 'cursor',
      agent_provider: 'cursor',
      config_override: {
        enabled: false
      },
      override_reason: 'Security incident - temporary block',
      is_enabled: false,
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T09:15:00Z',
      updated_by: 'security@company.com',
      org_config: {
        enabled: true
      }
    },
  ];

  it('should render employee agent overrides table', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    expect(screen.getByText('Claude Code')).toBeInTheDocument();
    expect(screen.getByText('Cursor')).toBeInTheDocument();
  });

  it('should show empty state when no overrides exist', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={[]}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    expect(screen.getByText(/no employee overrides/i)).toBeInTheDocument();
    expect(screen.getByText(/using default organization/i)).toBeInTheDocument();
  });

  it('should display override reason for each config', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    expect(screen.getByText(/senior engineer needs higher limits/i)).toBeInTheDocument();
    expect(screen.getByText(/security incident/i)).toBeInTheDocument();
  });

  it('should show status badges (enabled/disabled)', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    const enabledBadge = screen.getByText('Enabled');
    expect(enabledBadge).toBeInTheDocument();

    const disabledBadge = screen.getByText('Disabled');
    expect(disabledBadge).toBeInTheDocument();
  });

  it('should display comparison with org config', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    // Should show increased rate limit
    expect(screen.getByText(/200 req\/day/i)).toBeInTheDocument();
    expect(screen.getByText(/org: 100/i)).toBeInTheDocument();
  });

  it('should show who updated the override', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    expect(screen.getByText(/manager@company.com/i)).toBeInTheDocument();
    expect(screen.getByText(/security@company.com/i)).toBeInTheDocument();
  });

  it('should call onEdit when Edit button is clicked', async () => {
    const onEdit = vi.fn();
    const user = userEvent.setup();

    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={onEdit}
        onDelete={vi.fn()}
      />
    );

    const editButtons = screen.getAllByRole('button', { name: /edit/i });
    await user.click(editButtons[0]);

    expect(onEdit).toHaveBeenCalledWith(mockOverrides[0]);
  });

  it('should call onDelete when Remove button is clicked', async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();

    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={onDelete}
      />
    );

    const deleteButtons = screen.getAllByRole('button', { name: /remove/i });
    await user.click(deleteButtons[0]);

    expect(onDelete).toHaveBeenCalledWith(mockOverrides[0]);
  });

  it('should call onToggleEnabled when Enable/Disable button is clicked', async () => {
    const onToggleEnabled = vi.fn();
    const user = userEvent.setup();

    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={onToggleEnabled}
      />
    );

    // Click disable on first override (enabled)
    const disableButton = screen.getByRole('button', { name: /disable agent/i });
    await user.click(disableButton);

    expect(onToggleEnabled).toHaveBeenCalledWith(mockOverrides[0]);
  });

  it('should be keyboard accessible', async () => {
    const onEdit = vi.fn();
    const user = userEvent.setup();

    render(
      <EmployeeAgentOverridesTable
        overrides={[mockOverrides[0]]}
        onEdit={onEdit}
        onDelete={vi.fn()}
      />
    );

    // Tab to first Edit button
    await user.tab();
    const editButton = screen.getByRole('button', { name: /edit/i });
    expect(editButton).toHaveFocus();

    // Activate with Enter
    await user.keyboard('{Enter}');
    expect(onEdit).toHaveBeenCalled();
  });

  it('should have accessible labels for action buttons', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
      />
    );

    // Use getAllByLabelText since we have multiple overrides
    expect(screen.getAllByLabelText('Edit configuration')).toHaveLength(2);
    expect(screen.getAllByLabelText('Remove configuration')).toHaveLength(2);
    expect(screen.getByLabelText('Disable agent')).toBeInTheDocument();
  });

  it('should format dates correctly', () => {
    render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    // Check for formatted dates (format: Jan 1, 2024)
    expect(screen.getByText(/Jan 1, 2024/i)).toBeInTheDocument();
    expect(screen.getByText(/Jan 2, 2024/i)).toBeInTheDocument();
  });

  it('should render responsive table layout', () => {
    const { container } = render(
      <EmployeeAgentOverridesTable
        overrides={mockOverrides}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    );

    expect(container.querySelector('table')).toBeInTheDocument();
  });
});
