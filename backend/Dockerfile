# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates and wget for HTTPS and health checks
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy the binary from builder
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Render provides PORT at runtime. The application uses SERVER_PORT, so map
# it here while keeping a local default for Docker Compose.
EXPOSE 10000

# Render performs the external health check. Avoid a fixed internal port here
# because PORT is assigned at runtime.
CMD ["sh", "-c", "SERVER_PORT=${PORT:-8080} exec ./main"]
