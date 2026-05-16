## Why

After extracting `internal/testutil/schemautil`, test code can still regress by reintroducing direct `Schema.Create` calls in individual test packages. That would bypass the serialized helper, recreate import-cycle pressure, and risk reintroducing Atlas-related flake paths.

## What Changes

- add an executable guard that blocks direct `Schema.Create` usage in test code outside the approved helper boundary
- sync the main test-coverage spec

## Impact

- keeps test schema bootstrap centralized
- reduces future concurrency flakes and import-cycle regressions in tests
- preserves the small helper boundary we just introduced
