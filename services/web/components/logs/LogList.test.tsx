import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LogList } from './LogList';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];

const mockLogs: ActivityLog[] = [
  {
    id: '1',
    org_id: 'org-1',
    session_id: 'session-1',
    employee_id: 'emp-1',
    employee_name: 'John Doe',
    agent_id: 'agent-1',
    agent_name: 'claude-code',
    event_type: 'session_start',
    event_category: 'io',
    content: 'Session started',
    created_at: '2024-01-01T10:00:00Z',
  },
  {
    id: '2',
    org_id: 'org-1',
    session_id: 'session-1',
    employee_id: 'emp-1',
    employee_name: 'John Doe',
    agent_id: 'agent-1',
    agent_name: 'claude-code',
    event_type: 'input',
    event_category: 'io',
    content: 'Implement JWT authentication',
    created_at: '2024-01-01T10:00:10Z',
  },
  {
    id: '3',
    org_id: 'org-1',
    session_id: 'session-1',
    employee_id: 'emp-1',
    employee_name: 'John Doe',
    agent_id: 'agent-1',
    agent_name: 'claude-code',
    event_type: 'output',
    event_category: 'io',
    content: 'I will help you implement JWT authentication...',
    created_at: '2024-01-01T10:00:15Z',
  },
  {
    id: '4',
    org_id: 'org-1',
    session_id: 'session-1',
    employee_id: 'emp-1',
    employee_name: 'John Doe',
    agent_id: 'agent-1',
    agent_name: 'claude-code',
    event_type: 'session_end',
    event_category: 'io',
    content: 'Session ended',
    created_at: '2024-01-01T10:05:30Z',
  },
];

describe('LogList', () => {
  it('should render empty state when no logs', () => {
    render(<LogList logs={[]} />);

    expect(screen.getByText('No logs found')).toBeInTheDocument();
    expect(screen.getByText(/try adjusting your filters/i)).toBeInTheDocument();
  });

  it('should group logs by session', () => {
    render(<LogList logs={mockLogs} />);

    // Check for session text
    expect(screen.getByText(/Session:/)).toBeInTheDocument();
    // Check for event count badge
    expect(screen.getByText('4 events')).toBeInTheDocument();
    // Verify session details are rendered
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('claude-code')).toBeInTheDocument();
  });

  it('should display session metadata', () => {
    render(<LogList logs={mockLogs} />);

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('claude-code')).toBeInTheDocument();
    expect(screen.getByText('5m 30s')).toBeInTheDocument(); // Duration
  });

  it('should expand/collapse session on click', async () => {
    const user = userEvent.setup();
    render(<LogList logs={mockLogs} />);

    // Find the clickable session header div
    const sessionCard = screen.getByText(/Session:/i).closest('[class*="cursor-pointer"]');
    expect(sessionCard).toBeInTheDocument();

    // Initially collapsed
    expect(screen.queryByText('Session started')).not.toBeInTheDocument();

    // Click to expand
    await user.click(sessionCard!);
    await waitFor(() => {
      expect(screen.getByText('Session started')).toBeInTheDocument();
      expect(screen.getByText('Implement JWT authentication')).toBeInTheDocument();
    });

    // Click to collapse
    await user.click(sessionCard!);
    await waitFor(() => {
      expect(screen.queryByText('Session started')).not.toBeInTheDocument();
    });
  });

  it('should display event type badges', async () => {
    const user = userEvent.setup();
    render(<LogList logs={mockLogs} />);

    const sessionCard = screen.getByText(/Session:/i).closest('[class*="cursor-pointer"]');
    await user.click(sessionCard!);

    await waitFor(() => {
      expect(screen.getByText('session_start')).toBeInTheDocument();
      expect(screen.getByText('input')).toBeInTheDocument();
      expect(screen.getByText('output')).toBeInTheDocument();
      expect(screen.getByText('session_end')).toBeInTheDocument();
    });
  });

  it('should show expandable log content', async () => {
    const user = userEvent.setup();
    const longContent = 'A'.repeat(200);
    const logsWithLongContent: ActivityLog[] = [
      {
        ...mockLogs[0],
        content: longContent,
      },
    ];

    render(<LogList logs={logsWithLongContent} />);

    const sessionCard = screen.getByText(/Session:/i).closest('[class*="cursor-pointer"]');
    await user.click(sessionCard!);

    await waitFor(() => {
      expect(screen.getByText('Show more')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Show more'));
    expect(screen.getByText('Show less')).toBeInTheDocument();
  });

  it('should display metadata in details', async () => {
    const user = userEvent.setup();
    const logsWithMetadata: ActivityLog[] = [
      {
        ...mockLogs[0],
        metadata: { workspace: '/path/to/workspace', command: 'ubik' },
      },
    ];

    render(<LogList logs={logsWithMetadata} />);

    const sessionCard = screen.getByText(/Session:/i).closest('[class*="cursor-pointer"]');
    await user.click(sessionCard!);

    await waitFor(() => {
      const metadataToggle = screen.getByText('Metadata');
      expect(metadataToggle).toBeInTheDocument();
    });
  });
});
