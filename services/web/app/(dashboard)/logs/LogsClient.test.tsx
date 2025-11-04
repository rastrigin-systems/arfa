import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LogsClient } from './LogsClient';

// Mock hooks
vi.mock('@/lib/hooks/useActivityLogs', () => ({
  useActivityLogs: vi.fn(() => ({
    logs: [],
    isLoading: false,
    error: null,
    refetch: vi.fn(),
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
vi.mock('@/components/logs/LogList', () => ({
  LogList: ({ logs }: any) => <div data-testid="log-list">LogList: {logs.length} logs</div>,
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

  it.skip('should display error message on API failure', () => {
    const { useActivityLogs } = require('@/lib/hooks/useActivityLogs');
    useActivityLogs.mockReturnValue({
      logs: null,
      isLoading: false,
      error: new Error('API Error'),
      refetch: vi.fn(),
    });

    render(<LogsClient />);

    expect(screen.getByText(/failed to load logs/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
  });

  it.skip('should show loading state', () => {
    const { useActivityLogs } = require('@/lib/hooks/useActivityLogs');
    useActivityLogs.mockReturnValue({
      logs: null,
      isLoading: true,
      error: null,
      refetch: vi.fn(),
    });

    render(<LogsClient />);

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByText(/loading logs/i)).toBeInTheDocument();
  });
});
