## Why

Line-oriented input reading is currently duplicated across the CLI prompt package, TTY approval fallback, and non-interactive passphrase stdin acquisition. That duplication keeps a low-level I/O concern scattered across layers instead of being handled once in a shared helper.

## What Changes

- Add a small shared line-reader helper package for raw line-oriented input
- Reuse it from `internal/cli/prompt`, `internal/approval`, and `internal/security/passphrase`
- Add focused helper tests and keep existing behavior-specific tests green
- Record the shared helper contract in OpenSpec

## Capabilities

### New Capabilities
- `shared-line-reader`: common raw line-oriented input helper for internal prompt-like flows

### Modified Capabilities
- `cli-prompt-helpers`: CLI prompt helpers build on the shared raw line reader
- `channel-approval`: TTY approval fallback uses the shared raw line reader
- `passphrase-acquisition`: stdin-pipe acquisition uses the shared raw line reader

## Impact

- Affected code: `internal/lineio/*`, `internal/cli/prompt/*`, `internal/approval/*`, `internal/security/passphrase/*`
- Affected specs: new `shared-line-reader` plus updates to existing capabilities above
- No user-facing behavior change; this is an internal architecture consolidation
