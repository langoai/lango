## 1. Context Key and Middleware

- [x] 1.1 Add `WithPrincipal`/`PrincipalFromContext` to `internal/ctxkeys/ctxkeys.go`
- [x] 1.2 Create `internal/toolchain/mw_principal.go` — WithPrincipal middleware (agent name → principal)
- [x] 1.3 Insert `toolchain.WithPrincipal()` in middleware chain at B4c2 in `internal/app/app.go`

## 2. Permission Types

- [x] 2.1 Add `Permission` type (int, PermRead/PermWrite/PermAdmin) and `ErrPermissionDenied` to `internal/ontology/types.go`

## 3. ACL Policy Implementation

- [x] 3.1 Create `internal/ontology/acl.go` — ACLPolicy interface, AllowAllPolicy, RoleBasedPolicy
- [x] 3.2 Create `internal/ontology/acl_test.go` — tests for AllowAll, RoleBasedPolicy (grant, deny, system, unknown)

## 4. Service Integration

- [x] 4.1 Add `acl` field and `SetACLPolicy` setter to `ServiceImpl` in `internal/ontology/service.go`
- [x] 4.2 Add `checkPermission` private helper to `ServiceImpl`
- [x] 4.3 Add `checkPermission` guard to all 28 public methods (PredicateValidator excluded)
- [x] 4.4 Add service-level ACL integration tests to `acl_test.go`

## 5. Config and Wiring

- [x] 5.1 Add `OntologyACLConfig` to `internal/config/types_ontology.go`
- [x] 5.2 Add ACL policy creation and injection in `internal/app/wiring_ontology.go`

## 6. Downstream

- [x] 6.1 Update `prompts/agents/ontologist/IDENTITY.md` with ACL description
- [x] 6.2 Build and test: `go build -tags fts5 ./...` and `go test -tags fts5 ./internal/ontology/... -v`
