# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

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
