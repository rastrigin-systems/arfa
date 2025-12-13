# Multi-stage build for Next.js Web UI (GCP Cloud Run) - Optimized for minimal size
FROM node:20-alpine AS deps
WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@9 --activate

# Install dependencies only when needed
COPY services/web/package.json services/web/pnpm-lock.yaml ./
RUN pnpm install --prod --frozen-lockfile

# Builder stage
FROM node:20-alpine AS builder
WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@9 --activate

# Copy dependencies
COPY services/web/package.json services/web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY services/web/ .

# Copy OpenAPI spec for type generation
RUN mkdir -p ../../platform/api-spec
COPY platform/api-spec/spec.yaml ../../platform/api-spec/spec.yaml

# Generate API types
RUN pnpm run generate:api

# Build Next.js app with standalone output
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
# NEXT_PUBLIC_* vars must be set at build time to be included in client bundle
ENV NEXT_PUBLIC_API_URL=https://ubik-api-754414213269.us-central1.run.app/api/v1
RUN pnpm run build

# Production stage - Ultra-minimal with standalone output
FROM node:20-alpine AS production
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV HOSTNAME="0.0.0.0"

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy standalone server and dependencies
# Standalone output creates a self-contained server in .next/standalone
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./

# Static files must be copied separately into their expected location
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Public files
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

USER nextjs

EXPOSE 8080

# Start Next.js standalone server
# The standalone server is at the root and expects to be run as "node server.js"
CMD ["node", "server.js"]
