import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LogList } from './LogList';
import type { components } from '@/lib/api/schema';

type ActivityLog = components['schemas']['ActivityLog'];
type PaginationMeta = components['schemas']['PaginationMeta'];

const mockLogs: ActivityLog[] = [
  {
    id: '1',
    org_id: 'org-1',
    session_id: 'session-1',
    employee_id: 'emp-1',
    agent_id: 'agent-1',
    event_type: 'session_start',
    event_category: 'io',
    content: 'Session started',
    payload: {},
    created_at: '2024-01-01T10:00:00Z',
  },
  {
    id: '2',
    org_id: 'org-1',
    session_id: 'session-1',
    employee_id: 'emp-1',
    agent_id: 'agent-1',
    event_type: 'input',
    event_category: 'io',
    content: 'This is a very long content that should be expandable when clicked because it exceeds the threshold',
    payload: { command: 'test' },
    created_at: '2024-01-01T10:00:10Z',
  },
];

const mockPagination: PaginationMeta = {
  total: 100,
  page: 1,
  per_page: 20,
  total_pages: 5,
};

describe('LogList', () => {
  it('should render empty state when no logs', () => {
    render(<LogList logs={[]} pagination={null} onPageChange={vi.fn()} />);

    expect(screen.getByText('No logs found')).toBeInTheDocument();
    expect(screen.getByText(/try adjusting your filters/i)).toBeInTheDocument();
  });

  it('should render logs in table format', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    // Check table headers
    expect(screen.getByText('Time')).toBeInTheDocument();
    expect(screen.getByText('Event Type')).toBeInTheDocument();
    expect(screen.getByText('Category')).toBeInTheDocument();
    expect(screen.getByText('Session')).toBeInTheDocument();
    expect(screen.getByText('Content')).toBeInTheDocument();
  });

  it('should display session_id as truncated text', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    // Session ID should be truncated to first 8 characters
    expect(screen.getAllByText('session-').length).toBeGreaterThan(0);
  });

  it('should display event type badges', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    expect(screen.getByText('session_start')).toBeInTheDocument();
    expect(screen.getByText('input')).toBeInTheDocument();
  });

  it('should display event category badges', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    // There should be 'io' category badges
    const ioBadges = screen.getAllByText('io');
    expect(ioBadges.length).toBeGreaterThan(0);
  });

  it('should expand row to show full content on click', async () => {
    const user = userEvent.setup();
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    // Find the expandable row (has long content)
    const expandButton = screen.getAllByRole('button')[0];
    await user.click(expandButton);

    // Should show expanded content
    await waitFor(() => {
      expect(screen.getByText('Payload')).toBeInTheDocument();
    });
  });

  it('should show pagination controls when multiple pages', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    expect(screen.getByLabelText('Next page')).toBeInTheDocument();
    expect(screen.getByLabelText('Previous page')).toBeInTheDocument();
  });

  it('should not show pagination when only one page', () => {
    const singlePagePagination: PaginationMeta = {
      total: 5,
      page: 1,
      per_page: 20,
      total_pages: 1,
    };

    render(
      <LogList
        logs={mockLogs}
        pagination={singlePagePagination}
        onPageChange={vi.fn()}
      />
    );

    expect(screen.queryByLabelText('Next page')).not.toBeInTheDocument();
  });

  it('should call onPageChange when pagination is clicked', async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();

    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={onPageChange}
      />
    );

    await user.click(screen.getByLabelText('Next page'));
    expect(onPageChange).toHaveBeenCalledWith(2);
  });

  it('should highlight new logs with green background', () => {
    const newLogIds = new Set(['1']);

    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
        newLogIds={newLogIds}
      />
    );

    // The row with id '1' should have green background class
    const rows = screen.getAllByRole('row');
    // First row is header, second row should be highlighted
    expect(rows[1]).toHaveClass('bg-green-50');
  });

  it('should show pagination summary', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    expect(screen.getByText(/Showing 1 - 20 of 100 logs/)).toBeInTheDocument();
  });

  it('should display content preview in table row', () => {
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    expect(screen.getByText('Session started')).toBeInTheDocument();
  });

  it('should show payload in expanded view', async () => {
    const user = userEvent.setup();
    render(
      <LogList
        logs={mockLogs}
        pagination={mockPagination}
        onPageChange={vi.fn()}
      />
    );

    // Click the row with payload (second log)
    const rows = screen.getAllByRole('row');
    await user.click(rows[2]); // Row index 2 is the second data row

    await waitFor(() => {
      expect(screen.getByText(/"command": "test"/)).toBeInTheDocument();
    });
  });
});
