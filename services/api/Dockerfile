# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /workspace

# Copy workspace files
COPY go.work go.work.sum ./

# Copy all module files for dependency download
COPY services/api/go.mod services/api/go.sum ./services/api/
COPY services/cli/go.mod services/cli/go.sum ./services/cli/
COPY pkg/types/go.mod ./pkg/types/
COPY generated/go.mod generated/go.sum ./generated/

# Download dependencies (workspace-aware)
RUN go work sync
RUN cd services/api && go mod download

# Copy source code
COPY services/api/ ./services/api/
COPY pkg/ ./pkg/
COPY generated/ ./generated/

# Build the binary
WORKDIR /workspace/services/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/ubik-server ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /bin/ubik-server /bin/ubik-server

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the server
CMD ["/bin/ubik-server"]
