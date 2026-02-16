## Why

`NewBoltStore()` receives paths like `~/.lango/graph.db` but does not expand the tilde or create parent directories. BoltDB's `bolt.Open()` creates the file but not parent directories, causing initialization to fail with "no such file or directory" and the graph store to be silently skipped.

## What Changes

- Expand `~` prefix to `os.UserHomeDir()` in `NewBoltStore()` before opening the database
- Auto-create parent directories with `os.MkdirAll()` before `bolt.Open()`
- Align with the existing pattern used by session store (`internal/session/ent_store.go:71-76`)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `graph-store`: Add tilde expansion and parent directory auto-creation to `NewBoltStore()` path handling

## Impact

- `internal/graph/bolt_store.go`: Modified `NewBoltStore()` to handle path expansion and directory creation
- No API changes, no breaking changes
- Fixes silent graph store initialization failures on first run
