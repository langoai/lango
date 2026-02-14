## 1. State Layer

- [x] 1.1 Add `NewConfigStateWith(cfg *config.Config)` constructor to `state.go`
- [x] 1.2 Refactor `NewConfigState()` to delegate to `NewConfigStateWith(config.DefaultConfig())`

## 2. Wizard Layer

- [x] 2.1 Add `NewWizardWithConfig(cfg *config.Config)` constructor to `wizard.go`
- [x] 2.2 Remove `SaveConfig()` method from Wizard
- [x] 2.3 Add `Config() *config.Config` accessor method to Wizard

## 3. Menu Update

- [x] 3.1 Change "Save & Exit" description from "Write config to file" to "Save encrypted profile" in `menu.go`

## 4. Onboard Command Refactoring

- [x] 4.1 Add `--profile` flag to `NewCommand()` (default: "default")
- [x] 4.2 Update Long description to reference `~/.lango/lango.db` instead of `lango.json`
- [x] 4.3 Implement `runOnboard(profileName)` with bootstrap → load → TUI → save → activate flow
- [x] 4.4 Implement `loadOrDefault()` helper using `configstore.Store.Load()` with `ErrProfileNotFound` fallback
- [x] 4.5 Convert `PrintNextSteps` from Wizard method to `printNextSteps(cfg, profileName)` package function
- [x] 4.6 Add encrypted profile name, storage path, and profile management hints to post-save output
- [x] 4.7 Convert `generateEnvExample` from Wizard method to `generateEnvExample(cfg)` package function

## 5. Verification

- [x] 5.1 Verify `go build ./...` compiles successfully
- [x] 5.2 Verify existing tests pass (`go test ./internal/cli/onboard/...`)
