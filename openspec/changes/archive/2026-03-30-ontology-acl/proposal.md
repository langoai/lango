## Why

OntologyService exposes 29 methods (schema queries, mutations, truth maintenance, entity resolution, property store) with no access control. Any agent or programmatic caller can perform any operation — including destructive ones like `MergeEntities` or `SplitEntity`. Stage 2 requires operation-level ACL so that sub-agents (e.g., chronicler=read-only, ontologist=admin) are restricted to their intended permission scope.

## What Changes

- Add `Permission` type (`read < write < admin`) and `ACLPolicy` interface to the ontology subsystem
- Implement `AllowAllPolicy` (default, backward compatible) and `RoleBasedPolicy` (principal→permission mapping)
- Add `checkPermission` guard to all 28 public methods on `ServiceImpl` (PredicateValidator excluded — no ctx parameter)
- Add `WithPrincipal`/`PrincipalFromContext` context keys in `ctxkeys` package
- Add `WithPrincipal()` middleware in `toolchain` to map agent name → ontology principal at tool execution boundary
- Add `OntologyACLConfig` to config with `ontology.acl.enabled` and `ontology.acl.roles` fields
- Wire ACL policy creation in `wiring_ontology.go`

## Capabilities

### New Capabilities
- `ontology-acl`: Operation-level access control for ontology service methods. Defines Permission model, ACLPolicy interface, RoleBasedPolicy implementation, principal context propagation, and service-layer permission guards.

### Modified Capabilities
- `ontology-tools`: Tool handlers pass ctx unchanged; ACL is enforced inside service layer. No tool-level changes, but ontologist agent identity doc updated to mention ACL.

## Impact

- `internal/ontology/` — new `acl.go`, modified `service.go` (28 method guards), `types.go` (Permission type)
- `internal/ctxkeys/` — new principal context key pair
- `internal/toolchain/` — new `mw_principal.go` middleware
- `internal/app/app.go` — middleware chain insertion (B4c2)
- `internal/app/wiring_ontology.go` — ACL policy wiring
- `internal/config/types_ontology.go` — ACLConfig struct
- `prompts/agents/ontologist/IDENTITY.md` — ACL description
- No breaking changes: `acl == nil` preserves existing allow-all behavior
