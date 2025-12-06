import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

export async function GET(request: NextRequest) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const searchParams = request.nextUrl.searchParams;
  const query: Record<string, any> = {};
  
  // Forward common query params
  ['page', 'limit', 'search', 'team_id', 'role_id', 'status'].forEach(key => {
    const value = searchParams.get(key);
    if (value) query[key] = value;
  });

  try {
    const { data, error } = await apiClient.GET('/employees', {
      params: { query },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: (error as any).message || 'Failed to fetch employees' },
        { status: 500 }
      );
    }

    return NextResponse.json(data);
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Unknown error' },
      { status: 500 }
    );
  }
}
