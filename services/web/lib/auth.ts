import { cookies } from 'next/headers';
import { apiClient } from './api/client';

const TOKEN_COOKIE_NAME = 'ubik_token';

/**
 * Server-side: Get token from cookies
 */
export async function getServerToken(): Promise<string | null> {
  const cookieStore = await cookies();
  return cookieStore.get(TOKEN_COOKIE_NAME)?.value || null;
}

/**
 * Server-side: Set token in cookies
 */
export async function setServerToken(token: string) {
  const cookieStore = await cookies();
  cookieStore.set(TOKEN_COOKIE_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: '/',
  });
}

/**
 * Server-side: Clear token from cookies
 */
export async function clearServerToken() {
  const cookieStore = await cookies();
  cookieStore.delete(TOKEN_COOKIE_NAME);
}

/**
 * Server-side: Verify if user is authenticated
 */
export async function isAuthenticated(): Promise<boolean> {
  const token = await getServerToken();
  if (!token) return false;

  // Verify token with /auth/me endpoint
  try {
    const { response } = await apiClient.GET('/auth/me', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.ok;
  } catch {
    return false;
  }
}

/**
 * Server-side: Get current authenticated employee
 */
export async function getCurrentEmployee() {
  const token = await getServerToken();
  if (!token) return null;

  try {
    const { data, error } = await apiClient.GET('/auth/me', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) return null;
    return data;
  } catch {
    return null;
  }
}
