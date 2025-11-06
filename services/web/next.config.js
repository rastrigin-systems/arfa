/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Enable standalone output for optimal Docker images
  output: 'standalone',
  // Enable instrumentation for MSW during E2E tests
  experimental: {
    instrumentationHook: true,
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
