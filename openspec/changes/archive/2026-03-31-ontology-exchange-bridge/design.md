## Context

Changes 3-1 through 3-4 are complete. OntologyService now has 39 methods (ExportSchema, ImportSchema, AssertP2PFact, VerifyP2PFact). P2P protocol handler has OntologyHandler interface (HandleSchemaQuery, HandleSchemaPropose) with no implementation. This change connects them.

## Goals / Non-Goals

**Goals:**
- End-to-end schema query/propose via P2P
- Trust-gated exchange (reputation-based)
- Peer principal ACL integration
- Conservative defaults (shadow mode, high trust thresholds)

**Non-Goals:**
- Gossip-based auto-sync (request-response only)
- Fact exchange via P2P (future — protocol messages not yet defined for facts)
- Automatic promotion of imported schemas

## Decisions

### D1: Bridge in separate package

`internal/p2p/ontologybridge/` — imports both `ontology` and `p2p/protocol`. Neither imports it. Clean dependency graph.

### D2: P2PPermission config instead of pattern matching

`OntologyACLConfig.P2PPermission` (string, default "write"). `RoleBasedPolicy.Check` detects `peer:` prefix → uses this permission level. No regex/glob pattern matching needed.

### D3: Bridge sets `peer:<did>` principal

`ctxkeys.WithPrincipal(ctx, "peer:"+peerDID)` before calling ontology service methods. Audit trail shows which peer triggered the operation.

### D4: json.RawMessage for cross-package serialization

Protocol uses `json.RawMessage` for Bundle/Result fields. Bridge marshals/unmarshals between `ontology.SchemaBundle`↔`json.RawMessage` and `ontology.ImportResult`↔`json.RawMessage`.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Import cycle | Bridge is standalone package, callback pattern |
| Untrusted peer floods proposals | MaxTypesPerImport + trust threshold + governance rate limit |
| ACL blocks all P2P | P2PPermission default=PermWrite, configurable |
