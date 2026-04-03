## 1. P2P Browser DNS Rebinding Fix

- [x] 1.1 Remove `finalURL != rawURL` condition from post-navigation re-validation in `internal/tools/browser/tools.go`

## 2. P2P Sandbox Executor Gate

- [x] 2.1 Restore `if cfg.P2P.ToolIsolation.Enabled` gate in `internal/app/app.go`
- [x] 2.2 Add startup warning log in `else` branch when P2P enabled but toolIsolation disabled

## 3. Filesystem Delete Symlink Safety

- [x] 3.1 Rewrite `Delete` to check `os.Lstat` before `validatePath` for symlink detection
- [x] 3.2 For symlinks: resolve parent dir only via `EvalSymlinks`, build canonical link path
- [x] 3.3 Validate canonical link location via `checkPathAccess` before deletion
- [x] 3.4 Delete symlink itself (`os.Remove` on unresolved path), not resolved target
- [x] 3.5 For non-symlinks: use standard `validatePath` flow (no behavior change)

## 4. checkPathAccess Helper

- [x] 4.1 Extract blocked/allowed checks from `validatePath` into `checkPathAccess(absPath, origPath)` helper
- [x] 4.2 Compare input path against both unresolved and resolved versions of each config entry
- [x] 4.3 Update `validatePath` to call `checkPathAccess`

## 5. Verification

- [x] 5.1 `go build ./...` succeeds
- [x] 5.2 `go test ./internal/tools/filesystem/... ./internal/tools/browser/... ./internal/app/...` passes
