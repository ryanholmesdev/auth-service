# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application with production flags
RUN go build -ldflags="-s -w" -o auth-service

# Stage 2: Create a minimal runtime image
FROM alpine:3.18

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Expose the application port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./auth-service"]