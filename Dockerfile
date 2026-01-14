# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o bots ./cmd

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bots .

# Run the bot
ENTRYPOINT ["./bots"]
CMD ["-config=/app/config/config.yaml"]
