import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AgentCatalog } from './AgentCatalog';

describe('AgentCatalog', () => {
  const mockAgents = [
    {
      id: '123e4567-e89b-12d3-a456-426614174000',
      name: 'Claude Code',
      type: 'claude-code',
      description: 'AI-powered code assistant',
      provider: 'anthropic',
      llm_provider: 'anthropic',
      llm_model: 'claude-3-5-sonnet-20241022',
      is_public: true,
      capabilities: ['code_generation', 'code_review'],
      default_config: {},
    },
    {
      id: '123e4567-e89b-12d3-a456-426614174001',
      name: 'Cursor',
      type: 'cursor',
      description: 'AI code editor',
      provider: 'cursor',
      llm_provider: 'openai',
      llm_model: 'gpt-4',
      is_public: true,
      capabilities: ['code_generation'],
      default_config: {},
    },
  ];

  const mockEnabledAgentIds = new Set(['123e4567-e89b-12d3-a456-426614174000']);

  it('should render all agents in a grid', () => {
    render(<AgentCatalog agents={mockAgents} enabledAgentIds={mockEnabledAgentIds} />);

    expect(screen.getByText('Claude Code')).toBeInTheDocument();
    expect(screen.getByText('Cursor')).toBeInTheDocument();
  });

  it('should show loading skeleton when loading', () => {
    render(<AgentCatalog agents={[]} enabledAgentIds={new Set()} isLoading={true} />);

    const skeletons = screen.getAllByRole('status', { name: /loading/i });
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('should show error state with retry button when error occurs', () => {
    const onRetry = vi.fn();
    render(<AgentCatalog agents={[]} enabledAgentIds={new Set()} error={new Error('Failed to load')} onRetry={onRetry} />);

    expect(screen.getByText(/Failed to load agents/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
  });

  it('should call onRetry when retry button is clicked', async () => {
    const onRetry = vi.fn();
    const user = userEvent.setup();

    render(<AgentCatalog agents={[]} enabledAgentIds={new Set()} error={new Error('Failed')} onRetry={onRetry} />);

    const retryButton = screen.getByRole('button', { name: /retry/i });
    await user.click(retryButton);

    expect(onRetry).toHaveBeenCalled();
  });

  it('should show empty state when no agents available', () => {
    render(<AgentCatalog agents={[]} enabledAgentIds={new Set()} />);

    expect(screen.getByText(/No agents available/i)).toBeInTheDocument();
  });

  it('should pass correct isEnabled state to AgentCard components', () => {
    render(<AgentCatalog agents={mockAgents} enabledAgentIds={mockEnabledAgentIds} />);

    // Claude Code is enabled - should show "Configure Claude Code"
    const claudeCard = screen.getByLabelText('Claude Code agent card');
    expect(within(claudeCard).getByRole('button', { name: 'Configure Claude Code' })).toBeInTheDocument();

    // Cursor is not enabled - should show "Enable Cursor for organization"
    const cursorCard = screen.getByLabelText('Cursor agent card');
    expect(within(cursorCard).getByRole('button', { name: 'Enable Cursor for organization' })).toBeInTheDocument();
  });

  it('should call onEnable when Enable button is clicked', async () => {
    const onEnable = vi.fn();
    const user = userEvent.setup();

    render(<AgentCatalog agents={mockAgents} enabledAgentIds={mockEnabledAgentIds} onEnable={onEnable} />);

    // Click "Enable Cursor for organization" on Cursor (not enabled)
    const cursorCard = screen.getByLabelText('Cursor agent card');
    const enableButton = within(cursorCard).getByRole('button', { name: 'Enable Cursor for organization' });
    await user.click(enableButton);

    expect(onEnable).toHaveBeenCalledWith('123e4567-e89b-12d3-a456-426614174001');
  });

  it('should call onConfigure when Configure button is clicked', async () => {
    const onConfigure = vi.fn();
    const user = userEvent.setup();

    render(<AgentCatalog agents={mockAgents} enabledAgentIds={mockEnabledAgentIds} onConfigure={onConfigure} />);

    // Click "Configure Claude Code" on Claude Code (enabled)
    const claudeCard = screen.getByLabelText('Claude Code agent card');
    const configureButton = within(claudeCard).getByRole('button', { name: 'Configure Claude Code' });
    await user.click(configureButton);

    expect(onConfigure).toHaveBeenCalledWith('123e4567-e89b-12d3-a456-426614174000');
  });

  it('should have responsive grid layout classes', () => {
    const { container } = render(<AgentCatalog agents={mockAgents} enabledAgentIds={mockEnabledAgentIds} />);

    const grid = container.querySelector('[class*="grid"]');
    expect(grid).toBeInTheDocument();
    expect(grid?.className).toMatch(/grid-cols-1/); // mobile
    expect(grid?.className).toMatch(/md:grid-cols-2/); // tablet
    expect(grid?.className).toMatch(/lg:grid-cols-3/); // desktop
  });

  it('should be keyboard accessible', async () => {
    const onEnable = vi.fn();
    const user = userEvent.setup();

    render(<AgentCatalog agents={[mockAgents[1]]} enabledAgentIds={new Set()} onEnable={onEnable} />);

    // Tab through to the button
    await user.tab();
    const button = screen.getByRole('button', { name: 'Enable Cursor for organization' });
    expect(button).toHaveFocus();

    // Activate with Enter
    await user.keyboard('{Enter}');
    expect(onEnable).toHaveBeenCalled();
  });
});
