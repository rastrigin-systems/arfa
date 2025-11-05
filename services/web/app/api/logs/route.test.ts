import { describe, it, expect, vi, beforeEach } from 'vitest';
import { NextRequest } from 'next/server';
import { GET } from './route';
import * as auth from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

// Mock dependencies
vi.mock('@/lib/auth');
vi.mock('@/lib/api/client', () => ({
  apiClient: {
    GET: vi.fn(),
  },
}));

describe('GET /api/logs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns 401 when no token is present', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue(null);
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
    vi.mocked(auth.getServerToken).mockResolvedValue('test-token');
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { logs: [] },
      error: undefined,
      response: new Response(),
    });

    const url = new URL('http://localhost:3000/api/logs');
    url.searchParams.set('session_id', 'session-123');
    url.searchParams.set('employee_id', 'emp-123');
    url.searchParams.set('agent_id', 'agent-123');
    url.searchParams.set('event_type', 'input');
    url.searchParams.set('event_category', 'io');
    url.searchParams.set('start_date', '2024-01-01');
    url.searchParams.set('end_date', '2024-12-31');
    url.searchParams.set('limit', '50');
    url.searchParams.set('offset', '10');

    const request = new NextRequest(url);

    // Act
    await GET(request);

    // Assert
    expect(apiClient.GET).toHaveBeenCalledWith('/logs', {
      params: {
        query: {
          session_id: 'session-123',
          employee_id: 'emp-123',
          agent_id: 'agent-123',
          event_type: 'input',
          event_category: 'io',
          start_date: '2024-01-01',
          end_date: '2024-12-31',
          limit: 50,
          offset: 10,
        },
      },
      headers: {
        Authorization: 'Bearer test-token',
      },
    });
  });

  it('adds Authorization header with token', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue('my-secret-token');
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { logs: [] },
      error: undefined,
      response: new Response(),
    });

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    await GET(request);

    // Assert
    expect(apiClient.GET).toHaveBeenCalledWith(
      '/logs',
      expect.objectContaining({
        headers: {
          Authorization: 'Bearer my-secret-token',
        },
      })
    );
  });

  it('returns backend response data', async () => {
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

    vi.mocked(auth.getServerToken).mockResolvedValue('test-token');
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { logs: mockLogs },
      error: undefined,
      response: new Response(),
    });

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    const response = await GET(request);
    const data = await response.json();

    // Assert
    expect(response.status).toBe(200);
    expect(data).toEqual({ logs: mockLogs });
  });

  it('handles backend API errors', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue('test-token');
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: undefined,
      error: { message: 'Internal server error' },
      response: new Response(),
    });

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    const response = await GET(request);
    const data = await response.json();

    // Assert
    expect(response.status).toBe(500);
    expect(data).toEqual({ error: 'Internal server error' });
  });

  it('handles undefined query parameters', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue('test-token');
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { logs: [] },
      error: undefined,
      response: new Response(),
    });

    const request = new NextRequest('http://localhost:3000/api/logs');

    // Act
    await GET(request);

    // Assert
    expect(apiClient.GET).toHaveBeenCalledWith('/logs', {
      params: {
        query: {
          session_id: undefined,
          employee_id: undefined,
          agent_id: undefined,
          event_type: undefined,
          event_category: undefined,
          start_date: undefined,
          end_date: undefined,
          limit: 100,
          offset: 0,
        },
      },
      headers: {
        Authorization: 'Bearer test-token',
      },
    });
  });

  it('converts string limit and offset to numbers', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue('test-token');
    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { logs: [] },
      error: undefined,
      response: new Response(),
    });

    const url = new URL('http://localhost:3000/api/logs');
    url.searchParams.set('limit', '25');
    url.searchParams.set('offset', '5');
    const request = new NextRequest(url);

    // Act
    await GET(request);

    // Assert
    expect(apiClient.GET).toHaveBeenCalledWith('/logs', {
      params: {
        query: expect.objectContaining({
          limit: 25,
          offset: 5,
        }),
      },
      headers: expect.any(Object),
    });
  });
});
