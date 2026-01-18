import { describe, it, expect, beforeEach, mock } from 'bun:test';
import { NextRequest } from 'next/server';

// Types for mock returns
interface LogsResponse {
  logs: unknown[];
  pagination?: {
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
  };
}

interface ApiResponse {
  data: LogsResponse | undefined;
  error: { message: string } | undefined;
  response: Response;
}

// Create mock functions with proper types
const mockGetServerToken = mock<() => Promise<string | null>>(() => Promise.resolve('test-token'));
const mockApiClientGET = mock<() => Promise<ApiResponse>>(() => Promise.resolve({ data: { logs: [] }, error: undefined, response: new Response() }));

// Mock dependencies
mock.module('@/lib/auth', () => ({
  getServerToken: mockGetServerToken,
}));

mock.module('@/lib/api/client', () => ({
  apiClient: {
    GET: mockApiClientGET,
  },
}));

// Import after mocking
import { GET } from './route';

describe('GET /api/logs', () => {
  beforeEach(() => {
    mockGetServerToken.mockClear();
    mockApiClientGET.mockClear();
  });

  it('returns 401 when no token is present', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve(null));
    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    const response = await GET(request);
    const data = await response.json();

    // Assert
    expect(response.status).toBe(401);
    expect(data).toEqual({ error: 'Unauthorized' });
  });

  it('forwards all query parameters to backend API', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('test-token'));
    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: { logs: [] },
        error: undefined,
        response: new Response(),
      })
    );

    const url = new URL('http://localhost:3000/api/logs');
    url.searchParams.set('session_id', 'session-123');
    url.searchParams.set('employee_id', 'emp-123');
    url.searchParams.set('agent_id', 'agent-123');
    url.searchParams.set('event_type', 'input');
    url.searchParams.set('event_category', 'io');
    url.searchParams.set('start_date', '2024-01-01');
    url.searchParams.set('end_date', '2024-12-31');
    url.searchParams.set('page', '2');
    url.searchParams.set('per_page', '50');

    const request = new NextRequest(url);

    // Act
    await GET(request);

    // Assert
    expect(mockApiClientGET).toHaveBeenCalledWith('/logs', {
      params: {
        query: {
          session_id: 'session-123',
          employee_id: 'emp-123',
          agent_id: 'agent-123',
          event_type: 'input',
          event_category: 'io',
          start_date: '2024-01-01',
          end_date: '2024-12-31',
          page: 2,
          per_page: 50,
        },
      },
      headers: {
        Authorization: 'Bearer test-token',
      },
    });
  });

  it('adds Authorization header with token', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('my-secret-token'));
    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: { logs: [] },
        error: undefined,
        response: new Response(),
      })
    );

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    await GET(request);

    // Assert
    expect(mockApiClientGET).toHaveBeenCalledWith(
      '/logs',
      expect.objectContaining({
        headers: {
          Authorization: 'Bearer my-secret-token',
        },
      })
    );
  });

  it('returns backend response data with pagination', async () => {
    // Arrange
    const mockLogs = [
      {
        id: 'log-1',
        session_id: 'session-123',
        employee_id: 'emp-123',
        event_type: 'input',
        event_category: 'io',
        timestamp: '2024-01-01T00:00:00Z',
        data: { message: 'test' },
      },
    ];
    const mockPagination = {
      total: 100,
      page: 1,
      per_page: 20,
      total_pages: 5,
    };

    mockGetServerToken.mockImplementation(() => Promise.resolve('test-token'));
    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: { logs: mockLogs, pagination: mockPagination },
        error: undefined,
        response: new Response(),
      })
    );

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    const response = await GET(request);
    const data = await response.json();

    // Assert
    expect(response.status).toBe(200);
    expect(data).toEqual({ logs: mockLogs, pagination: mockPagination });
  });

  it('handles backend API errors', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('test-token'));
    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: undefined,
        error: { message: 'Internal server error' },
        response: new Response(),
      })
    );

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    const response = await GET(request);
    const data = await response.json();

    // Assert
    expect(response.status).toBe(500);
    expect(data).toEqual({ error: 'Internal server error' });
  });

  it('handles undefined query parameters with default pagination', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('test-token'));
    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: { logs: [] },
        error: undefined,
        response: new Response(),
      })
    );

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    await GET(request);

    // Assert
    expect(mockApiClientGET).toHaveBeenCalledWith('/logs', {
      params: {
        query: {
          session_id: undefined,
          employee_id: undefined,
          agent_id: undefined,
          event_type: undefined,
          event_category: undefined,
          start_date: undefined,
          end_date: undefined,
          page: 1,
          per_page: 20,
        },
      },
      headers: {
        Authorization: 'Bearer test-token',
      },
    });
  });

  it('converts string page and per_page to numbers', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('test-token'));
    mockApiClientGET.mockImplementation(() =>
      Promise.resolve({
        data: { logs: [] },
        error: undefined,
        response: new Response(),
      })
    );

    const url = new URL('http://localhost:3000/api/logs');
    url.searchParams.set('page', '3');
    url.searchParams.set('per_page', '25');
    const request = new NextRequest(url);

    // Act
    await GET(request);

    // Assert
    expect(mockApiClientGET).toHaveBeenCalledWith('/logs', {
      params: {
        query: expect.objectContaining({
          page: 3,
          per_page: 25,
        }),
      },
      headers: expect.any(Object),
    });
  });
});
