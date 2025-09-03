# Stage 1: Build the Go binary
FROM golang:1.23.9-alpine AS builder

# Set environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the main file 
COPY cmd/ /app/auth

# Copy the internal libs 
COPY internal/ /app/internal 

# Run tests before building 
RUN go test ./... -v

# Build the Go binary for auth server 
RUN go build -o /bin/auth/server /app/auth

# Stage 2: Create a minimal runtime image for auth
FROM alpine:latest AS crypto_rates_auth_image

# Set working directory in runtime container
WORKDIR /app

COPY --from=builder /bin/auth . 
COPY config.yaml .
COPY web/ ./web

# Set the entrypoint
ENTRYPOINT ["./server"]
