## Context

Docker containers running lango encounter two classes of permission failures:

1. **Skill directory ownership mismatch**: When a Docker named volume (`lango-data`) is created from a previous build with a different UID for the `lango` user, the volume retains stale ownership. The Go bootstrap creates `~/.lango/` with `os.MkdirAll` which succeeds silently for existing directories regardless of ownership, causing later writes to fail deep in the skill store.

2. **No writable binary path**: The runtime image runs as non-root `lango` user with no writable directory on PATH. Agents using the exec tool cannot install CLI tools (`go install`, downloads to `/usr/local/bin/`).

Additionally, the codebase has an inconsistency: bootstrap creates `~/.lango/` with `0700` while `FileSkillStore` creates subdirectories with `0755`, which is unnecessarily permissive for a directory containing encrypted secrets.

## Goals / Non-Goals

**Goals:**
- Docker containers detect and report volume ownership mismatches at startup with actionable error messages
- Skills directory is pre-created with correct ownership before the skill system initializes
- A user-writable binary directory exists on PATH for CLI tool installation
- Directory permission modes are consistent (0700) across all `~/.lango/` operations
- Optional Go toolchain available via build argument for development images

**Non-Goals:**
- Automatic ownership repair (too dangerous — user should decide)
- Root-level operations in container (breaks security model)
- Changing the `blockedPaths` mechanism for `~/.lango/` (agent filesystem tool should remain blocked; meta-tools handle skill operations)

## Decisions

### Decision 1: Writability probe instead of ownership check
Use a file-write probe (`os.WriteFile` + `os.Remove`) to verify directory writability rather than checking UID ownership directly.

**Rationale**: A write probe catches all failure modes (UID mismatch, read-only filesystem, permission bits) with a single test. Checking `os.Getuid()` against `stat.Uid` only catches one scenario and requires platform-specific code.

**Alternative considered**: `syscall.Stat_t` UID comparison — platform-specific, doesn't cover all cases.

### Decision 2: Unified 0700 permission mode
Align all `~/.lango/` directory creation to 0700 (owner-only rwx), replacing the skill store's 0755.

**Rationale**: The `.lango/` directory contains encrypted database and keyfiles. The skill store's 0755 was gratuitously permissive. Since only the owner process accesses these directories, 0700 is the correct security posture.

### Decision 3: `~/bin` on PATH via Dockerfile ENV
Add `~/bin` as a user-writable binary directory and configure PATH in the Dockerfile.

**Rationale**: The non-root user cannot write to `/usr/local/bin/`. Providing `~/bin` follows Unix conventions and requires no runtime configuration. Setting PATH via `ENV` ensures it's available in both entrypoint and exec tool commands.

### Decision 4: Optional Go toolchain via build argument
Provide `INSTALL_GO=false` build arg rather than always including Go in the runtime image.

**Rationale**: Go toolchain adds ~500MB to the image. Most production deployments don't need it. Build args allow opt-in without branching Dockerfiles.

### Decision 5: Dual-layer permission check (entrypoint + bootstrap)
Check permissions in both `docker-entrypoint.sh` (shell-level) and `phases.go` (Go-level).

**Rationale**: The entrypoint catches problems before any Go code runs, providing immediate shell-level error messages. The bootstrap probe catches issues in non-Docker environments and serves as defense-in-depth.

## Risks / Trade-offs

- **[Write probe leaves artifact on crash]** → Probe file is removed immediately after test; if process crashes between write and remove, a 0-byte `.write-test` file remains (harmless, cleaned up on next run)
- **[INSTALL_GO build arg increases complexity]** → Guarded by `if [ "$INSTALL_GO" = "true" ]`; no-op when disabled; no image size impact on default builds
- **[Entrypoint exit on permission failure]** → Fail-fast is intentional; provides actionable hint (`docker volume rm lango-data`) instead of cryptic Go stack traces
