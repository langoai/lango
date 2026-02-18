ARG WITH_BROWSER=false

FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Install SQLite dev headers (required by sqlite-vec-go-bindings)
RUN apt-get update && apt-get install -y --no-install-recommends libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with CGO enabled (required by mattn/go-sqlite3)
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o lango ./cmd/lango

# Runtime image
FROM debian:bookworm-slim
ARG WITH_BROWSER=false

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && if [ "$WITH_BROWSER" = "true" ]; then \
         apt-get install -y --no-install-recommends chromium; \
       fi \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r lango && useradd -r -g lango -m -d /home/lango lango \
    && mkdir -p /data && chown -R lango:lango /data && chmod 700 /data

COPY --from=builder /app/lango /usr/local/bin/lango
COPY --from=builder /app/prompts/ /usr/share/lango/prompts/
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

USER lango
WORKDIR /home/lango

EXPOSE 18789

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/lango", "health"]

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["serve"]
