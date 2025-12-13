import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';

type RouteParams = { params: Promise<{ id: string; configId: string }> };

export async function PATCH(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id, configId } = await params;

  try {
    const body = await request.json();

    const { data, error } = await apiClient.PATCH('/employees/{employee_id}/agent-configs/{config_id}', {
      params: { path: { employee_id: id, config_id: configId } },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        { error: (error as any).message || 'Failed to update employee agent config' },
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

export async function DELETE(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id, configId } = await params;

  try {
    const { error } = await apiClient.DELETE('/employees/{employee_id}/agent-configs/{config_id}', {
      params: { path: { employee_id: id, config_id: configId } },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        { error: (error as any).message || 'Failed to delete employee agent config' },
        { status: 500 }
      );
    }

    return NextResponse.json({ success: true });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Unknown error' },
      { status: 500 }
    );
  }
}
