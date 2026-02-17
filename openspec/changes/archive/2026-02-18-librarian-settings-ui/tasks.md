## 1. Config Defaults

- [x] 1.1 Add Librarian defaults to `DefaultConfig()` in `internal/config/loader.go` (Enabled=false, ObservationThreshold=2, InquiryCooldownTurns=3, MaxPendingInquiries=2, AutoSaveConfidence="high")
- [x] 1.2 Add viper `SetDefault` calls for all 5 librarian fields in `Load()` after workflow defaults

## 2. Settings Menu

- [x] 2.1 Add `{"librarian", "Librarian", "Proactive knowledge extraction, inquiries"}` menu entry in `internal/cli/settings/menu.go` between "Workflow Engine" and "Save & Exit"

## 3. Settings Form

- [x] 3.1 Add `NewLibrarianForm(cfg *config.Config)` in `internal/cli/settings/forms_impl.go` with 7 fields: lib_enabled (InputBool), lib_obs_threshold (InputInt, positive), lib_cooldown (InputInt, non-negative), lib_max_inquiries (InputInt, non-negative), lib_auto_save (InputSelect: high/medium/low), lib_provider (InputSelect: "" + providers), lib_model (InputText)

## 4. State Update

- [x] 4.1 Add librarian cases to `UpdateConfigFromForm()` switch in `internal/cli/tuicore/state_update.go`: lib_enabled, lib_obs_threshold, lib_cooldown, lib_max_inquiries, lib_auto_save, lib_provider, lib_model

## 5. Editor Routing

- [x] 5.1 Add `case "librarian"` to `handleMenuSelection()` in `internal/cli/settings/editor.go` that creates the librarian form

## 6. Verification

- [x] 6.1 Run `go build ./...` and verify no compile errors
- [x] 6.2 Run `go test ./...` and verify all tests pass
