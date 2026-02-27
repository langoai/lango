## 1. Banner Component

- [x] 1.1 Create `internal/cli/tui/banner.go` with SetVersionInfo, SetProfile, squirrelFace, Banner, BannerBox, ServeBanner
- [x] 1.2 Create `internal/cli/tui/banner_test.go` with tests for all banner functions

## 2. Version Injection

- [x] 2.1 Import `tui` in `cmd/lango/main.go` and call `tui.SetVersionInfo(Version, BuildTime)` at startup
- [x] 2.2 Call `tui.SetProfile(profileName)` in `settings.go` runSettings before TUI launch
- [x] 2.3 Call `tui.SetProfile(profileName)` in `onboard.go` runOnboard before TUI launch

## 3. Screen Clear

- [x] 3.1 Change `Editor.Init()` to return `tea.ClearScreen` instead of nil
- [x] 3.2 Change `Wizard.Init()` to return `tea.ClearScreen` instead of nil

## 4. Banner Integration

- [x] 4.1 Replace `viewWelcome()` box in `editor.go` with `tui.BannerBox()` + description text
- [x] 4.2 Replace title in `wizard.go` View() with `tui.Banner()` + "Setup Wizard" subtitle
- [x] 4.3 Add `tui.ServeBanner()` output in `serveCmd()` after logging init

## 5. Verification

- [x] 5.1 Run `go build ./...` — build passes
- [x] 5.2 Run `go test ./internal/cli/tui/...` — all banner tests pass
- [x] 5.3 Run `go test ./internal/cli/...` — all CLI tests pass
