## Why

The Docker deployment files (Dockerfile, docker-compose.yml) and README documentation are out of sync with the current encrypted-profile architecture. The Dockerfile uses CGO_ENABLED=0 (incompatible with mattn/go-sqlite3), docker-compose.yml still mounts lango.json and passes API keys as environment variables (readable by the AI agent), and there is no secure entrypoint for headless config import. These gaps create both build failures and security risks in containerized deployments.

## What Changes

- Fix Dockerfile to use CGO_ENABLED=1 for SQLite CGO driver compatibility
- Add `docker-entrypoint.sh` for secure first-run config import using Docker secrets
- Replace lango.json bind mount and environment variable secrets with Docker secrets (tmpfs-backed)
- Copy passphrase to `~/.lango/keyfile` (agent-blocked path) instead of exposing via env var
- Update README Docker Headless Configuration section to document the secure secrets-based workflow
- Update `docker-deployment` spec to reflect new security model

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `docker-deployment`: Replace lango.json mount and env var secrets with Docker secrets, add entrypoint script, fix CGO build flag, update compose orchestration model

## Impact

- **Dockerfile**: CGO flag fix, entrypoint script addition
- **docker-compose.yml**: Secrets block, removed env var passthrough, profile env var
- **docker-entrypoint.sh**: New shell script for secure config bootstrap
- **README.md**: Docker Headless Configuration section rewritten
- **Security**: Agent can no longer read plaintext secrets at runtime (no env vars, keyfile path blocked)
