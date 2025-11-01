'use server';

import { redirect } from 'next/navigation';
import { clearServerToken, getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

export async function logoutAction() {
  const token = await getServerToken();

  if (token) {
    try {
      // Call logout API to invalidate session
      await apiClient.POST('/auth/logout', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
    } catch (error) {
      console.error('Logout API error:', error);
      // Continue with logout even if API call fails
    }
  }

  // Clear token from cookies
  await clearServerToken();

  // Redirect to login
  redirect('/login');
}
