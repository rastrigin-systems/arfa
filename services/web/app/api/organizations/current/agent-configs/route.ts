import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';

export async function GET() {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const { data, error, response } = await apiClient.GET('/organizations/current/agent-configs', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to fetch org agent configs') },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (err) {
    return NextResponse.json(
      { error: getErrorMessage(err, 'Unknown error') },
      { status: 500 }
    );
  }
}

export async function POST(request: NextRequest) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const body = await request.json();

    const { data, error, response } = await apiClient.POST('/organizations/current/agent-configs', {
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to create org agent config') },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (err) {
    return NextResponse.json(
      { error: getErrorMessage(err, 'Unknown error') },
      { status: 500 }
    );
  }
}
