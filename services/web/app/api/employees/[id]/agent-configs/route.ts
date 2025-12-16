import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';

type RouteParams = { params: Promise<{ id: string }> };

export async function GET(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id } = await params;

  try {
    const { data, error, response } = await apiClient.GET('/employees/{employee_id}/agent-configs', {
      params: { path: { employee_id: id } },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to fetch employee agent configs') },
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

export async function POST(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id } = await params;

  try {
    const body = await request.json();

    const { data, error, response } = await apiClient.POST('/employees/{employee_id}/agent-configs', {
      params: { path: { employee_id: id } },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to create employee agent config') },
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
