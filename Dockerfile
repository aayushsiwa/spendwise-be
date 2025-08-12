# syntax=docker/dockerfile:1

FROM golang:1.24.3-alpine

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod tidy

# Copy source files
COPY . .

# Build the application

# Use a minimal image for runtime
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=0 /app/main .
COPY records.db .
COPY .env .

# Expose port if needed
EXPOSE 8080

CMD ["./main"]
