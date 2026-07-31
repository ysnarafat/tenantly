# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached if go.mod/go.sum don't change)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .
# Ensure migrations folder exists
COPY --from=builder /app/migrations ./migrations

RUN chown -R app:app /app

# Configuration (DATABASE_URL, JWT_SECRET, CORS_ALLOWED_ORIGINS, etc.) is
# injected at container runtime via environment variables / a secrets store —
# never baked into the image.

USER app

EXPOSE 8080

CMD ["./main"]
