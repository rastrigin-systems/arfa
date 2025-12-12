/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Enable standalone output for optimal Docker images
  output: 'standalone',
  // Disable instrumentation in production to prevent memory issues
  // Only enable locally for E2E tests if needed
  // experimental: {
  //   instrumentationHook: true,
  // },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  },
  async rewrites() {
    // Derive API base URL from API_URL or NEXT_PUBLIC_API_URL
    // These typically end with /api/v1, so strip that suffix to get the base
    const apiUrl = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
    const apiBaseUrl = process.env.API_BASE_URL || apiUrl.replace(/\/api\/v1$/, '');
    return {
      // afterFiles rewrites are checked AFTER Next.js API routes
      // This lets /api/employees, /api/teams, etc. be handled by our route handlers
      // Only /api/v1/* gets forwarded to the backend
      afterFiles: [
        {
          source: '/api/v1/:path*',
          destination: `${apiBaseUrl}/api/v1/:path*`,
        },
      ],
    };
  },
};

module.exports = nextConfig;
