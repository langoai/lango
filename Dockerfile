FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Install SQLite dev headers (required by sqlite-vec-go-bindings)
RUN apt-get update && apt-get install -y libsqlite3-dev && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with CGO enabled (required by mattn/go-sqlite3)
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o lango ./cmd/lango

# Runtime image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    chromium \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Set Chrome path for rod
ENV ROD_BROWSER=/usr/bin/chromium

# Create user and group
RUN groupadd -r lango && useradd -r -g lango -m -d /home/lango lango

WORKDIR /app

COPY --from=builder /app/lango /usr/local/bin/lango
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Create data directory and set permissions
RUN mkdir -p /data && chown -R lango:lango /data && chmod 700 /data

# Switch to non-root user
USER lango

EXPOSE 18789

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:18789/health || exit 1

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
