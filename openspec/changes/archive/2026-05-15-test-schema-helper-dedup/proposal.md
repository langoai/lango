## Why

After introducing migration-serialization guards, several tests still duplicated direct `Schema.Create` setup logic. Reusing the full `internal/testutil` package from those tests created import cycles, so the suite needed a smaller shared helper boundary.

## What Changes

- add a minimal `internal/testutil/schemautil` helper for serialized Ent schema creation in tests
- switch direct test call sites in configstore, payment, security, and shared test helpers to that subpackage
- sync the main test-coverage spec for this shared helper boundary

## Impact

- removes import-cycle pressure from test-only schema setup
- keeps Atlas migration serialization reusable across tests
- reduces duplicated low-level schema bootstrap logic
