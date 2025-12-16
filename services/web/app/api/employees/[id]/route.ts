import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';

type RouteParams = { params: Promise<{ id: string }> };

export async function PATCH(request: NextRequest, { params }: RouteParams) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id } = await params;

  try {
    const body = await request.json();

    const { data, error } = await apiClient.PATCH('/employees/{employee_id}', {
      params: { path: { employee_id: id } },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to update employee') },
        { status: 500 }
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

  const { id } = await params;

  try {
    const { error } = await apiClient.DELETE('/employees/{employee_id}', {
      params: { path: { employee_id: id } },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to delete employee') },
        { status: 500 }
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
