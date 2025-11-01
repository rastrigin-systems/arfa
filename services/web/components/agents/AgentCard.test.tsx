import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AgentCard } from './AgentCard';

describe('AgentCard', () => {
  const mockAgent = {
    id: '123e4567-e89b-12d3-a456-426614174000',
    name: 'Claude Code',
    type: 'claude-code',
    description: 'AI-powered code assistant with deep codebase understanding',
    provider: 'anthropic',
    llm_provider: 'anthropic',
    llm_model: 'claude-3-5-sonnet-20241022',
    is_public: true,
    capabilities: ['code_generation', 'code_review', 'refactoring'],
    default_config: { model: 'claude-3-5-sonnet-20241022', max_tokens: 8192 },
  };

  it('should render agent details', () => {
    render(<AgentCard agent={mockAgent} isEnabled={false} />);

    expect(screen.getByText('Claude Code')).toBeInTheDocument();
    expect(screen.getByText(/AI-powered code assistant/)).toBeInTheDocument();
    expect(screen.getByText('anthropic')).toBeInTheDocument();
  });

  it('should show "Enable for Org" button when agent is not enabled', () => {
    render(<AgentCard agent={mockAgent} isEnabled={false} />);

    const button = screen.getByRole('button', { name: /Enable for Org/i });
    expect(button).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Configure/i })).not.toBeInTheDocument();
  });

  it('should show "Configure" button when agent is enabled', () => {
    render(<AgentCard agent={mockAgent} isEnabled={true} />);

    const button = screen.getByRole('button', { name: /Configure/i });
    expect(button).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Enable for Org/i })).not.toBeInTheDocument();
  });

  it('should call onEnable when Enable button is clicked', async () => {
    const onEnable = vi.fn();
    const user = userEvent.setup();

    render(<AgentCard agent={mockAgent} isEnabled={false} onEnable={onEnable} />);

    const button = screen.getByRole('button', { name: /Enable for Org/i });
    await user.click(button);

    expect(onEnable).toHaveBeenCalledWith(mockAgent.id);
  });

  it('should call onConfigure when Configure button is clicked', async () => {
    const onConfigure = vi.fn();
    const user = userEvent.setup();

    render(<AgentCard agent={mockAgent} isEnabled={true} onConfigure={onConfigure} />);

    const button = screen.getByRole('button', { name: /Configure/i });
    await user.click(button);

    expect(onConfigure).toHaveBeenCalledWith(mockAgent.id);
  });

  it('should display capabilities as tags', () => {
    render(<AgentCard agent={mockAgent} isEnabled={false} />);

    expect(screen.getByText('code_generation')).toBeInTheDocument();
    expect(screen.getByText('code_review')).toBeInTheDocument();
    expect(screen.getByText('refactoring')).toBeInTheDocument();
  });

  it('should be keyboard accessible', async () => {
    const onEnable = vi.fn();
    const user = userEvent.setup();

    render(<AgentCard agent={mockAgent} isEnabled={false} onEnable={onEnable} />);

    const button = screen.getByRole('button', { name: /Enable for Org/i });

    // Tab to the button
    await user.tab();
    expect(button).toHaveFocus();

    // Press Enter to activate
    await user.keyboard('{Enter}');
    expect(onEnable).toHaveBeenCalledWith(mockAgent.id);
  });

  it('should have proper ARIA labels', () => {
    render(<AgentCard agent={mockAgent} isEnabled={false} />);

    const card = screen.getByRole('article');
    expect(card).toHaveAccessibleName('Claude Code agent card');
  });
});
