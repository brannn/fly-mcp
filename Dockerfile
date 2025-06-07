# Multi-stage build optimized for Fly.io infrastructure
FROM golang:1.21-alpine AS builder

# Install security updates and required packages
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o fly-mcp-server ./cmd/fly-mcp

# Final stage - minimal runtime
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/fly-mcp-server /fly-mcp-server

# Copy configuration files
COPY --from=builder /app/config.production.yaml /config.yaml

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/fly-mcp-server", "version"]

# Run the application
ENTRYPOINT ["/fly-mcp-server"]
CMD ["--config", "/config.yaml"]
