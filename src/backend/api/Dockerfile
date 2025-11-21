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

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .
# Ensure migrations folder exists
COPY --from=builder /app/migrations ./migrations

# Create .env file with default values
RUN echo "DATABASE_URL=postgres://postgres:password@postgres:5432/tenantly?sslmode=disable" > .env && \
    echo "JWT_SECRET=your-secret-key-change-in-production" >> .env && \
    echo "JWT_EXPIRATION=8h" >> .env && \
    echo "ENVIRONMENT=production" >> .env

EXPOSE 8080

CMD ["./main"]
