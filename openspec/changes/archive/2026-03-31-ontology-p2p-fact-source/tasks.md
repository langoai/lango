# Tasks: Trust-Weighted P2P Fact Source

## Implementation

- [x] 1. Add `"p2p_exchange": 1` to SourcePrecedence in types.go
- [x] 2. Create p2p_source.go with P2PFactInput, P2PConfidenceScale, assertP2PFact, filterVerifiedTriples
- [x] 3. Add AssertP2PFact + VerifyP2PFact to OntologyService interface in service.go
- [x] 4. Implement AssertP2PFact + VerifyP2PFact on ServiceImpl in service.go
- [x] 5. Add exclude_unverified parameter to buildFactsAt, buildGetEntity, buildQueryEntities in tools.go
- [x] 6. Create p2p_source_test.go with table-driven tests
- [x] 7. Build verification: `go build -tags fts5 ./internal/ontology/...`
- [x] 8. Test verification: `go test -tags fts5 ./internal/ontology/... -v -count=1`
