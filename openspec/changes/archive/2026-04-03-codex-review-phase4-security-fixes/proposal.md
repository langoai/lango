## Why

Codex automated review of Phase 4 changes identified three security and correctness issues: (1) P2P browser navigation vulnerable to DNS rebinding due to conditional URL re-validation, (2) P2P sandbox executor gate not respecting `toolIsolation.enabled` config, and (3) filesystem `Delete` destroying symlink targets instead of the symlink itself, with subsequent rounds revealing path boundary bypass via OS aliases and symlink-as-config-entry edge cases.

## What Changes

- Remove `finalURL != rawURL` condition from P2P browser post-navigation validation so DNS rebinding attacks are blocked regardless of URL string equality
- Restore `toolIsolation.enabled` gate for P2P sandbox executor wiring with startup warning when P2P is enabled but tool isolation is disabled
- Rewrite `Delete` to use a symlink-specific validation flow: Lstat before resolve, canonicalize parent directory only, validate link location against blocked/allowed, delete the link itself
- Extract `checkPathAccess` helper from `validatePath` and make it compare config entries both resolved and unresolved to handle symlink-as-config-entry edge cases

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `tool-browser`: Post-navigation URL re-validation now always runs in P2P context (no string equality skip)
- `tool-filesystem`: Delete operation uses separate symlink-aware validation flow; `checkPathAccess` compares both resolved and unresolved config entries
- `container-sandbox`: P2P sandbox executor respects `toolIsolation.enabled` gate with explicit startup warning

## Impact

- `internal/tools/browser/tools.go` — Remove `finalURL != rawURL` condition
- `internal/app/app.go` — Restore `if cfg.P2P.ToolIsolation.Enabled` gate + warning log
- `internal/tools/filesystem/filesystem.go` — Rewrite `Delete`, extract `checkPathAccess`, dual config comparison
