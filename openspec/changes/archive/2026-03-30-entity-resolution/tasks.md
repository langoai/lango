## Tasks

- [x] Create EntityResolver interface and MergeResult type in resolution.go
- [x] Create EntityAlias Ent schema
- [x] Run Ent codegen
- [x] Implement AliasStore (Ent-backed CRUD) in resolution_ent.go
- [x] Implement entityResolver — Resolve, RegisterAlias, Aliases delegation
- [x] Implement entityResolver — DeclareSameAs with second-arg-canonical convention
- [x] Implement entityResolver — Merge with safe ordering (snapshot → replicate → retract → alias)
- [x] Implement entityResolver — Split (alias removal)
- [x] Add 6 entity resolution methods to OntologyService interface
- [x] Add ServiceImpl delegation methods with nil-check guards
- [x] Add SetEntityResolver setter on ServiceImpl
- [x] Upgrade StoreTriple with Resolve pipeline (write-path canonicalization)
- [x] Add QueryTriples method (read-path canonicalization)
- [x] Wire EntityResolver in wiring_ontology.go (AliasStore + NewEntityResolver + inject)
- [x] Write 11 tests covering Resolve, DeclareSameAs, Merge, Split, Aliases, StoreTriple, QueryTriples
- [x] Verify build and regression (ontology, graph, learning, memory packages)
