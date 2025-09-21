# syntax=docker/dockerfile:1

FROM golang:1.24.3 as builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source files
COPY . .

COPY .env .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal image for runtime
FROM golang:1.24.3

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .
COPY records.db .

# Expose port if needed
EXPOSE 8080

ENTRYPOINT ["./main"]
