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
    return [
      {
        source: '/api/:path*',
        destination: `${apiBaseUrl}/api/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
