## 1. PostLoad One-Stop Function

- [x] 1.1 Add exported `PostLoad(*Config) error` to `internal/config/loader.go` chaining: MigrateEmbeddingProvider, substituteEnvVars, NormalizePaths, ValidateDataPaths, Validate
- [x] 1.2 Refactor `Load()` to replace 5 individual calls with single `PostLoad(cfg)` call

## 2. Bootstrap Phase Fix

- [x] 2.1 Remove early returns from all 3 branches in `phaseLoadProfile()` in `internal/bootstrap/phases.go`
- [x] 2.2 Add single `config.PostLoad(s.Result.Config)` call at end of `phaseLoadProfile()` after all branches
- [x] 2.3 Add `config` import to phases.go

## 3. Store.Save PostLoad

- [x] 3.1 Add `config.PostLoad(cfg)` call at start of `Store.Save()` in `internal/configstore/store.go` before marshal

## 4. Config Set Cleanup Pattern

- [x] 4.1 Change `NewSetCmd` cfgLoader signature to `func() (*config.Config, func(), error)` in `internal/cli/configcmd/getset.go`
- [x] 4.2 Update RunE to call `defer cleanup()` with nil check
- [x] 4.3 Update `cmd/lango/main.go` config set wiring: single bootstrap, cleanup closure that closes DBClient

## 5. Collaborator Preset Fix

- [x] 5.1 Add `cfg.Payment.Network.RPCURL = "https://sepolia.base.org"` to collaborator case in `internal/config/presets.go`

## 6. Verification

- [x] 6.1 Run `go build ./...` — all packages compile
- [x] 6.2 Run `go test ./internal/config/... ./internal/configstore/... ./internal/bootstrap/... ./internal/cli/configcmd/...` — all tests pass
