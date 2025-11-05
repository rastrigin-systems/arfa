import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GET } from './route';
import * as auth from '@/lib/auth';

// Mock dependencies
vi.mock('@/lib/auth');

describe('GET /api/logs/ws-token', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns 401 when no token is present', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue(null);

    // Act
    const response = await GET();
    const data = await response.json();

    // Assert
    expect(response.status).toBe(401);
    expect(data).toEqual({ error: 'Unauthorized' });
  });

  it('returns token when authenticated', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue('my-secret-token');

    // Act
    const response = await GET();
    const data = await response.json();

    // Assert
    expect(response.status).toBe(200);
    expect(data).toEqual({ token: 'my-secret-token' });
  });

  it('does not expose token in plain text unnecessarily', async () => {
    // Arrange
    vi.mocked(auth.getServerToken).mockResolvedValue('sensitive-token-123');

    // Act
    const response = await GET();
    const data = await response.json();

    // Assert - token is returned for WebSocket connection
    expect(data.token).toBe('sensitive-token-123');
    expect(response.headers.get('Cache-Control')).toBe('no-store');
  });
});
