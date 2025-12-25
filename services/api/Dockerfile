# Development Dockerfile for API Server
FROM golang:1.24-alpine AS base

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Development target
FROM base AS development

# Copy go workspace files
COPY go.work go.work.sum ./

# Copy platform files (schema, API spec)
COPY platform/ platform/

# Copy all service modules referenced in go.work
COPY services/api/ services/api/
COPY services/cli/ services/cli/

# Copy generated code (or create empty module for workspace)
COPY generated/ generated/

# Build the server
WORKDIR /app/services/api
RUN go build -v -o /app/bin/arfa-server cmd/server/main.go

# Expose port
EXPOSE 8080

# Start server
CMD ["/app/bin/arfa-server"]
