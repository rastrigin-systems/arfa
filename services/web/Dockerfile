# Multi-stage build for Next.js Web UI (GCP Cloud Run) - Optimized for minimal size
FROM node:20-alpine AS deps
WORKDIR /app

# Install dependencies only when needed
COPY services/web/package*.json ./
RUN npm ci --only=production && \
    npm cache clean --force

# Builder stage
FROM node:20-alpine AS builder
WORKDIR /app

# Copy dependencies
COPY services/web/package*.json ./
RUN npm ci

COPY services/web/ .

# Copy OpenAPI spec for type generation
RUN mkdir -p ../../shared/openapi
COPY shared/openapi/spec.yaml ../../shared/openapi/spec.yaml

# Generate API types
RUN npm run generate:api

# Build Next.js app with standalone output
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN npm run build

# Production stage - Ultra-minimal with standalone output
FROM node:20-alpine AS production
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy only the standalone output (includes server + minimal dependencies)
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Start Next.js standalone server (no npm needed!)
CMD ["node", "server.js"]
