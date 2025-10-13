# --- Stage 1: Build ---
# Use a specific version of the official Go image as the builder.
# Using alpine for a smaller builder image.
FROM --platform=linux/amd64 golang:1.25.0 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the dependency files first to leverage Docker's build cache.
# This step will only be re-run if go.mod or go.sum changes.
COPY go.mod go.sum ./

# Download all the dependencies
RUN go mod download

# Copy the rest of the source code into the container
COPY . .

# Build the Go application, creating a statically linked binary.
# CGO_ENABLED=0 is important for creating a pure Go binary without C dependencies.
# -o specifies the output file name.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/server .


# --- Stage 2: Final Image ---
# Use a minimal, non-root base image for the final container.
# alpine is very small and secure.
FROM --platform=linux/amd64 alpine:latest

# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/server /app/server

# Expose the port that the application listens on.
EXPOSE 8080

# Set the command to run when the container starts.
ENTRYPOINT ["/app/server"]