## 1. Go Core — Permission Constant and Bootstrap Hardening

- [x] 1.1 Add `dataDirPerm` constant (0700) to `internal/bootstrap/phases.go`
- [x] 1.2 Update `phaseEnsureDataDir` to use `dataDirPerm` instead of hardcoded 0700
- [x] 1.3 Add writability probe (write + remove `.write-test` file) after MkdirAll in `phaseEnsureDataDir`
- [x] 1.4 Add skills directory pre-creation (`~/.lango/skills/`) in `phaseEnsureDataDir`
- [x] 1.5 Update `openDatabase` in `internal/bootstrap/bootstrap.go` to use `dataDirPerm`

## 2. Skill Store — Permission Alignment

- [x] 2.1 Change `FileSkillStore.Save()` directory creation from 0755 to 0700 in `internal/skill/file_store.go`
- [x] 2.2 Change `FileSkillStore.SaveResource()` directory creation from 0755 to 0700
- [x] 2.3 Change `FileSkillStore.EnsureDefaults()` directory creation from 0755 to 0700 (both skills root and individual skill dirs)

## 3. Docker — Dockerfile Improvements

- [x] 3.1 Pre-create `.lango/skills/` and `~/bin` directories with `chown -R` in Dockerfile
- [x] 3.2 Add `ENV PATH` with `~/bin`, `~/go/bin`, `/usr/local/go/bin`
- [x] 3.3 Add `INSTALL_GO` build arg with conditional Go toolchain installation
- [x] 3.4 Add `GOPATH` environment variable

## 4. Docker — Entrypoint Permission Verification

- [x] 4.1 Update `docker-entrypoint.sh` to create `skills/` and `~/bin` directories
- [x] 4.2 Add writability verification loop for critical directories
- [x] 4.3 Add actionable error message with `docker volume rm` hint on failure

## 5. Verification

- [x] 5.1 Run `go build ./...` and verify no compilation errors
- [x] 5.2 Run `go test ./internal/bootstrap/... ./internal/skill/...` and verify all tests pass
- [x] 5.3 Run `go test ./...` and verify full test suite passes
