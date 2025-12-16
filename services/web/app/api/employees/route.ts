import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { getErrorMessage } from '@/lib/api/errors';
import type { components, paths } from '@/lib/api/schema';

type EmployeeStatus = components["parameters"]["Status"];

function isValidEmployeeStatus(status: string): status is EmployeeStatus {
  return ['active', 'suspended', 'inactive'].includes(status);
}

export async function GET(request: NextRequest) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const searchParams = request.nextUrl.searchParams;
  const query: paths["/employees"]["get"]["parameters"]["query"] = {};
  
  const page = searchParams.get('page');
  if (page !== null) {
    const numPage = parseInt(page, 10);
    if (!isNaN(numPage)) {
      query.page = numPage;
    }
  }

  const limit = searchParams.get('limit'); // Frontend sends 'limit', backend expects 'per_page'
  if (limit !== null) {
    const numLimit = parseInt(limit, 10);
    if (!isNaN(numLimit)) {
      query.per_page = numLimit;
    }
  }

  const status = searchParams.get('status');
  if (status !== null && isValidEmployeeStatus(status)) {
    query.status = status;
  }

  try {
    const { data, error }: { data?: components["schemas"]["ListEmployeesResponse"]; error?: components["schemas"]["Error"] } = await apiClient.GET('/employees', {
      params: { query },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: error.message || 'Failed to fetch employees' },
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

export async function POST(request: NextRequest) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const body = await request.json();

    const { data, error } = await apiClient.POST('/employees', {
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (error) {
      return NextResponse.json(
        { error: getErrorMessage(error, 'Failed to create employee') },
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
