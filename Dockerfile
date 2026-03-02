FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Install SQLite dev headers (required by sqlite-vec-go-bindings)
# Install libsqlcipher-dev for SQLCipher transparent DB encryption support
RUN apt-get update && apt-get install -y --no-install-recommends \
        libsqlite3-dev \
        libsqlcipher-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Version and build time injection (override via --build-arg)
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Build with CGO enabled (required by mattn/go-sqlite3 and sqlite-vec)
# Link against libsqlcipher for transparent DB encryption support
RUN CGO_ENABLED=1 go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" -o lango ./cmd/lango

# Runtime image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        chromium \
        git \
        curl \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r lango && useradd -r -g lango -m -d /home/lango lango \
    && mkdir -p /home/lango/.lango/skills \
    && mkdir -p /home/lango/bin \
    && chown -R lango:lango /home/lango/.lango /home/lango/bin

COPY --from=builder /app/lango /usr/local/bin/lango
COPY --from=builder /app/prompts/ /usr/share/lango/prompts/
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Optional: install Go toolchain for agents that need `go install` capability.
# Build with --build-arg INSTALL_GO=true to enable.
ARG INSTALL_GO=false
RUN if [ "$INSTALL_GO" = "true" ]; then \
        curl -fsSL https://go.dev/dl/go1.25.linux-amd64.tar.gz \
        | tar -C /usr/local -xzf - ; \
    fi

ENV PATH="/home/lango/bin:/home/lango/go/bin:/usr/local/go/bin:${PATH}"
ENV GOPATH="/home/lango/go"

USER lango
WORKDIR /home/lango

EXPOSE 18789

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/lango", "health"]

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
