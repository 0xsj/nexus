# ============================================================================
# Builder stage
# ============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy go.mod and go.sum first for layer caching
COPY platform/go.mod platform/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY platform/ .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/api \
    ./cmd/api

# ============================================================================
# Runtime stage
# ============================================================================
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/api .

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Non-root user (distroless includes nonroot user)
USER nonroot:nonroot

# Expose port
EXPOSE 8090

# Environment variables (can be overridden at runtime)
ENV PORT=8090
ENV ENV=production
ENV LOG_LEVEL=info

ENTRYPOINT ["/app/api"]