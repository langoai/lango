## Tasks

- [x] 1. Add `RequestSchemaQuery` and `RequestSchemaPropose` constants to `internal/p2p/protocol/messages.go`
- [x] 2. Create `internal/p2p/protocol/ontology_messages.go` with typed request/response structs and `OntologyHandler` interface
- [x] 3. Add `ontologyHandler` field and `SetOntologyHandler` setter to `Handler` in `internal/p2p/protocol/handler.go`
- [x] 4. Add `schema_query` and `schema_propose` dispatch cases to `handleRequest` in `handler.go`
- [x] 5. Add `handleSchemaQuery` and `handleSchemaPropose` private methods to `handler.go`
- [x] 6. Build check: `go build -tags fts5 ./internal/p2p/protocol/...`
