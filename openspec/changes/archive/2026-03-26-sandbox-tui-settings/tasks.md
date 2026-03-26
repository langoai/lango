## 1. Form Implementation

- [x] 1.1 Create `internal/cli/settings/forms_sandbox.go` with `NewOSSandboxForm()` — 9 fields with `os_sandbox_*` keys
- [x] 1.2 Add `os_sandbox` category to Security section in `internal/cli/settings/menu.go`
- [x] 1.3 Add `case "os_sandbox"` to `createFormForCategory()` in `internal/cli/settings/setup_flow.go`
- [x] 1.4 Add `case "os_sandbox"` to `categoryIsEnabled()` in `internal/cli/settings/editor.go`

## 2. State Update Handlers

- [x] 2.1 Add 9 `os_sandbox_*` case handlers to `UpdateConfigFromForm()` in `internal/cli/tuicore/state_update.go`

## 3. Tests

- [x] 3.1 `TestNewOSSandboxForm_AllFields` — 9 fields, select types verified
- [x] 3.2 `TestNewMenuModel_HasOSSandboxCategory` — menu contains `os_sandbox`
- [x] 3.3 `TestUpdateConfigFromForm_OSSandbox_IsolatesFromP2P` — all 9 fields map to `cfg.Sandbox.*`, P2P config unchanged

## 4. Verification

- [x] 4.1 `go build ./...` passes
- [x] 4.2 `go test ./internal/cli/settings/...` passes
- [x] 4.3 `go test ./internal/cli/tuicore/...` passes
