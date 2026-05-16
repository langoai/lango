## Why

After fixing concurrent Ent Atlas migration crashes in `dbopen` and `session`, the codebase still needs a guard that prevents new unsynchronized production `Schema.Create` call sites from appearing silently elsewhere.

## What Changes

- add an executable production-quality guard that restricts non-test `Schema.Create` usage to the approved serialized constructors
- require those approved constructors to keep their `schemaCreateMu` locking in place
- sync the main test-coverage spec

## Impact

- prevents future reintroduction of unsynchronized Ent schema-migration paths
- keeps migration serialization policy explicit and enforceable
- narrows a subtle crash class at the repository level
