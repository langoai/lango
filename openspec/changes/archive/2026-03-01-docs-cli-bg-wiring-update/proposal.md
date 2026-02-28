## Why

README, docs/cli, and docs/architecture documentation are out of sync with codebase after recent security hardening (keyring, SQLCipher, Cloud KMS), P2P session/sandbox features, and lifecycle management commits. Additionally, the `lango bg` CLI command exists in code but is not wired in main.go.

## What Changes

- Wire `lango bg` command in `cmd/lango/main.go` (stub provider since background.Manager is in-memory)
- Update README.md: Features (security expansion), CLI Commands (security keyring/db/kms, p2p session/sandbox, bg), Architecture (new packages), Skills description (removed built-in skills)
- Update docs/cli/index.md: Add security extension commands, P2P Network section, background task commands
- Update docs/index.md: Expand Security card description with keyring, SQLCipher, KMS
- Update docs/architecture/project-structure.md: Add lifecycle, keyring, sandbox, dbmigrate packages; update security and skills descriptions

## Capabilities

### New Capabilities
- `bg-cli-wiring`: Wire the existing `lango bg` CLI commands (list/status/cancel/result) into main.go

### Modified Capabilities
- `cli-reference`: Update CLI reference documentation to include all security, P2P, and background commands
- `project-docs`: Update README and architecture docs to reflect current package structure and removed skills

## Impact

- `cmd/lango/main.go` — new import and command registration
- `README.md` — Features, CLI Commands, Architecture sections
- `docs/cli/index.md` — Security, P2P, Automation tables
- `docs/index.md` — Security feature card
- `docs/architecture/project-structure.md` — package tables, skills description
