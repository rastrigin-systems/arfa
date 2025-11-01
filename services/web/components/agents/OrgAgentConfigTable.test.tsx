import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { OrgAgentConfigTable } from './OrgAgentConfigTable';

describe('OrgAgentConfigTable', () => {
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
    {
      id: '123e4567-e89b-12d3-a456-426614174002',
      org_id: '123e4567-e89b-12d3-a456-426614174999',
      agent_id: '123e4567-e89b-12d3-a456-426614174003',
      agent_name: 'Cursor',
      agent_type: 'cursor',
      agent_provider: 'cursor',
      config: { api_key: 'sk-xxx', model: 'gpt-4' },
      is_enabled: false,
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    },
  ];

  it('should render config table with agent names', () => {
    render(<OrgAgentConfigTable configs={mockConfigs} onEdit={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText('Claude Code')).toBeInTheDocument();
    expect(screen.getByText('Cursor')).toBeInTheDocument();
  });

  it('should show empty state when no configs exist', () => {
    render(<OrgAgentConfigTable configs={[]} onEdit={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText(/no agent configurations/i)).toBeInTheDocument();
  });

  it('should display enabled status badge', () => {
    render(<OrgAgentConfigTable configs={mockConfigs} onEdit={vi.fn()} onDelete={vi.fn()} />);

    const enabledBadge = screen.getByText('Enabled');
    expect(enabledBadge).toBeInTheDocument();

    const disabledBadge = screen.getByText('Disabled');
    expect(disabledBadge).toBeInTheDocument();
  });

  it('should call onEdit when Edit button is clicked', async () => {
    const onEdit = vi.fn();
    const user = userEvent.setup();

    render(<OrgAgentConfigTable configs={mockConfigs} onEdit={onEdit} onDelete={vi.fn()} />);

    // Find the first Edit button and click it
    const editButtons = screen.getAllByRole('button', { name: /edit/i });
    await user.click(editButtons[0]);

    expect(onEdit).toHaveBeenCalledWith(mockConfigs[0]);
  });

  it('should call onDelete when Delete button is clicked', async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();

    render(<OrgAgentConfigTable configs={mockConfigs} onEdit={vi.fn()} onDelete={onDelete} />);

    // Find the first Delete button and click it
    const deleteButtons = screen.getAllByRole('button', { name: /delete/i });
    await user.click(deleteButtons[0]);

    expect(onDelete).toHaveBeenCalledWith(mockConfigs[0]);
  });

  it('should have responsive table layout', () => {
    const { container } = render(<OrgAgentConfigTable configs={mockConfigs} onEdit={vi.fn()} onDelete={vi.fn()} />);

    expect(container.querySelector('table')).toBeInTheDocument();
  });

  it('should be keyboard accessible', async () => {
    const onEdit = vi.fn();
    const user = userEvent.setup();

    render(<OrgAgentConfigTable configs={[mockConfigs[0]]} onEdit={onEdit} onDelete={vi.fn()} />);

    // Tab to first Edit button
    await user.tab();
    const editButton = screen.getByRole('button', { name: /edit/i });
    expect(editButton).toHaveFocus();

    // Activate with Enter
    await user.keyboard('{Enter}');
    expect(onEdit).toHaveBeenCalled();
  });
});
