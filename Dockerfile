# Build stage
FROM golang:1.23-alpine AS builder

# Install required packages
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o concert-ticket-api ./cmd/server/

# Final stage
FROM alpine:3.16

# Install necessary dependencies
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# Set working directory
WORKDIR /app

# Create directories for config and migrations
RUN mkdir -p /app/config /app/scripts/migrations

# Copy binary from builder stage
COPY --from=builder /app/concert-ticket-api /app/

# Copy config and migrations
COPY --from=builder /app/config/config.yaml /app/config/
COPY --from=builder /app/scripts/migrations /app/scripts/migrations/

# Copy and make the entrypoint script executable
COPY scripts/docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

# Expose ports
EXPOSE 8080 50051

# Set the entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]