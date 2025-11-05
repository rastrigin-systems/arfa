import { NextResponse } from 'next/server';
import { getServerToken } from '@/lib/auth';

export async function GET() {
  // Get token from httpOnly cookie
  const token = await getServerToken();

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  // Return token for WebSocket connection
  return NextResponse.json(
    { token },
    {
      status: 200,
      headers: {
        'Cache-Control': 'no-store',
      },
    }
  );
}
