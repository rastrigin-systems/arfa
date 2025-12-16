import { NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';

export async function GET() {
  // Get token from httpOnly cookie
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    // Call backend API with Authorization header
    const { data, error } = await apiClient.GET('/teams', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to fetch teams') },
        { status: 500 }
      );
    }

    // Return the teams array - data is typed as ListTeamsResponse
    return NextResponse.json({ teams: data?.teams ?? [] });
  } catch (err) {
    return NextResponse.json(
      { error: getErrorMessage(err, 'Unknown error') },
      { status: 500 }
    );
  }
}
