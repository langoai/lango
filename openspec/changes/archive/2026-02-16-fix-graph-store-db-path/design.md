## Context

`NewBoltStore(path)` passes the path directly to `bolt.Open()`. BoltDB creates the file but not parent directories. When the config specifies `~/.lango/graph.db`, the tilde is not expanded by the OS (only shells do this), and the parent directory may not exist on first run. The session store already handles both cases at `internal/session/ent_store.go:71-76`.

## Goals / Non-Goals

**Goals:**
- Expand `~` prefix to the user's home directory in `NewBoltStore()`
- Auto-create parent directories before opening the BoltDB file
- Follow the same pattern used by session store for consistency

**Non-Goals:**
- Centralizing path expansion into a shared utility (single caller, not worth the abstraction)
- Handling environment variables like `$HOME` in paths (not needed, tilde is the only issue)

## Decisions

1. **Inline tilde expansion in `NewBoltStore()`** rather than a shared helper.
   - Only two callers in the codebase (session store, graph store), each in different packages.
   - A shared utility would add a dependency for a 5-line operation.

2. **Use `os.MkdirAll` with `0o700` permissions** for parent directory creation.
   - Matches session store convention.
   - Restrictive permissions appropriate for a data directory.

3. **Expand only `~/` prefix**, not standalone `~` or `~user/`.
   - `~user/` expansion requires system user lookup — unnecessary complexity.
   - Config values always use `~/` form.

## Risks / Trade-offs

- [Duplicate logic] Two places now do tilde expansion independently → Acceptable for two callers; extract if a third appears.
