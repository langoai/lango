## 1. Slim Wire Types

- [x] 1.1 Add `SchemaTypeSlim`, `SchemaPredicateSlim`, `SchemaPropertySlim` to `internal/ontology/types.go`
- [x] 1.2 Add `SchemaBundle`, `ImportMode` (shadow/governed/dry_run), `ImportOptions`, `ImportResult` to `internal/ontology/types.go`

## 2. Conversion and Digest

- [x] 2.1 Create `internal/ontology/exchange.go` — `TypeToSlim`, `PredicateToSlim`, `SlimToType`, `SlimToPredicate` converters
- [x] 2.2 Add `ComputeDigest` function (canonical JSON SHA256, order-independent)

## 3. Service Methods

- [x] 3.1 Add `ExportSchema` and `ImportSchema` to `OntologyService` interface and `ServiceImpl`
- [x] 3.2 Implement `ExportSchema` — filter active+shadow, convert to slim, compute digest
- [x] 3.3 Implement `ImportSchema` — slim→full conversion, governance-aware status, conflict detection, dry_run

## 4. Tests

- [x] 4.1 Create `internal/ontology/exchange_test.go` — roundtrip, conflict, governance, dry_run, digest stability, order independence

## 5. Verification

- [x] 5.1 Build and test: `go build -tags fts5 ./...` and `go test -tags fts5 ./internal/ontology/... -v`
