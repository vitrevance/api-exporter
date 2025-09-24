# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY pkg pkg
COPY main main

# Build the statically linked binary without CGO
RUN CGO_ENABLED=0 go build -o api-exporter ./main

# Final stage using scratch
FROM scratch

# Copy CA certificates for HTTPS if needed (optional, remove if not required)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /app/api-exporter /api-exporter

# Command to run
ENTRYPOINT ["/api-exporter"]
CMD ["-config", "config.yaml"]
