import { describe, it, expect, beforeEach, mock, spyOn } from 'bun:test';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Create mock function
const mockApiClientGET = mock(() => Promise.resolve({ data: '[]', error: undefined, response: {} }));

mock.module('@/lib/api/client', () => ({
  apiClient: {
    GET: mockApiClientGET,
  },
}));

// Import after mocking
import { ExportMenu } from './ExportMenu';

describe.skip('ExportMenu', () => {
  beforeEach(() => {
    mockApiClientGET.mockClear();

    // Mock URL.createObjectURL and document methods
    global.URL.createObjectURL = mock(() => 'blob:mock-url');
    global.URL.revokeObjectURL = mock(() => {});

    // Mock link click
    const mockLink = document.createElement('a');
    mockLink.click = mock(() => {});
    spyOn(document, 'createElement').mockReturnValue(mockLink as unknown as HTMLElement);
    spyOn(document.body, 'appendChild').mockImplementation(() => mockLink as Node);
    spyOn(document.body, 'removeChild').mockImplementation(() => mockLink as Node);
  });

  it('should render format selector and export button', () => {
    render(<ExportMenu filters={{}} />);

    expect(screen.getByRole('combobox')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /export/i })).toBeInTheDocument();
  });

  it('should allow changing export format', async () => {
    const user = userEvent.setup();
    render(<ExportMenu filters={{}} />);

    const formatSelect = screen.getByRole('combobox');
    await user.click(formatSelect);

    // Should show JSON and CSV options
    await waitFor(() => {
      expect(screen.getByText('CSV')).toBeInTheDocument();
    });
  });

  it('should export logs as JSON', async () => {
    const user = userEvent.setup();
    const mockData = JSON.stringify([{ id: '1', content: 'test' }]);

    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: mockData,
        error: undefined,
        response: {} as Response,
      })
    );

    render(<ExportMenu filters={{ employee_id: 'emp-1' }} />);

    const exportButton = screen.getByRole('button', { name: /export/i });
    await user.click(exportButton);

    await waitFor(() => {
      expect(mockApiClientGET).toHaveBeenCalledWith('/logs/export', {
        params: {
          query: {
            format: 'json',
            employee_id: 'emp-1',
          },
        },
      });
    });
  });

  it('should show loading state during export', async () => {
    const user = userEvent.setup();

    mockApiClientGET.mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(() => resolve({ data: '[]', error: undefined, response: {} as Response }), 100)
        )
    );

    render(<ExportMenu filters={{}} />);

    const exportButton = screen.getByRole('button', { name: /export/i });
    await user.click(exportButton);

    expect(screen.getByText('Exporting...')).toBeInTheDocument();
    expect(exportButton).toBeDisabled();

    await waitFor(
      () => {
        expect(screen.queryByText('Exporting...')).not.toBeInTheDocument();
      },
      { timeout: 200 }
    );
  });

  it('should handle export errors', async () => {
    const user = userEvent.setup();
    const alertSpy = spyOn(window, 'alert').mockImplementation(() => {});

    interface ErrorType {
      message: string;
    }

    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: undefined,
        error: { message: 'Export failed' } as ErrorType,
        response: {} as Response,
      })
    );

    render(<ExportMenu filters={{}} />);

    const exportButton = screen.getByRole('button', { name: /export/i });
    await user.click(exportButton);

    await waitFor(() => {
      expect(alertSpy).toHaveBeenCalledWith('Failed to export logs. Please try again.');
    });

    alertSpy.mockRestore();
  });
});
