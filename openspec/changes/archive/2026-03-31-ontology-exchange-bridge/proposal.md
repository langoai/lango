## Why

Changes 3-1 (Schema Codec), 3-3 (Protocol Messages), and 3-4 (P2P Fact Source) are implemented but disconnected. The P2P protocol handler has an `OntologyHandler` interface with no implementation. The ontology service has `ExportSchema`/`ImportSchema`/`AssertP2PFact` but nothing calls them from P2P. This change wires them together via `OntologyBridge`.

## What Changes

- Create `OntologyBridge` implementing `protocol.OntologyHandler` interface — bridges P2P handler to ontology service
- Add `OntologyExchangeConfig` to config (enabled, trust thresholds, import mode, max types)
- Add `P2PPermission` to `OntologyACLConfig` — default permission for `peer:` principals
- Modify `RoleBasedPolicy.Check` to handle `peer:` prefix principals using P2PPermission
- Wire bridge in `wiring_ontology.go` when both P2P and Ontology are enabled
- Add `SchemaExchangeEvent` to eventbus

## Capabilities

### New Capabilities
- `ontology-exchange-bridge`: P2P-to-ontology wiring layer with trust-gated schema exchange, peer principal ACL, configurable import modes.

### Modified Capabilities
- `ontology-acl`: RoleBasedPolicy gains `peer:` prefix handling via P2PPermission config.

## Impact

- `internal/p2p/ontologybridge/bridge.go` — NEW package + bridge implementation
- `internal/ontology/acl.go` — `peer:` prefix handling (3 lines)
- `internal/config/types_ontology.go` — OntologyExchangeConfig, P2PPermission field
- `internal/app/wiring_ontology.go` — bridge creation + handler injection
- `internal/eventbus/events.go` — SchemaExchangeEvent
