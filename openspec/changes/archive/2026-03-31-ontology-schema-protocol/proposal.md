## Why

The P2P protocol layer has no typed messages for ontology schema exchange. Peers cannot query each other's schema bundles or propose schema imports, which blocks the ontology-bridge (Change 3-5) from wiring schema exchange into the P2P stack. Adding typed protocol messages now enables future schema negotiation without import cycles.

## What Changes

- Add `RequestSchemaQuery` and `RequestSchemaPropose` request type constants to the P2P protocol
- Add typed request/response structs for schema exchange (`SchemaQueryRequest`, `SchemaQueryResponse`, `SchemaProposeRequest`, `SchemaProposeResponse`)
- Define `OntologyHandler` interface in the protocol package (uses `json.RawMessage` to avoid ontology import cycles)
- Wire `OntologyHandler` into the existing `Handler` struct with a setter and dispatch cases in `handleRequest`

## Capabilities

### New Capabilities
- `ontology-schema-protocol`: Typed P2P protocol messages and handler interface for schema query and schema propose exchanges between peers

### Modified Capabilities
<!-- No existing spec-level requirements are changing -->

## Impact

- **Code**: `internal/p2p/protocol/messages.go` (2 new constants), new file `internal/p2p/protocol/ontology_messages.go`, `internal/p2p/protocol/handler.go` (new field, setter, dispatch cases)
- **APIs**: New `OntologyHandler` interface that bridge packages will implement
- **Dependencies**: None — uses `json.RawMessage` to stay cycle-free
