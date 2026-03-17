---
title: Docker
---

# Docker

Lango provides a Docker image that includes all runtime dependencies: Chromium for browser automation, `git` and `curl` for skill imports.

## Image Build

The Docker image uses a multi-stage build:

- **Builder stage**: `golang:1.25-bookworm` -- compiles the Go binary with CGO enabled (required by mattn/go-sqlite3 and sqlite-vec). Links against `libsqlcipher` for transparent database encryption support.
- **Runtime stage**: `debian:bookworm-slim` -- minimal runtime with Chromium and utilities

Build the image:

```bash
docker build -t lango:latest .
```

### Build Arguments

| Argument | Default | Description |
|----------|---------|-------------|
| `VERSION` | `dev` | Version string injected via `-ldflags` |
| `BUILD_TIME` | `unknown` | Build timestamp injected via `-ldflags` |
| `INSTALL_GO` | `false` | Install Go 1.25 toolchain in runtime image (for agents that need `go install`) |

Build with version info and Go toolchain:

```bash
docker build \
  --build-arg VERSION=1.2.0 \
  --build-arg BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --build-arg INSTALL_GO=true \
  -t lango:1.2.0 .
```

## Docker Compose

```yaml
services:
  lango:
    build: .
    image: lango:latest
    container_name: lango
    restart: unless-stopped
    ports:
      - "18789:18789"
      # - "9000:9000"   # P2P libp2p (uncomment to enable P2P networking)
    volumes:
      - lango-data:/home/lango/.lango
    secrets:
      - lango_config
      - lango_passphrase
    environment:
      - LANGO_PROFILE=default
      # Feature flags are managed via encrypted config profiles, not env vars.
      # Use 'lango settings' or 'lango config import' to configure features.

secrets:
  lango_config:
    file: ./config.json
  lango_passphrase:
    file: ./passphrase.txt

volumes:
  lango-data:
```

### Volumes

The named volume `lango-data` is mounted at `/home/lango/.lango` inside the container. This directory holds the encrypted database, skills, and configuration. Persisting this volume across container restarts preserves all agent state without re-importing config.

### P2P Networking

To expose the libp2p listener for P2P agent communication, uncomment the `9000:9000` port mapping and enable `p2p.enabled` in your config profile. The default listen addresses are `/ip4/0.0.0.0/tcp/9000` and `/ip4/0.0.0.0/udp/9000/quic-v1`.

### Presidio PII Profile

To include the [Presidio](https://microsoft.github.io/presidio/) PII redaction service:

```bash
docker compose --profile presidio up
```

## Headless Configuration

For automated or server deployments without interactive onboarding:

**1. Create `config.json`** with your provider keys and settings:

```json
{
  "agent": {
    "provider": "openai",
    "model": "gpt-4o"
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "botToken": "your-telegram-token"
    }
  }
}
```

**2. Create `passphrase.txt`** with your encryption passphrase:

```
your-secure-passphrase
```

**3. Start the container:**

```bash
docker compose up -d
```

## Entrypoint Script

The `docker-entrypoint.sh` script handles first-run setup:

1. Creates `~/.lango/skills` and `~/bin` directories
2. Verifies write permissions on critical directories -- named Docker volumes can inherit stale ownership from previous builds. If a directory is not writable, the script exits with instructions to recreate the volume.
3. Copies the passphrase secret to `~/.lango/keyfile` with mode `0600`
4. **First run**: copies `config.json` to `/tmp`, imports it into the encrypted config store, and the temp file is auto-deleted
5. **Subsequent restarts**: the existing encrypted profile is reused without re-importing

## Healthcheck

The Dockerfile includes a built-in healthcheck that runs `lango health` every 30 seconds:

```
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/lango", "health"]
```

Use `docker inspect --format='{{.State.Health.Status}}' lango` to check container health.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LANGO_PROFILE` | `default` | Configuration profile name |
| `LANGO_CONFIG_FILE` | `/run/secrets/lango_config` | Path to config JSON for import |
| `LANGO_PASSPHRASE_FILE` | `/run/secrets/lango_passphrase` | Path to passphrase file |

### Feature Flags

Feature flags (`agent.multiAgent`, `p2p.enabled`, `agentMemory.enabled`, `hooks.enabled`, `mcp.enabled`, etc.) are managed via encrypted config profiles, not environment variables. To configure features in a containerized deployment:

- **Pre-built config**: Include the desired flags in your `config.json` before first run. The entrypoint imports it via `lango config import`.
- **Running container**: Use `lango config set <key> <value>` or `lango settings` inside the container.
- **Docker secrets**: Mount an updated `config.json` as a secret and recreate the container (remove `lango.db` to trigger re-import).

## Related

- [Configuration](../getting-started/configuration.md) -- Full configuration reference
