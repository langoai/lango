## Why

Lango's core flow (onboard → serve) works well, but the surrounding UX has five problems: serve startup is a black box (no feature visibility), Settings TUI exposes 47 categories without filtering, CLI help uses module-centric groups instead of user-intent groups, onboard provides no guidance after completion, and there's no unified status command.

## What Changes

- Add startup summary to `lango serve` showing activated features, channels, and gateway address
- Add Basic/Advanced tier system to Settings TUI with Tab toggle (14 basic vs 47 total categories)
- Reorganize Settings sections: Infrastructure(11) → Automation(3) + Payment & Account(5) + P2P & Economy(12) + Integrations(3)
- Reorganize CLI help from 4 groups to 5: Getting Started, AI & Knowledge, Automation, Network & Economy, Security & System
- Add `lango status` unified dashboard combining health, config, and feature status
- Add preset profile system (`lango config create --preset researcher`) with 4 presets: minimal, researcher, collaborator, full
- Improve onboard completion with config-aware feature recommendations and preset hints
- Add `--preset` flag to onboard command

## Capabilities

### New Capabilities
- `cli-status-dashboard`: Unified status command combining health/config/metrics into a single dashboard
- `config-presets`: Purpose-built configuration presets (minimal, researcher, collaborator, full)
- `settings-tier-system`: Basic/Advanced tier filtering for Settings TUI categories

### Modified Capabilities
- `brand-banner`: Add startup summary rendering (StartupSummary, FeatureLine type)
- `cli-command-groups`: Reorganize from 4 to 5 user-intent groups
- `cli-settings`: Reorganize sections and add tier filtering with Tab toggle
- `cli-onboard`: Add --preset flag and config-aware next steps guide

## Impact

- `internal/cli/tui/banner.go` — new GetVersion(), StartupSummary(), FeatureLine type
- `internal/cli/settings/menu.go` — Category.Tier field, showAdvanced toggle, section reorganization
- `internal/cli/status/` — new package (status.go, render.go, status_test.go)
- `internal/config/presets.go` — new PresetConfig(), IsValidPreset(), AllPresets()
- `cmd/lango/main.go` — CLI groups, status command wiring, --preset flag on config create
- `internal/cli/onboard/onboard.go` — --preset flag, improved printNextSteps
- `internal/cli/settings/settings.go` — updated Long description
