import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import LogsPage from './page';

// Mock the LogsClient component
vi.mock('./LogsClient', () => ({
  LogsClient: () => <div data-testid="logs-client">Mocked LogsClient</div>,
}));

describe('LogsPage', () => {
  it('should render logs page with title', () => {
    render(<LogsPage />);

    expect(screen.getByText('Activity Logs')).toBeInTheDocument();
  });

  it('should render logs client component', () => {
    render(<LogsPage />);

    expect(screen.getByTestId('logs-client')).toBeInTheDocument();
  });

  it('should have responsive container', () => {
    render(<LogsPage />);

    const container = screen.getByTestId('logs-container');
    expect(container).toHaveClass('space-y-6');
  });
});
