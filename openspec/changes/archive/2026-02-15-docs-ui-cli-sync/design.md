## Context

Lango's configuration storage was unified to encrypted SQLite profiles (AES-256-GCM via `~/.lango/lango.db`). Docker deployment artifacts still reference the old pattern: mounting `lango.json` as a read-only bind mount and injecting API keys as environment variables. Both patterns expose plaintext secrets to the AI agent at runtime—bind-mounted files are readable, and environment variables are accessible via the exec tool.

The passphrase acquisition priority in `internal/configstore/crypto.go` is: keyfile (`~/.lango/keyfile`) → interactive terminal → stdin pipe. The agent's filesystem tool blocks `~/.lango/` access, making keyfile the ideal secure storage location for Docker deployments.

## Goals / Non-Goals

**Goals:**
- Eliminate all plaintext secret exposure to the AI agent in Docker deployments
- Fix CGO build flag for mattn/go-sqlite3 compatibility
- Provide a secure first-run config import flow via entrypoint script
- Document the Docker secrets workflow in README

**Non-Goals:**
- Kubernetes/Helm chart support (separate change)
- Config export from running container
- Multi-profile Docker support beyond LANGO_PROFILE env var

## Decisions

### Decision 1: Docker secrets over environment variables

**Choice**: Use Docker secrets (tmpfs at `/run/secrets/`) for both config and passphrase.

**Rationale**: Environment variables are readable by the agent's exec tool. Docker secrets are tmpfs-backed and readable only by the entrypoint script (before exec replaces the process with lango). The entrypoint copies secrets to safe locations and the originals remain in `/run/secrets/` which the agent cannot enumerate.

**Alternative considered**: Encrypted environment variables — rejected because the decryption key would itself need to be stored somewhere accessible.

### Decision 2: Entrypoint script for config bootstrap

**Choice**: A `docker-entrypoint.sh` that handles passphrase keyfile setup and first-run config import before `exec lango`.

**Rationale**: The entrypoint runs as PID 1 (before lango starts) so the agent has no opportunity to intercept the plaintext config. After import, the JSON temp copy is auto-deleted by `config import`. The passphrase is copied to `~/.lango/keyfile` (chmod 0600, agent-blocked path) so the crypto provider can acquire it non-interactively.

### Decision 3: CGO_ENABLED=1 in Dockerfile

**Choice**: Set `CGO_ENABLED=1` in builder stage using `golang:1.22-bookworm` (which includes gcc).

**Rationale**: `mattn/go-sqlite3` is a CGO-based driver. The existing Makefile already uses `CGO_ENABLED=1`. The bookworm base image includes build-essential tools. No additional apt-get installs needed.

## Risks / Trade-offs

- **[Image size increase]** → CGO_ENABLED=1 produces a dynamically linked binary requiring libc in the runtime image. Already mitigated by using debian:bookworm-slim (which includes libc).
- **[Secret file on host]** → `config.json` and `passphrase.txt` must exist on the host for docker-compose. Users should delete them after first `docker-compose up`. Documented in README.
- **[Restart idempotency]** → Entrypoint checks if profile already exists before re-importing. Safe for `restart: unless-stopped`.
