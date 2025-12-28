# Development Dockerfile for Web UI
FROM oven/bun:1 AS base

WORKDIR /app

# Development target
FROM base AS development

# Copy package files
COPY package.json bun.lock* ./

# Install dependencies
RUN bun install --frozen-lockfile || bun install

# Copy source code (will be overridden by volume mount in dev)
COPY . .

# Expose port
EXPOSE 3000

# Start development server
CMD ["bun", "run", "dev"]
