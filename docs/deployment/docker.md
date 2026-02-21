---
title: Docker
---

# Docker

Lango provides a Docker image that includes all runtime dependencies: Chromium for browser automation, `git` and `curl` for skill imports.

## Image Build

The Docker image uses a multi-stage build:

- **Builder stage**: `golang:1.25-bookworm` -- compiles the Go binary with CGO enabled
- **Runtime stage**: `debian:bookworm-slim` -- minimal runtime with Chromium and utilities

Build the image:

```bash
docker build -t lango:latest .
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
    volumes:
      - lango-data:/data
    secrets:
      - lango_config
      - lango_passphrase
    environment:
      - LANGO_PROFILE=default

secrets:
  lango_config:
    file: ./config.json
  lango_passphrase:
    file: ./passphrase.txt

volumes:
  lango-data:
```

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

1. Copies the passphrase secret to `~/.lango/keyfile` with mode `0600`
2. **First run**: copies `config.json` to `/tmp`, imports it into the encrypted config store, and the temp file is auto-deleted
3. **Subsequent restarts**: the existing encrypted profile is reused without re-importing

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LANGO_PROFILE` | `default` | Configuration profile name |
| `LANGO_CONFIG_FILE` | -- | Path to config JSON for import |
| `LANGO_PASSPHRASE_FILE` | -- | Path to passphrase file |

## Related

- [Production Checklist](production.md) -- Pre-deployment security and configuration checks
- [Configuration](../getting-started/configuration.md) -- Full configuration reference
