## Why

P2P peers (`peer:` prefix principals) currently default to `PermWrite` in the ontology ACL system, allowing any remote peer to write arbitrary data to the local ontology without explicit trust establishment. This is a security risk for a system positioning itself as a trustworthy runtime. The default should follow the principle of least privilege.

## What Changes

- Change `NewRoleBasedPolicy()` default P2P permission from `PermWrite` to `PermRead`
- Update `DefaultConfig()` to set explicit `ontology.acl.p2pPermission: "read"` default
- Update TUI settings fallback from `"write"` to `"read"`
- Update config documentation comment to reflect new default
- Add test coverage for P2P peer permission behavior (default + override)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ontology-acl`: Default P2P peer permission changes from `PermWrite` to `PermRead`. Existing users can restore previous behavior via `ontology.acl.p2pPermission: "write"`.

## Impact

- **Code**: `internal/ontology/acl.go`, `internal/config/loader.go`, `internal/config/types_ontology.go`, `internal/cli/settings/forms_ontology.go`, `internal/ontology/acl_test.go`
- **Behavior**: Remote P2P peers will be read-only by default. Peers needing write access must be explicitly configured.
- **Migration**: Non-breaking for new installations. Existing users who rely on P2P write must add `ontology.acl.p2pPermission: "write"` to config.
