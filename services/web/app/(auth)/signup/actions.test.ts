import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signupAction, checkSlugAvailability } from './actions';

// Mock API client
vi.mock('@/lib/api/client', () => ({
  apiClient: {
    POST: vi.fn(),
    GET: vi.fn(),
  },
}));

// Mock auth utilities
vi.mock('@/lib/auth', () => ({
  setServerToken: vi.fn(),
}));

// Mock Next.js navigation
vi.mock('next/navigation', () => ({
  redirect: vi.fn(() => {
    throw new Error('NEXT_REDIRECT');
  }),
}));

describe('signupAction', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should validate email format', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'invalid-email');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.email).toContain('Invalid email address');
  });

  it('should validate full name is at least 2 characters', async () => {
    const formData = new FormData();
    formData.append('full_name', 'A');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.full_name).toBeDefined();
  });

  it('should validate organization name is at least 2 characters', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'A');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.org_name).toBeDefined();
  });

  it('should validate organization slug format', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'Invalid Slug!');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.org_slug).toBeDefined();
  });

  it('should validate slug starts with letter', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', '123org');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.org_slug).toBeDefined();
  });

  it('should validate slug is 3-50 characters', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'ab');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.org_slug).toBeDefined();
  });

  it('should validate password is at least 8 characters', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Short1!');
    formData.append('confirm_password', 'Short1!');

    const result = await signupAction({}, formData);

    expect(result.errors?.password).toBeDefined();
  });

  it('should validate password contains uppercase letter', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'password123!');
    formData.append('confirm_password', 'password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.password).toBeDefined();
  });

  it('should validate password contains number', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password!');
    formData.append('confirm_password', 'Password!');

    const result = await signupAction({}, formData);

    expect(result.errors?.password).toBeDefined();
  });

  it('should validate password contains special character', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123');
    formData.append('confirm_password', 'Password123');

    const result = await signupAction({}, formData);

    expect(result.errors?.password).toBeDefined();
  });

  it('should validate passwords match', async () => {
    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'DifferentPass123!');

    const result = await signupAction({}, formData);

    expect(result.errors?.confirm_password).toContain('Passwords do not match');
  });

  it('should call register API with valid data', async () => {
    const { apiClient } = await import('@/lib/api/client');
    const { setServerToken } = await import('@/lib/auth');

    vi.mocked(apiClient.POST).mockResolvedValue({
      data: {
        token: 'jwt-token',
        employee: {
          id: 'emp-123',
          email: 'john@example.com',
          full_name: 'John Doe',
        },
        organization: {
          id: 'org-123',
          name: 'Acme Corp',
          slug: 'acme',
        },
      },
      error: undefined,
      response: { ok: true } as Response,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    try {
      await signupAction({}, formData);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      // Expect redirect error
      expect(error.message).toBe('NEXT_REDIRECT');
    }

    expect(apiClient.POST).toHaveBeenCalledWith('/auth/register', {
      body: {
        full_name: 'John Doe',
        email: 'john@example.com',
        org_name: 'Acme Corp',
        org_slug: 'acme',
        password: 'Password123!',
      },
    });

    expect(setServerToken).toHaveBeenCalledWith('jwt-token');
  });

  it('should handle duplicate email error', async () => {
    const { apiClient } = await import('@/lib/api/client');

    vi.mocked(apiClient.POST).mockResolvedValue({
      data: undefined,
      error: {
        message: 'Email already registered',
      },
      response: { ok: false, status: 409 } as Response,
    });

    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'existing@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?._form).toBeDefined();
  });

  it('should handle duplicate slug error', async () => {
    const { apiClient } = await import('@/lib/api/client');

    vi.mocked(apiClient.POST).mockResolvedValue({
      data: undefined,
      error: {
        message: 'Organization slug already taken',
      },
      response: { ok: false, status: 409 } as Response,
    });

    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'existing-slug');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?._form).toBeDefined();
  });

  it('should handle network errors', async () => {
    const { apiClient } = await import('@/lib/api/client');

    vi.mocked(apiClient.POST).mockRejectedValue(new Error('Network error'));

    const formData = new FormData();
    formData.append('full_name', 'John Doe');
    formData.append('email', 'john@example.com');
    formData.append('org_name', 'Acme Corp');
    formData.append('org_slug', 'acme');
    formData.append('password', 'Password123!');
    formData.append('confirm_password', 'Password123!');

    const result = await signupAction({}, formData);

    expect(result.errors?._form).toBeDefined();
    expect(result.errors?._form?.[0]).toContain('An unexpected error occurred');
  });
});

describe('checkSlugAvailability', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should return available when slug is not taken', async () => {
    const { apiClient } = await import('@/lib/api/client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { available: true },
      error: undefined,
      response: { ok: true } as Response,
    });

    const result = await checkSlugAvailability('available-slug');

    expect(result.available).toBe(true);
    expect(apiClient.GET).toHaveBeenCalledWith('/auth/check-slug', {
      params: {
        query: { slug: 'available-slug' },
      },
    });
  });

  it('should return unavailable when slug is taken', async () => {
    const { apiClient } = await import('@/lib/api/client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { available: false },
      error: undefined,
      response: { ok: true } as Response,
    });

    const result = await checkSlugAvailability('taken-slug');

    expect(result.available).toBe(false);
  });

  it('should handle API errors gracefully', async () => {
    const { apiClient } = await import('@/lib/api/client');

    vi.mocked(apiClient.GET).mockRejectedValue(new Error('Network error'));

    const result = await checkSlugAvailability('error-slug');

    // Should default to unavailable on error
    expect(result.available).toBe(false);
  });
});
