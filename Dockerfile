FROM golang:1.20 as builder
WORKDIR /app
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

# Use a minimal image to run the compiled binary
FROM alpine:latest

# Install MySQL client for testing connectivity
RUN apk --no-cache add mysql-client

# Set the working directory
WORKDIR /root/

# Copy the binary from the build stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the binary
CMD ["./main"]
