# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with static linking for smaller size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o go-dev-mcp ./cmd/server

# Stage 2: Create the minimal runtime image
FROM alpine:3.18

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/go-dev-mcp /app/go-dev-mcp

# Create default config directory
RUN mkdir -p /etc/go-dev-mcp

# Set environment variables
ENV PATH="/app:${PATH}"

# Expose any necessary ports
# Note: MCP servers typically communicate via stdin/stdout, so no ports needed unless for monitoring

# Create non-root user for better security
RUN adduser -D -u 1000 mcp
USER mcp

# Set the entrypoint
ENTRYPOINT ["/app/go-dev-mcp"]