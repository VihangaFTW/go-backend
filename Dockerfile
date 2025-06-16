
#? Stage 1: Build the Go binary and tools
# Use Go compiler image with Alpine for building
FROM golang:1.24.4-alpine3.22 AS builder

# Set working directory inside the build container
WORKDIR /app

# Copy all source code and dependencies to the build container
COPY . .    

# Compile the Go application into a single binary
RUN go build -o main main.go

# Install golang-migrate CLI for database migrations
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

#? Stage 2: Create minimal runtime image
# Use minimal Alpine Linux base image for production
FROM alpine:3.22

# Set working directory in the runtime container
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Copy configuration file (TODO: remove before production)
COPY app.env .

# Copy the migrate executable from the builder stage
COPY --from=builder /go/bin/migrate ./migrate

# Copy database migration files
COPY db/migration ./migration/

# Copy startup script
COPY start.sh .

# Make startup script executable
RUN chmod +x start.sh

# Expose port 8080 for the web server
EXPOSE 8080

# Set the startup script as entrypoint to run migrations before starting the app
ENTRYPOINT [ "/app/start.sh" ]

# Default command to run the application
CMD ["/app/main"]