## 1. ACL Peer Principal

- [x] 1.1 Add `P2PPermission string` to `OntologyACLConfig` in `internal/config/types_ontology.go`
- [x] 1.2 Modify `RoleBasedPolicy` to accept P2PPermission and handle `peer:` prefix principals in `internal/ontology/acl.go`
- [x] 1.3 Update ACL wiring in `internal/app/wiring_ontology.go` to pass P2PPermission

## 2. Config

- [x] 2.1 Add `OntologyExchangeConfig` to `internal/config/types_ontology.go`

## 3. Bridge

- [x] 3.1 Create `internal/p2p/ontologybridge/bridge.go` — OntologyBridge implementing protocol.OntologyHandler
- [x] 3.2 Create `internal/p2p/ontologybridge/bridge_test.go`

## 4. Events

- [x] 4.1 Add `SchemaExchangeEvent` to `internal/eventbus/events.go`

## 5. Wiring

- [x] 5.1 Wire bridge in `internal/app/wiring_ontology.go` (when P2P + Ontology + Exchange all enabled)

## 6. Verification

- [x] 6.1 Build and test: `go build -tags fts5 ./...` and `go test -tags fts5 ./... -count=1`
