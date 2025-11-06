# Multi-stage build for Next.js Web UI (GCP Cloud Run)
FROM node:20-alpine AS deps
WORKDIR /app

# Copy package files from web directory
COPY services/web/package*.json ./
RUN npm ci

# Builder stage
FROM node:20-alpine AS builder
WORKDIR /app

# Copy dependencies
COPY --from=deps /app/node_modules ./node_modules
COPY services/web/ .

# Copy OpenAPI spec for type generation (match the path expected by generate:api script)
RUN mkdir -p ../../shared/openapi
COPY shared/openapi/spec.yaml ../../shared/openapi/spec.yaml

# Generate API types
RUN npm run generate:api

# Build Next.js app
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN npm run build

# Production stage
FROM node:20-alpine AS production
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy package files and install production dependencies only
COPY services/web/package*.json ./
RUN npm ci --omit=dev

# Copy built app from builder
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.js ./next.config.js

# Set ownership
RUN chown -R nextjs:nodejs /app

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Start production server
CMD ["npm", "start"]
