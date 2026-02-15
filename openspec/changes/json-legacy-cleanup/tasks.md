## 1. Config System Cleanup

- [x] 1.1 Remove deprecated `config.Save()` function from `internal/config/loader.go`
- [x] 1.2 Remove `encoding/json` import from `internal/config/loader.go` (no longer needed after Save removal)

## 2. Config Import Auto-Delete

- [x] 2.1 Add `os.Remove(jsonPath)` to `MigrateFromJSON()` in `internal/configstore/migrate.go` after successful import
- [x] 2.2 Add warning log on delete failure (non-blocking)

## 3. Bootstrap Cleanup

- [x] 3.1 Remove `MigrationPath` field from `Options` struct in `internal/bootstrap/bootstrap.go`
- [x] 3.2 Remove `candidateJSONPaths()` function
- [x] 3.3 Simplify `handleNoProfile()` to only create default profile (remove JSON migration branch)

## 4. CLI Main Cleanup

- [x] 4.1 Remove `cfgFile` variable from `cmd/lango/main.go`
- [x] 4.2 Remove `--config` persistent flag from root command
- [x] 4.3 Remove all `MigrationPath: cfgFile` references in bootstrap calls
- [x] 4.4 Update `configImportCmd` short description and add "Source file deleted" message
- [x] 4.5 Update `configExportCmd` short description to indicate passphrase requirement

## 5. Doctor Command Update

- [x] 5.1 Remove `ConfigPath` from `Options` struct and `--config` flag in `internal/cli/doctor/doctor.go`
- [x] 5.2 Replace JSON file search loop with bootstrap-based config loading
- [x] 5.3 Rewrite `ConfigCheck.Run()` in `checks/config.go` to check encrypted profile
- [x] 5.4 Rewrite `ConfigCheck.Fix()` to guide user to `lango onboard`
- [x] 5.5 Remove `findConfigPath()` function
- [x] 5.6 Update tests in `checks_test.go` for new ConfigCheck behavior

## 6. README Update

- [x] 6.1 Replace Configuration section: remove JSON block, add encrypted profile description
- [x] 6.2 Replace Run section: remove `export` and `--config` commands
- [x] 6.3 Update CLI Commands table: update import/export descriptions
- [x] 6.4 Remove AI Providers Configuration Example JSON block
- [x] 6.5 Remove Embedding & RAG Configuration JSON block, replace with onboard guidance
- [x] 6.6 Remove Self-Learning JSON block and "configured via lango.json only" note
- [x] 6.7 Remove Security Configuration JSON block, replace with onboard guidance
- [x] 6.8 Remove Authentication JSON block, replace with onboard guidance
- [x] 6.9 Add Docker Headless Configuration section with import→delete pattern
- [x] 6.10 Remove `security.passphrase` DEPRECATED row from Config Reference table
- [x] 6.11 Remove `${ENV_VAR}` references from Config Reference table
- [x] 6.12 Add "All settings managed via lango onboard/config" note to Config Reference

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./...` passes
- [x] 7.3 Zero occurrences of `lango.json` in Go code (bootstrap, main, doctor)
- [x] 7.4 Zero occurrences of `config.Save(` in Go code
- [x] 7.5 Zero occurrences of `MigrationPath` in Go code
- [x] 7.6 Zero occurrences of `candidateJSONPaths` in Go code
- [x] 7.7 Zero occurrences of `lango.json`, `${`, `export .*_KEY` in README
