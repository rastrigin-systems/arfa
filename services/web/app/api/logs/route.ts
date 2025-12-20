import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

type EventType = 'input' | 'output' | 'error' | 'session_start' | 'session_end' | 'agent.installed' | 'mcp.configured' | 'config.synced' | 'user_prompt' | 'ai_text' | 'tool_call' | 'tool_result' | 'api_request' | 'api_response';
type EventCategory = 'io' | 'agent' | 'mcp' | 'auth' | 'admin' | 'classified' | 'proxy';

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
  const per_page = parseInt(searchParams.get('per_page') || '100', 10);

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

    return NextResponse.json({ logs: data?.logs || [] });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Unknown error' },
      { status: 500 }
    );
  }
}
