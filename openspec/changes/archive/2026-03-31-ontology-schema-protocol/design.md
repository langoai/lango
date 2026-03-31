## Context

The P2P protocol layer (`internal/p2p/protocol/`) uses a stream-based request/response pattern over libp2p. Each request type has typed payload structs and is dispatched in `Handler.handleRequest`. Team and negotiation messages already follow a well-established pattern: constants in a `*_messages.go` file, handler interface/func type in the package, setter on Handler, and switch-case dispatch.

The ontology subsystem (`internal/ontology/`) defines `SchemaBundle` and `ImportResult` types. The protocol package cannot import ontology without creating cycles, since ontology may need protocol types in the future (Change 3-5 bridge).

## Goals / Non-Goals

**Goals:**
- Add typed `schema_query` and `schema_propose` request types to the P2P protocol
- Define request/response structs using `json.RawMessage` for bundle payloads (cycle-free)
- Define `OntologyHandler` interface with typed method signatures
- Wire the handler into `Handler` struct following existing setter/dispatch pattern

**Non-Goals:**
- Implementing the `OntologyHandler` (that's Change 3-5: bridge package)
- Adding CLI/TUI commands for schema exchange
- Modifying any code outside `internal/p2p/protocol/`

## Decisions

1. **`json.RawMessage` for bundle fields** — Avoids importing `internal/ontology` package. The bridge layer (Change 3-5) will handle encoding/decoding between `SchemaBundle`/`ImportResult` and raw JSON. Alternative: `interface{}` — rejected because it loses compile-time documentation of the expected shape.

2. **Interface type for OntologyHandler** (not func type) — Schema exchange has two related methods (`HandleSchemaQuery`, `HandleSchemaPropose`). An interface groups them naturally, unlike negotiate which uses a single func. Matches the `SecurityEventTracker` pattern already in handler.go.

3. **New file `ontology_messages.go`** — Follows the `team_messages.go` convention of grouping domain-specific constants and payload types in a dedicated file. Keeps `messages.go` focused on core types.

4. **Dispatch directly in handleRequest** (not a sub-router) — Only 2 request types, so a `TeamRouter`-style sub-router adds unnecessary indirection. The handler method decodes JSON and delegates to the interface.

## Risks / Trade-offs

- **[Risk] json.RawMessage requires callers to know the expected schema** → Mitigated by clear doc comments naming the expected types (SchemaBundle, ImportResult)
- **[Risk] OntologyHandler nil at runtime** → Mitigated by nil check with descriptive error, same pattern as negotiator/teamHandler
