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
    // API_BASE_URL should be the base URL without /api/v1 suffix
    // e.g., http://localhost:8080 or https://ubik-api-xxx.run.app
    const apiBaseUrl = process.env.API_BASE_URL || 'http://localhost:8080';
    return [
      {
        source: '/api/:path*',
        destination: `${apiBaseUrl}/api/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
