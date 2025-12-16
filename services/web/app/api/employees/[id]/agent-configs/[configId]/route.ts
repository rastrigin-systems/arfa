import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';

type RouteParams = { params: Promise<{ id: string; configId: string }> };

export async function PATCH(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id, configId } = await params;

  try {
    const body = await request.json();

    const { data, error, response } = await apiClient.PATCH('/employees/{employee_id}/agent-configs/{config_id}', {
      params: { path: { employee_id: id, config_id: configId } },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to update employee agent config') },
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

export async function DELETE(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id, configId } = await params;

  try {
    const { error, response } = await apiClient.DELETE('/employees/{employee_id}/agent-configs/{config_id}', {
      params: { path: { employee_id: id, config_id: configId } },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to delete employee agent config') },
        { status: response.status }
      );
    }

    return NextResponse.json({ success: true });
  } catch (err) {
    return NextResponse.json(
      { error: getErrorMessage(err, 'Unknown error') },
      { status: 500 }
    );
  }
}
