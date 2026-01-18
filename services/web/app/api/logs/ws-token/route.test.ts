import { describe, it, expect, beforeEach, mock } from 'bun:test';

// Create mock function with proper type
const mockGetServerToken = mock<() => Promise<string | null>>(() => Promise.resolve('test-token'));

// Mock dependencies
mock.module('@/lib/auth', () => ({
  getServerToken: mockGetServerToken,
}));

// Import after mocking
import { GET } from './route';

describe('GET /api/logs/ws-token', () => {
  beforeEach(() => {
    mockGetServerToken.mockClear();
  });

  it('returns 401 when no token is present', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve(null));

    // Act
    const response = await GET();
    const data = await response.json();

    // Assert
    expect(response.status).toBe(401);
    expect(data).toEqual({ error: 'Unauthorized' });
  });

  it('returns token when authenticated', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('my-secret-token'));

    // Act
    const response = await GET();
    const data = await response.json();

    // Assert
    expect(response.status).toBe(200);
    expect(data).toEqual({ token: 'my-secret-token' });
  });

  it('does not expose token in plain text unnecessarily', async () => {
    // Arrange
    mockGetServerToken.mockImplementation(() => Promise.resolve('sensitive-token-123'));

    // Act
    const response = await GET();
    const data = await response.json();

    // Assert - token is returned for WebSocket connection
    expect(data.token).toBe('sensitive-token-123');
    expect(response.headers.get('Cache-Control')).toBe('no-store');
  });
});
