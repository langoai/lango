## Why

Docker containers running lango fail when creating skills in `.lango/skills` due to volume ownership mismatches, and CLI tool installation fails because the non-root `lango` user has no writable binary directory on PATH. These issues prevent agents from fully operating in containerized environments.

## What Changes

- Pre-create `.lango/skills/` subdirectory in Dockerfile with correct ownership to prevent Docker volume ownership drift
- Add `~/bin` as a user-writable binary installation path and include it in PATH
- Add runtime permission verification in `docker-entrypoint.sh` with fail-fast and actionable error messages
- Unify directory permission mode to `0700` across bootstrap and skill store (was inconsistent: bootstrap used `0700`, skill store used `0755`)
- Add writability probe in bootstrap to detect Docker volume ownership mismatches early
- Pre-create skills directory during bootstrap phase before skill system initialization
- Add optional Go toolchain installation via `--build-arg INSTALL_GO=true` for development images

## Capabilities

### New Capabilities

_(none — this change hardens existing capabilities)_

### Modified Capabilities

- `docker-deployment`: Add skills subdirectory pre-creation, user-writable bin path, PATH configuration, runtime permission verification, and optional Go toolchain build arg
- `bootstrap-pipeline`: Add writability probe for data directory, pre-create skills subdirectory, unify permission constant
- `skill-system`: Align directory permissions from 0755 to 0700 for consistency with parent data directory security posture

## Impact

- `Dockerfile`: New directory creation, ENV, optional build arg
- `docker-entrypoint.sh`: Permission verification loop, fail-fast on ownership mismatch
- `internal/bootstrap/phases.go`: New `dataDirPerm` constant, writability probe, skills dir pre-creation
- `internal/bootstrap/bootstrap.go`: Use shared permission constant
- `internal/skill/file_store.go`: Permission mode change (0755 → 0700) on 4 `MkdirAll` calls
