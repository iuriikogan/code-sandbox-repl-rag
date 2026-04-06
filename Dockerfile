# Stage 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary statically so it runs seamlessly on Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o code-sandbox ./cmd/sandbox

# Stage 2: Create the runtime environment
# We use a Python base image because the Go orchestrator uses os/exec to run Python scripts
FROM python:3.12-alpine

WORKDIR /app

# Copy the statically compiled Go binary from the builder stage
COPY --from=builder /app/code-sandbox .

# Set default entrypoint
ENTRYPOINT ["./code-sandbox"]
