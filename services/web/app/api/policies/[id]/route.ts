import { NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getServerToken();
  const { id } = await params;

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const { data, error } = await apiClient.GET('/policies/{policy_id}', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: { path: { policy_id: id } },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to fetch policy') },
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

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getServerToken();
  const { id } = await params;

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const body = await request.json();

    const { data, error } = await apiClient.PATCH('/policies/{policy_id}', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: { path: { policy_id: id } },
      body,
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to update policy') },
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

export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getServerToken();
  const { id } = await params;

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const { error } = await apiClient.DELETE('/policies/{policy_id}', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: { path: { policy_id: id } },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to delete policy') },
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
