import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import type { paths } from '@/lib/api/schema';

// Use the schema types for query parameters
type LogsQueryParams = paths['/logs']['get']['parameters']['query'];
type EventType = NonNullable<LogsQueryParams>['event_type'];
type EventCategory = NonNullable<LogsQueryParams>['event_category'];

export async function GET(request: NextRequest) {
  // Get token from httpOnly cookie
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  // Extract query parameters
  const searchParams = request.nextUrl.searchParams;
  const session_id = searchParams.get('session_id') || undefined;
  const employee_id = searchParams.get('employee_id') || undefined;
  const agent_id = searchParams.get('agent_id') || undefined;
  const event_type = (searchParams.get('event_type') as EventType) || undefined;
  const event_category = (searchParams.get('event_category') as EventCategory) || undefined;
  const start_date = searchParams.get('start_date') || undefined;
  const end_date = searchParams.get('end_date') || undefined;
  const page = parseInt(searchParams.get('page') || '1', 10);
  const per_page = parseInt(searchParams.get('per_page') || '20', 10);

  try {
    // Call backend API with Authorization header
    const { data, error } = await apiClient.GET('/logs', {
      params: {
        query: {
          session_id,
          employee_id,
          agent_id,
          event_type,
          event_category,
          start_date,
          end_date,
          page,
          per_page,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: error.message || 'Failed to fetch logs' },
        { status: 500 }
      );
    }

    return NextResponse.json({
      logs: data?.logs || [],
      pagination: data?.pagination || {
        total: 0,
        page: 1,
        per_page: 20,
        total_pages: 0,
      },
    });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Unknown error' },
      { status: 500 }
    );
  }
}
