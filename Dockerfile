# Dockerfile
# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags="-s -w" .

# Stage 2: Create the final, smaller image
FROM alpine:latest

# Install timezone data and ca-certificates for HTTPS requests
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 3000

# Run the application
CMD ["./main"]