# Use the official Golang image to create a build artifact.
FROM golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o bot-builder ./cmd/bot-builder

# Start a new stage from scratch
FROM debian:bookworm-slim

# Set the Current Working Directory inside the container
WORKDIR /app/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bot-builder .

# Expose port 5005 to the outside world
EXPOSE 5005

# Command to run the executable
CMD ["./bot-builder"]