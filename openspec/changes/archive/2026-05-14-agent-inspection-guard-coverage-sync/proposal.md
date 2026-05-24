## Why

The generic migrated-family code guard intentionally excludes the whole agent CLI tree because trace and metrics commands still use legacy `--json` paths. After migrating the inspection subset (`status`, `list`, `tools`, `hooks`), that subset still lacked code-level regression protection.

## What Changes

- add a focused executable guard for the migrated agent inspection subset
- record that coverage explicitly in the main test-coverage spec

## Impact

- keeps the migrated agent inspection subset under code-level anti-regression coverage
- avoids overreaching into trace/metrics paths that have not migrated yet
