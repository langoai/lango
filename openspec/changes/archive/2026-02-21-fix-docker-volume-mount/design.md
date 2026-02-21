## Context

The Docker deployment mounts a named volume `lango-data` at `/data`, but the application stores all persistent state in `~/.lango/` (`/home/lango/.lango/`), including `lango.db`, `graph.db`, and `skills/`. The `/data` directory was created in the Dockerfile but never used by the application. This mismatch means container recreation causes full data loss.

## Goals / Non-Goals

**Goals:**
- Align the Docker volume mount with the application's actual data directory
- Ensure `lango.db`, `graph.db`, and `skills/` persist across container lifecycle events
- Remove dead infrastructure code (unused `/data` directory)

**Non-Goals:**
- Changing the application's data directory paths
- Supporting host bind mounts (users can configure this themselves)
- Data migration tooling for existing deployments

## Decisions

### Mount path: `/home/lango/.lango` instead of `/data`

**Rationale**: The application defaults all persistent data paths to `~/.lango/`. Changing the mount target to match the application is simpler and less invasive than configuring the application to use `/data`.

**Alternative considered**: Configure the app to use `/data` via config overrides. Rejected because it would require config changes in the entrypoint and diverge Docker from non-Docker behavior.

### Remove `/data` directory from Dockerfile

**Rationale**: With the volume no longer mounted at `/data`, the directory creation (`mkdir -p /data && chown && chmod`) is dead code. Removing it keeps the Dockerfile clean.

## Risks / Trade-offs

- [Existing deployments with `lango-data` volume at `/data`] → No data was stored there, so no migration needed. Users must recreate containers for the new mount to take effect.
- [Entrypoint writes to `~/.lango/`] → The entrypoint already does `mkdir -p "$HOME/.lango"` and writes `keyfile` there. With the named volume mounted, Docker initializes the volume on first use, then the entrypoint populates it. No conflict.
