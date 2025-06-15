# Multi-stage build for Go application
# Stage 1: Build the Go binary

# Use Go compiler image with Alpine for building
# This includes the full Go toolchain needed for compilation
FROM golang:1.24.4-alpine3.22 AS builder

# Set working directory inside the build container
WORKDIR /app

# Copy all source code and dependencies to the build container
COPY . .    

# Compile the Go application into a single binary
# -o main: output binary name as 'main'
# main.go: the entry point of our Go application
RUN go build -o main main.go

# Stage 2: Create minimal runtime image
# This stage creates the final production image

# Use minimal Alpine Linux (only ~5MB base image)
# No Go compiler or build tools - just what's needed to run the binary
FROM alpine:3.22

# Set working directory in the runtime container
WORKDIR /app

# Copy ONLY the compiled binary from the builder stage
# --from=builder: copy from the previous build stage
# This excludes source code, Go toolchain, and build cache
COPY --from=builder /app/main .
#todo: remove before production
COPY app.env .

# Expose port 8080 for the web server
# This is for documentation - actual port binding happens at runtime
EXPOSE 8080

# Start the application when container runs
# Execute the compiled binary
CMD ["/app/main"]