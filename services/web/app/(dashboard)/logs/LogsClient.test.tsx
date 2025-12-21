import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LogsClient } from './LogsClient';

// Mock hooks
vi.mock('@/lib/hooks/useActivityLogs', () => ({
  useActivityLogs: vi.fn(() => ({
    logs: [],
    pagination: { total: 0, page: 1, per_page: 20, total_pages: 0 },
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  })),
}));

vi.mock('@/lib/hooks/useEmployees', () => ({
  useEmployees: vi.fn(() => ({
    data: { employees: [], total: 0 },
  })),
}));

vi.mock('@/lib/hooks/useAgents', () => ({
  useAgents: vi.fn(() => ({
    data: { agents: [], total: 0 },
  })),
}));

vi.mock('@/lib/hooks/useLogWebSocket', () => ({
  useLogWebSocket: vi.fn(() => ({
    connected: true,
    newLogs: [],
    error: null,
    clearNewLogs: vi.fn(),
  })),
}));

// Mock components
interface LogListProps {
  logs: unknown[];
  pagination: unknown;
  onPageChange: (page: number) => void;
  newLogIds?: Set<string>;
}

vi.mock('@/components/logs/LogList', () => ({
  LogList: ({ logs, pagination, onPageChange }: LogListProps) => (
    <div data-testid="log-list">
      LogList: {logs.length} logs, Page {(pagination as { page: number })?.page || 1}
      <button onClick={() => onPageChange(2)}>Next Page</button>
    </div>
  ),
}));

vi.mock('@/components/logs/ExportMenu', () => ({
  ExportMenu: () => <button>Export</button>,
}));

describe('LogsClient', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render with filters and controls', () => {
    render(<LogsClient />);

    expect(screen.getByText('Live')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /export/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/search logs/i)).toBeInTheDocument();
  });

  it('should show live indicator when WebSocket connected', () => {
    render(<LogsClient />);

    const liveIndicator = screen.getByTestId('live-indicator');
    expect(liveIndicator).toBeInTheDocument();
    expect(liveIndicator).toHaveTextContent('Live');
  });

  it('should handle search input', async () => {
    const user = userEvent.setup();
    render(<LogsClient />);

    const searchInput = screen.getByLabelText(/search logs/i);
    await user.type(searchInput, 'test search');

    expect(searchInput).toHaveValue('test search');
  });

  it('should handle clear filters', async () => {
    const user = userEvent.setup();
    render(<LogsClient />);

    // Type in search
    const searchInput = screen.getByLabelText(/search logs/i);
    await user.type(searchInput, 'test');

    // Clear filters
    const clearButton = screen.getByRole('button', { name: /clear filters/i });
    await user.click(clearButton);

    // Search should be cleared
    expect(searchInput).toHaveValue('');
  });

  it('should render LogList with pagination props', () => {
    render(<LogsClient />);

    expect(screen.getByTestId('log-list')).toBeInTheDocument();
    expect(screen.getByText(/Page 1/)).toBeInTheDocument();
  });
});
