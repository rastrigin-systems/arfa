# Development Dockerfile for Web UI
FROM node:20-alpine AS base

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

# Development target
FROM base AS development

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code (will be overridden by volume mount in dev)
COPY . .

# Expose port
EXPOSE 3000

# Start development server
CMD ["pnpm", "dev"]
