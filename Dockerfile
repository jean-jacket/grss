# Build stage
FROM golang:1.25.3-alpine3.22 AS builder

WORKDIR /app

# Copy go mod files for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build optimized binary with stripped symbols
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o grss \
    cmd/grss/main.go

# Runtime stage - use scratch for minimal size
FROM scratch

# Copy CA certificates from builder for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /app/grss /grss

# Expose port
EXPOSE 1200

# Set default environment variables
ENV PORT=1200 \
    CACHE_TYPE=memory \
    MEMORY_MAX=256

# Run as non-privileged user (scratch doesn't have users, but the UID still applies)
USER 65534:65534

# Run
ENTRYPOINT ["/grss"]
