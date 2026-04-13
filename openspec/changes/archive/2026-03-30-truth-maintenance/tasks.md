## Tasks

- [x] Add temporal metadata constants and SourcePrecedence map to types.go
- [x] Create TruthMaintainer interface and supporting types in truth.go
- [x] Create OntologyConflict Ent schema
- [x] Run Ent codegen
- [x] Implement ConflictStore (Ent-backed CRUD) in truth_ent.go
- [x] Implement truthMaintainer — AssertFact with ValidateTriple-first, metadata-before-conflict order
- [x] Implement truthMaintainer — RetractFact (soft delete via ValidTo)
- [x] Implement truthMaintainer — FactsAt with ValidFrom+ValidTo window check
- [x] Implement truthMaintainer — ConflictSet, ResolveConflict, OpenConflicts
- [x] Implement isCurrentlyValid/isValidAt with both ValidFrom and ValidTo checks
- [x] Implement canAutoResolve with SourcePrecedence comparison
- [x] Implement auto-resolve audit trail (ConflictAutoResolved record)
- [x] Add 6 truth maintenance methods to OntologyService interface
- [x] Add ServiceImpl delegation methods with nil-check guards
- [x] Add SetTruthMaintainer setter on ServiceImpl
- [x] Wire TruthMaintainer in wiring_ontology.go (ConflictStore + NewTruthMaintainer + inject)
- [x] Write 17 tests covering all scenarios (conflict, auto-resolve, retract, FactsAt, backward compat)
- [x] Verify build and regression (graph, learning, memory packages)
