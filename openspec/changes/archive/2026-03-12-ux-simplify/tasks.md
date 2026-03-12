## 1. Config Presets

- [x] 1.1 Create `internal/config/presets.go` with PresetName type, PresetConfig(), IsValidPreset(), AllPresets()
- [x] 1.2 Create `internal/config/presets_test.go` with tests for all 4 presets plus validation
- [x] 1.3 Add `--preset` flag to `configCreateCmd()` in `cmd/lango/main.go`

## 2. Banner Startup Summary

- [x] 2.1 Add `GetVersion()` to `internal/cli/tui/banner.go`
- [x] 2.2 Add `FeatureLine` type and `StartupSummary()` function to banner.go
- [x] 2.3 Add `startupSummary(cfg)` helper in main.go called after `application.Start()`

## 3. Status Command

- [x] 3.1 Create `internal/cli/status/status.go` with NewStatusCmd, StatusInfo types, collectStatus, probeServer
- [x] 3.2 Create `internal/cli/status/render.go` with renderDashboard using lipgloss/tui styles
- [x] 3.3 Create `internal/cli/status/status_test.go` with 11 tests covering features, channels, rendering, JSON
- [x] 3.4 Wire status command in main.go with "start" group

## 4. Settings Tier System + Section Reorganization

- [x] 4.1 Add Tier field to Category struct (TierBasic=0, TierAdvanced=1)
- [x] 4.2 Add showAdvanced field to MenuModel and Tab key handler to toggle
- [x] 4.3 Add visibleCategories() method that respects tier filter
- [x] 4.4 Update renderGroupedView to filter by tier, skip empty sections
- [x] 4.5 Update help bar to show "Tab: Show Advanced" / "Tab: Show Basic"
- [x] 4.6 Reorganize sections: Core, AI & Knowledge, Automation, Payment & Account, P2P & Economy, Integrations, Security
- [x] 4.7 Assign tier to all 47 categories (~14 basic, ~33 advanced)
- [x] 4.8 Update settings.go Long description to match new sections

## 5. CLI Help Group Reorganization

- [x] 5.1 Change rootCmd groups from 4 to 5: start, ai, auto, net, sys
- [x] 5.2 Reassign all command GroupIDs to new groups
- [x] 5.3 Remove "core" GroupID from healthCmd function

## 6. Onboard Improvements

- [x] 6.1 Add `--preset` flag to onboard command
- [x] 6.2 Update loadOrDefault to accept preset parameter
- [x] 6.3 Rewrite printNextSteps to accept config and show disabled feature recommendations
- [x] 6.4 Add preset command hints to next steps output

## 7. Verification

- [x] 7.1 Run `go build ./...` — all packages build
- [x] 7.2 Run `go test ./...` — all tests pass
- [x] 7.3 Verify `lango --help` shows 5 new groups
- [x] 7.4 Verify `lango status --help` registered correctly
- [x] 7.5 Verify `lango config create --help` shows --preset flag
