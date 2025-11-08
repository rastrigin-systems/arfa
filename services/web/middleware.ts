import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const TOKEN_COOKIE_NAME = 'ubik_token';

// Public routes that don't require authentication
// These routes will handle their own auth logic (e.g., redirect after login)
const publicRoutes = [
  '/',             // Root page - redirects based on auth status
  '/login',        // Login page
  '/signup',       // Signup page
  '/accept-invite' // Accept invitation page
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Check if route is public
  const isPublicRoute = publicRoutes.some((route) => pathname.startsWith(route));

  // Get token from cookies
  const token = request.cookies.get(TOKEN_COOKIE_NAME)?.value;

  // If user is not authenticated and trying to access protected route
  if (!token && !isPublicRoute) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('from', pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Let pages handle their own post-auth redirects
  // This prevents redirect loops when tokens are invalid
  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.png$).*)',
  ],
};
