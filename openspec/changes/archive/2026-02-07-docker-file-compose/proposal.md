## Why

To make lango operational in Docker container environments, Dockerfile improvements and docker-compose.yml addition are needed. The current Dockerfile does not account for passphrase handling in headless environments or Companion RPC Provider integration, and the lack of docker-compose makes deployment and operations difficult.

## What Changes

- **Dockerfile improvements**: Health check addition, non-root user execution, security hardening
- **docker-compose.yml addition**: lango service definition, volume mounts, environment variable configuration
- **RPC Provider mode enforcement**: Use RPC Provider (Companion) instead of LocalCryptoProvider in Docker environments
- **Channel configuration**: All channels supported (Discord, Telegram, Slack)
- **Browser Tool retained**: Chromium included (required for tool-browser usage)

## Capabilities

### New Capabilities
- `docker-deployment`: Requirements related to lango deployment via Docker and docker-compose

### Modified Capabilities
- `secure-signer`: LocalCryptoProvider disabled and RPC Provider forced in Docker environments

## Impact

- `Dockerfile`: Improvements and security hardening
- `docker-compose.yml`: Newly created
- `internal/security`: Headless environment detection logic verification needed
- Operations documentation: Docker deployment guide addition needed
