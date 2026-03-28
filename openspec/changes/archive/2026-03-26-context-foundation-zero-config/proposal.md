## Why

Lango's context engineering subsystems (Knowledge, Observational Memory, Embedding/RAG, Graph Store, Librarian) require 26+ individual config knobs to set up. When a feature is silently disabled (e.g., empty embedding provider), users see no diagnostic output — it just doesn't work. There is no single-line way to activate a reasonable default configuration, and no structured mechanism to answer "what's off and why."

## What Changes

- **Context Profile system**: A new `contextProfile` config field (`off`/`lite`/`balanced`/`full`) that bundles knowledge, memory, embedding, graph, and librarian settings into named presets. One line replaces 26+ knobs.
- **Explicit key tracking**: `config.Load()` returns `*LoadResult{Config, ExplicitKeys}` instead of `*Config`, using a raw viper instance (no defaults) to detect which keys the user actually set in their config file. **BREAKING**: All `config.Load()` callers must adapt to the new return type.
- **FeatureStatus shared type**: A new `types.FeatureStatus` struct (name, enabled, healthy, reason, suggestion) used by doctor checks, status command, and TUI — replacing scattered ad-hoc status reporting.
- **StatusCollector**: Wiring functions (`initEmbedding`, `initKnowledge`, `initMemory`, `initGraph`) return `*types.FeatureStatus` alongside their components, collected by a `StatusCollector` in the app layer.
- **Doctor check**: A new "Context Engineering" check in `lango doctor` that reports context profile, silent-disabled feature count, and actionable suggestions.
- **Status command enrichment**: `lango status` shows profile name and per-feature reason/suggestion via `FeatureStatus` adapter.

## Capabilities

### New Capabilities
- `context-profile`: Profile-based preset system for context subsystem configuration (off/lite/balanced/full) with explicit-key-aware override protection
- `feature-status`: Shared FeatureStatus type and StatusCollector for structured init diagnostics across doctor, status, and TUI

### Modified Capabilities
- `config-system`: `Load()` return type changes to `*LoadResult`, `ApplyContextProfile()` added to post-load pipeline
- `cli-doctor`: New "Context Engineering" check added to `AllChecks()`
- `cli-status-dashboard`: Feature list enriched with profile name and reason/suggestion detail

## Impact

- **Config loader** (`internal/config/loader.go`): `Load()` signature change — all callers (`cmd/lango/main.go`, `internal/bootstrap/bootstrap.go`, `internal/configstore/migrate.go`, `internal/cli/doctor/doctor.go`) must adapt
- **Types** (`internal/types/`): New `feature_status.go` file
- **App wiring** (`internal/app/wiring_*.go`): 5 wiring functions gain `*types.FeatureStatus` return value
- **CLI** (`internal/cli/doctor/checks/`, `internal/cli/status/`): New check file + adapter files
- **No external dependency changes**
