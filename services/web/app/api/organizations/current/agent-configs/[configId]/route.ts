import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

type RouteParams = { params: Promise<{ configId: string }> };

export async function PATCH(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { configId } = await params;

  try {
    const body = await request.json();

    const { data, error, response } = await apiClient.PATCH('/organizations/current/agent-configs/{config_id}', {
      params: { path: { config_id: configId } },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const errorMessage = (error as any).error || 'Failed to update org agent config';
      return NextResponse.json({ error: errorMessage }, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Unknown error' },
      { status: 500 }
    );
  }
}

export async function DELETE(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { configId } = await params;

  try {
    const { error, response } = await apiClient.DELETE('/organizations/current/agent-configs/{config_id}', {
      params: { path: { config_id: configId } },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const errorMessage = (error as any).error || 'Failed to delete org agent config';
      return NextResponse.json({ error: errorMessage }, { status: response.status });
    }

    return NextResponse.json({ success: true });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Unknown error' },
      { status: 500 }
    );
  }
}
