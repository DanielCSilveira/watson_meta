# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set timezone (optional, adjust as needed)
ENV TZ=America/Sao_Paulo

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app .

# Copy .env file (optional, better to use env vars in production)
# COPY .env .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./app"]
