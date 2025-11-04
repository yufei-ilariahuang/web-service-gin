# Build stage
FROM --platform=linux/amd64 golang:1.25-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go mod files first for better cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application for linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o server

# Final stage
FROM --platform=linux/amd64 alpine:3.18

# Install curl for health checks and CA certificates for HTTPS
RUN apk add --no-cache curl ca-certificates

# Copy the binary from builder
COPY --from=builder /app/server /app/server

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/server"]

# Expose the port that the application listens on.
EXPOSE 8080

# Set the command to run when the container starts.
ENTRYPOINT ["/app/server"]