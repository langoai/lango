## 1. Bridge Wiring (Finding 1+4+5)

- [x] 1.1 Add TrustScorer interface, SetReputation/SetEventBus setters to bridge
- [x] 1.2 Add event publish in HandleSchemaQuery/HandleSchemaPropose
- [x] 1.3 Add bridge to intelligenceValues in modules.go
- [x] 1.4 Add post-build wiring in app.go (SetOntologyHandler + SetReputation + SetEventBus)

## 2. Promote Fix (Finding 2)

- [x] 2.1 Add UpdateTypeStatus/UpdatePredicateStatus to Registry interface
- [x] 2.2 Implement in EntRegistry (Update().Where().SetStatus() pattern)
- [x] 2.3 Fix PromoteType/PromotePredicate to use UpdateStatus + refreshPredicateCache + version bump

## 3. Import Cache Fix (Finding 3)

- [x] 3.1 Add refreshPredicateCache + version.Add after importSchema in ImportSchema

## 4. Regression Tests

- [x] 4.1 PromoteType success path test
- [x] 4.2 PromotePredicate success + cache refresh test
- [x] 4.3 ImportSchema predicate immediately usable test
- [x] 4.4 ImportSchema version bumped test
- [x] 4.5 Bridge trust rejection/acceptance tests

## 5. Verification

- [x] 5.1 go build ./... + go build -tags fts5 ./...
- [x] 5.2 go test -tags fts5 ./internal/ontology/... + bridge + app
