## 1. Implementation

- [x] 1.1 Add `os`, `path/filepath`, `strings` imports to `internal/graph/bolt_store.go`
- [x] 1.2 Add tilde expansion (`~/` → `os.UserHomeDir()`) before `bolt.Open()` in `NewBoltStore()`
- [x] 1.3 Add `os.MkdirAll(filepath.Dir(path), 0o700)` to create parent directories before `bolt.Open()`

## 2. Verification

- [x] 2.1 Run `go build ./...` to verify compilation
- [x] 2.2 Run `go test ./...` to verify existing tests pass
- [x] 2.3 Run `go vet ./...` to check for issues
