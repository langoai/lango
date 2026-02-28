## 1. Config Defaults

- [x] 1.1 Add ObservationalMemory defaults to DefaultConfig() in config/loader.go
- [x] 1.2 Register ObservationalMemory viper defaults in Load() function

## 2. Doctor Check Fixes

- [x] 2.1 Downgrade graph store databasePath empty check from StatusFail to StatusWarn in graph_store.go
- [x] 2.2 Fix database.go resolveDatabasePath fallback from sessions.db to lango.db

## 3. Test & Spec Updates

- [x] 3.1 Update checks_test.go database path reference from sessions.db to lango.db
- [x] 3.2 Update openspec/specs/server/spec.md sessions.db reference to lango.db
