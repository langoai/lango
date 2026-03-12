## Context

Lango's CLI has grown to 24+ commands and 47 settings categories. The core onboard→serve flow works well, but users face friction after initial setup: no startup feedback, overwhelming settings menu, no unified status view, and no fast path for purpose-specific configurations. All changes are additive UI/UX improvements with no breaking changes to internal APIs.

## Goals / Non-Goals

**Goals:**
- Provide immediate visual feedback after `lango serve` starts (which features are active)
- Reduce Settings TUI cognitive load from 47 to ~14 essential categories by default
- Reorganize CLI help into user-intent groups (Getting Started, AI, Automation, Network, Security)
- Provide a single `lango status` command for unified system overview
- Enable purpose-built profiles via presets (researcher, collaborator, full)
- Guide users to relevant features after onboard completion

**Non-Goals:**
- Changing any internal/core APIs or data models
- Modifying bootstrap, config encryption, or storage mechanisms
- Adding new features beyond UX/CLI layer changes
- Removing or deprecating any existing commands

## Decisions

### 1. Tier-based filtering over separate menus
Categories get a `Tier int` field (Basic=0, Advanced=1) rather than separate menu models. Tab toggles visibility. Search always searches all tiers. This keeps the data model simple and avoids duplicating navigation logic.

### 2. Section reorganization based on user intent
Infrastructure(11 items) split into Automation(3), Payment & Account(5), P2P & Economy(12), Integrations(3). Economy and P2P merged since they're used together. This matches how features are actually used rather than how they're implemented.

### 3. Config-based startup summary
`StartupSummary()` reads config flags directly rather than querying the running app. This is simpler and available immediately after `app.Start()` returns, without needing health endpoints.

### 4. Presets as config overlays
`PresetConfig()` calls `DefaultConfig()` then overrides specific fields. No separate config schema or preset files needed. Presets are code-defined (4 hardcoded), not user-extensible.

### 5. Status command probes server optionally
`lango status` works both with and without a running server. Config-based info is always shown. Server health is probed with a 3s timeout and gracefully degrades if unreachable.

## Risks / Trade-offs

- [Tier assignments are subjective] → Based on onboard's 5-step coverage as "basic" baseline; Tab toggle gives full access
- [Preset configs may become stale as features evolve] → Presets only set `Enabled` booleans; detailed config still requires settings editor
- [CLI group changes affect muscle memory] → Old group IDs gone, but command names unchanged; only `--help` output differs
