## Context

Documentation files across `docs/` display configuration examples in YAML format. However, the Lango system:
- Stores config encrypted in SQLite (`~/.lango/lango.db`) via `config_profile` table (AES-256-GCM)
- Uses `lango settings` TUI for interactive editing (21 menu categories)
- Imports/exports JSON only (`lango config import/export`)
- Has no YAML config file reading (Viper uses `SetConfigType("json")`)

This is a documentation-only change with no code modifications.

## Goals / Non-Goals

**Goals:**
- Convert all YAML config blocks to valid JSON in documentation
- Add TUI navigation hints so users know how to reach each setting
- Preserve legitimate YAML (Docker Compose, workflow DAG definitions)

**Non-Goals:**
- Adding YAML config file support to the system
- Changing any Go source code or config struct tags
- Modifying the TUI settings interface

## Decisions

1. **JSON format for config examples** -- Matches the actual `lango config import/export` format. Users can copy-paste JSON directly for import. Field names verified against `internal/config/types.go` struct tags.

2. **TUI navigation hints as blockquotes** -- Using `> **Settings:** lango settings -> <MenuName>` format provides a visual cue without disrupting the document flow. Menu names verified against `internal/cli/settings/` form definitions.

3. **Exceptions for Docker Compose and workflow DAGs** -- These are legitimately YAML file formats (not Lango config), so they remain as-is.

## Risks / Trade-offs

- [Stale documentation] -> Field names were verified against current `types.go` struct tags. Future config changes must update docs accordingly.
- [Missing files] -> Full grep verification ensures no YAML config blocks were missed.
