## 1. Theme & Icons

- [x] 1.1 Create `internal/cli/cockpit/theme/theme.go` with Surface0-3, TextPrimary/Secondary/Tertiary, BorderFocused/Default/Subtle color constants
- [x] 1.2 Create `internal/cli/cockpit/theme/icons.go` with unicode icon constants (◉⚙⚡◍◈)
- [x] 1.3 Create `internal/cli/cockpit/theme/logo.go` with `RenderLogo()` function — squirrel ASCII art with color gradient

## 2. Sidebar Component

- [x] 2.1 Create `internal/cli/cockpit/sidebar/sidebar.go` — `Model` struct implementing `tea.Model` with menu items, active highlight, non-interactive (no key consumption), `SetHeight(int)`, `SetVisible(bool)`, fixed 20ch width

## 3. Cockpit Root Model

- [x] 3.1 Create `internal/cli/cockpit/deps.go` — `Deps` struct (TurnRunner, Config, SessionKey)
- [x] 3.2 Create `internal/cli/cockpit/keymap.go` — `keyMap` struct with `ToggleSidebar` binding (Ctrl+B)
- [x] 3.3 Create `internal/cli/cockpit/cockpit.go` — `Model` struct with `childModel` interface, `New(deps)` constructor, `Init()`/`Update()`/`View()` methods, consume-or-forward delegation, synthetic WindowSizeMsg on toggle, `SetProgram(p)` delegation, `sidebarWidth()` helper
- [x] 3.4 Compile-time interface check: `var _ childModel = (*chat.ChatModel)(nil)`

## 4. Cockpit Command

- [x] 4.1 Add `cockpitCmd` Cobra subcommand in `cmd/lango/main.go` — `lango cockpit` with short description
- [x] 4.2 Implement `runCockpit()` function: cliboot.BootResult → App.New(WithLocalChat) → App.Start → cockpit.New(deps) → tea.NewProgram → model.SetProgram(p) → CompositeProvider type assertion for SetTTYFallback → p.Run()

## 5. Tests

- [x] 5.1 Create `internal/cli/cockpit/cockpit_test.go` — mock childModel, TestConsumeOrForward (ChunkMsg/DoneMsg/ApprovalRequestMsg), TestCtrlB_SyntheticResize, TestCtrlB_WidthCalculation, TestWindowSizeMsg_ReducedWidth, TestCockpitOnly_CtrlB, TestSetProgram_Delegation
- [x] 5.2 Create `internal/cli/cockpit/sidebar/sidebar_test.go` — TestSidebarView, TestSidebarHidden, TestSidebarHeight
- [x] 5.3 Create `internal/cli/cockpit/theme/theme_test.go` — TestColorsNotEmpty, TestIconsNotEmpty, TestRenderLogo
- [x] 5.4 Run `go build ./...` && `go test ./...` && `go vet ./...`
