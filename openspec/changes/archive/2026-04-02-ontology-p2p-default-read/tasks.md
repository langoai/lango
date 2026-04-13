## 1. Core Default Change

- [x] 1.1 Change `NewRoleBasedPolicy()` in `internal/ontology/acl.go` to use `PermRead` instead of `PermWrite`
- [x] 1.2 Update comment in `internal/config/types_ontology.go` from `(default "write")` to `(default "read")`

## 2. Config & UI Defaults

- [x] 2.1 Add `Ontology.ACL.P2PPermission: "read"` to `DefaultConfig()` in `internal/config/loader.go`
- [x] 2.2 Change TUI fallback in `internal/cli/settings/forms_ontology.go` from `"write"` to `"read"`

## 3. Test Coverage

- [x] 3.1 Add `TestRoleBasedPolicy_P2PPeerDefaultRead` verifying peer principals default to read-only
- [x] 3.2 Add `TestRoleBasedPolicy_P2PPeerOverride` verifying `SetP2PPermission(PermWrite)` restores write access

## 4. Verification

- [x] 4.1 Run `go build ./...` — passes
- [x] 4.2 Run `go test ./internal/ontology/... ./internal/config/... ./internal/cli/settings/...` — all pass
