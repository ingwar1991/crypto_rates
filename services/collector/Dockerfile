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

# Copy the main files 
COPY cmd/worker /app/worker 
COPY cmd/ohlcv_aggregator /app/aggregator 

# Copy the internal libs 
COPY internal/ /app/internal 

# Run tests before building 
RUN go test ./... -v

# Build the Go binary for worker
RUN go build -o /bin/worker/server /app/worker

# Build the Go binary for aggregator 
RUN go build -o /bin/aggregator/server /app/aggregator

# Stage 2.1: Create a minimal runtime image for worker
FROM alpine:latest AS crypto_rates_ticks_worker_image

# Set working directory in runtime container
WORKDIR /app

COPY --from=builder /bin/worker . 
COPY config.yaml .

# Set the entrypoint
ENTRYPOINT ["./server"]

# Stage 2.2: Create a minimal runtime image for aggregator
FROM alpine:latest AS crypto_rates_ticks_aggregator_image

# Set working directory in runtime container
WORKDIR /app

COPY --from=builder /bin/aggregator . 
COPY config.yaml .

# Set the entrypoint
ENTRYPOINT ["./server"]
