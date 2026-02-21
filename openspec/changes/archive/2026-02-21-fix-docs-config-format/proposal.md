## Why

All documentation files in `docs/` display configuration examples in YAML format, but the actual system stores config encrypted in SQLite via AES-256-GCM, uses `lango settings` TUI for interactive editing, and imports/exports JSON only (`lango config import/export`). Viper is configured with `SetConfigType("json")` and has no YAML file reading capability. Users reading these docs would incorrectly believe they can create a `config.yaml` file, which won't work.

## What Changes

- Convert all YAML config code blocks in documentation to JSON format across 23 doc files
- Add TUI navigation hints (`> **Settings:** lango settings -> <MenuName>`) before each config block showing how to reach that setting via the TUI
- Remove `# config.yaml` comments and YAML inline comments from config examples
- Add intro paragraph to `docs/configuration.md` explaining config is managed via TUI or JSON import
- Keep legitimate YAML as-is: Docker Compose in `docs/deployment/docker.md` and workflow DAG definitions in `docs/automation/workflows.md`

## Capabilities

### New Capabilities

_(none - this is a documentation-only change)_

### Modified Capabilities

_(none - no spec-level behavior changes, only documentation format corrections)_

## Impact

- **Documentation**: 23 files in `docs/` modified (configuration.md, features/*, security/*, automation/*, gateway/*, payments/*)
- **Code**: No code changes - all field names verified against `internal/config/types.go` struct tags
- **APIs**: No API changes
- **Dependencies**: None
