import { NextRequest, NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';

export async function GET(request: NextRequest) {
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const searchParams = request.nextUrl.searchParams;
  const backendUrl = process.env.API_URL || 'http://localhost:3001/api/v1';
  
  try {
    const url = new URL(`${backendUrl}/logs/export`);
    searchParams.forEach((value, key) => url.searchParams.append(key, value));

    const response = await fetch(url.toString(), {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Backend responded with ${response.status}`);
    }

    const blob = await response.blob();
    const headers = new Headers();
    headers.set('Content-Type', response.headers.get('Content-Type') || 'application/octet-stream');
    if (response.headers.get('Content-Disposition')) {
      headers.set('Content-Disposition', response.headers.get('Content-Disposition')!);
    }

    return new NextResponse(blob, { status: 200, headers });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Export failed' },
      { status: 500 }
    );
  }
}
