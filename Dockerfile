# --- Stage 1: Build the Go application ---
# On an M1 Mac, this 'FROM' automatically pulls the ARM64 image.
# We will use it to *natively* cross-compile to AMD64.
FROM golang:1.24-alpine AS builder

WORKDIR /app

# (Optional) Install git if you use private modules
# RUN apk add --no-cache git

# Set cross-compilation targets for Cloud Run (Linux AMD64)
# This is *required* for building on M1 for Cloud Run.
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Copy and download dependencies first to leverage build cache
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the static binary, stripping debug info (-w -s) to reduce size.
# The output path is /app/main
RUN go build -ldflags="-w -s" -o /app/main ./cmd/app/main.go

# --- Stage 2: Create the final minimal runtime container ---
# We MUST specify --platform=linux/amd64 for the final stage
# to ensure Docker pulls the AMD64 base image, not the ARM64 one.
# We use 'static-debian12' because our binary is fully static (CGO_ENABLED=0).
FROM --platform=linux/amd64 gcr.io/distroless/static-debian12

# Use a non-root user (this is the default in distroless, but good to be explicit)
USER nonroot:nonroot

WORKDIR /app

# Copy *only* the compiled binary from the builder stage
COPY --from=builder /app/main .

# (Optional) Informative port. Your app must read $PORT from the env.
EXPOSE 8080

# Set the entrypoint to run the application
ENTRYPOINT ["/app/main"]
