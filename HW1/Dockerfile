# Start with the official Go image.
# The version should match the one in your go.mod file.
FROM golang:1.25.0

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files.
# This is done separately to take advantage of Docker's layer caching.
# The dependencies will only be re-downloaded if go.mod or go.sum changes.
COPY go.mod go.sum ./

# Download the Go module dependencies.
RUN go mod download

# Copy the rest of your application's source code.
COPY *.go ./

# Build the Go application.
# The -o flag specifies the output file name.
RUN CGO_ENABLED=0 GOOS=linux go build -o /web-service-gin

# Expose port 8080 to the outside world.
EXPOSE 8080

# Command to run the executable.
CMD ["/web-service-gin"]