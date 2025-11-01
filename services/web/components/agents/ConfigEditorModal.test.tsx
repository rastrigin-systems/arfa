import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfigEditorModal } from './ConfigEditorModal';

// Mock Monaco Editor
vi.mock('@monaco-editor/react', () => ({
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  default: ({ value, onChange }: any) => (
    <textarea
      data-testid="monaco-editor"
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
      aria-label="JSON configuration editor"
    />
  ),
}));

// Mock useToast
vi.mock('@/components/ui/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('ConfigEditorModal', () => {
  const mockAgent = {
    id: '123e4567-e89b-12d3-a456-426614174001',
    name: 'Claude Code',
    type: 'claude-code',
    provider: 'anthropic',
    default_config: { model: 'claude-3-5-sonnet-20241022', temperature: 0.2 },
  };

  const mockExistingConfig = {
    id: '123e4567-e89b-12d3-a456-426614174000',
    agent_id: mockAgent.id,
    config: { model: 'claude-3-5-sonnet-20241022', temperature: 0.5, max_tokens: 4096 },
    is_enabled: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render modal when open', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(`Configure ${mockAgent.name}`)).toBeInTheDocument();
  });

  it('should not render modal when closed', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={false}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('should initialize with agent default config when creating new config', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    const editor = screen.getByTestId('monaco-editor');
    const expectedConfig = JSON.stringify(mockAgent.default_config, null, 2);
    expect(editor).toHaveValue(expectedConfig);
  });

  it('should initialize with existing config when editing', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={mockExistingConfig}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    const editor = screen.getByTestId('monaco-editor');
    const expectedConfig = JSON.stringify(mockExistingConfig.config, null, 2);
    expect(editor).toHaveValue(expectedConfig);
  });

  it('should show "Create Configuration" button for new config', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    expect(screen.getByRole('button', { name: /create configuration/i })).toBeInTheDocument();
  });

  it('should show "Update Configuration" button for existing config', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={mockExistingConfig}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    expect(screen.getByRole('button', { name: /update configuration/i })).toBeInTheDocument();
  });

  it('should validate JSON and show error for invalid JSON', async () => {
    const user = userEvent.setup();

    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    // Enter invalid JSON
    const editor = screen.getByTestId('monaco-editor');
    await user.clear(editor);
    await user.type(editor, '{{{{ invalid json');

    // Try to submit
    const submitButton = screen.getByRole('button', { name: /create configuration/i });
    await user.click(submitButton);

    // Should show validation error in an alert role element
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/invalid json syntax/i);
    });
  });

  it('should call onSuccess after successful creation', async () => {
    const onSuccess = vi.fn();
    const user = userEvent.setup();

    // Mock successful API call
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ ...mockExistingConfig, id: 'new-config-id' }),
    });

    render(
      <ConfigEditorModal agent={mockAgent} existingConfig={null} open={true} onOpenChange={vi.fn()} onSuccess={onSuccess} />
    );

    // Submit with valid config
    const submitButton = screen.getByRole('button', { name: /create configuration/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('should call onSuccess after successful update', async () => {
    const onSuccess = vi.fn();
    const user = userEvent.setup();

    // Mock successful API call
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockExistingConfig,
    });

    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={mockExistingConfig}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={onSuccess}
      />
    );

    // Submit with valid config
    const submitButton = screen.getByRole('button', { name: /update configuration/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('should show error message when API call fails', async () => {
    const user = userEvent.setup();

    // Mock failed API call
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ message: 'Configuration already exists' }),
    });

    render(
      <ConfigEditorModal agent={mockAgent} existingConfig={null} open={true} onOpenChange={vi.fn()} onSuccess={vi.fn()} />
    );

    // Submit
    const submitButton = screen.getByRole('button', { name: /create configuration/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/failed to save configuration/i)).toBeInTheDocument();
    });
  });

  it('should close modal when Cancel button is clicked', async () => {
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={onOpenChange}
        onSuccess={vi.fn()}
      />
    );

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('should close modal when X button is clicked', async () => {
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={onOpenChange}
        onSuccess={vi.fn()}
      />
    );

    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('should be keyboard accessible', async () => {
    const user = userEvent.setup();

    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    // Should be able to tab through interactive elements
    // First tab goes to Cancel button (due to dialog focus trap)
    await user.tab();
    expect(screen.getByRole('button', { name: /cancel/i })).toHaveFocus();

    await user.tab(); // Submit button
    expect(screen.getByRole('button', { name: /create configuration/i })).toHaveFocus();

    await user.tab(); // Close button
    expect(screen.getByRole('button', { name: /close/i })).toHaveFocus();
  });

  it('should show loading state during submission', async () => {
    const user = userEvent.setup();

    // Mock slow API call
    global.fetch = vi.fn().mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(
            () =>
              resolve({
                ok: true,
                json: async () => mockExistingConfig,
              }),
            100
          )
        )
    );

    render(
      <ConfigEditorModal agent={mockAgent} existingConfig={null} open={true} onOpenChange={vi.fn()} onSuccess={vi.fn()} />
    );

    const submitButton = screen.getByRole('button', { name: /create configuration/i });
    await user.click(submitButton);

    // Should show loading state
    expect(submitButton).toBeDisabled();
    expect(screen.getByText(/saving/i)).toBeInTheDocument();

    // Wait for completion
    await waitFor(
      () => {
        expect(submitButton).not.toBeDisabled();
      },
      { timeout: 200 }
    );
  });

  it('should have proper aria labels for accessibility', () => {
    render(
      <ConfigEditorModal
        agent={mockAgent}
        existingConfig={null}
        open={true}
        onOpenChange={vi.fn()}
        onSuccess={vi.fn()}
      />
    );

    expect(screen.getByRole('dialog')).toHaveAttribute('aria-labelledby');
    expect(screen.getByRole('dialog')).toHaveAttribute('aria-describedby');
    expect(screen.getByTestId('monaco-editor')).toHaveAttribute('aria-label');
  });
});
