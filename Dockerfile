# Build stage
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o git-file-sync .

# Final stage
FROM debian:bookworm-slim

# Install git and openssh-client
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        git \
        openssh-client \
        ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/git-file-sync /usr/local/bin/

# Create non-root user
RUN useradd -u 1000 -m -s /bin/bash syncer
USER syncer

ENTRYPOINT ["git-file-sync"]
