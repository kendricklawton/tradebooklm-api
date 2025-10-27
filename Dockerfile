# Use the official Go image as a base for building the application
FROM golang:1.24-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/app/main.go

# Use a minimal base image for the final container
FROM gcr.io/distroless/base-debian11

# Set the working directory inside the container
WORKDIR /

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]