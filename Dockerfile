# ==========================================
# Multi-stage Dockerfile for Go + MongoDB API
# Produces a tiny, secure, production image (~15MB)
# ==========================================

# Stage 1: Build the binary
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy application source code
COPY . .

# Build statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /app/server \
    cmd/api/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Add runtime dependencies (SSL certs for Atlas connection & timezone data)
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' appuser

# Copy binary from builder
COPY --from=builder /app/server /app/server

USER appuser

EXPOSE 8080

CMD ["/app/server"]
