## Why

`internal/sqlitedriver` and `internal/dbopen` are foundational production paths for opening and validating the SQLite store, but they currently rely on indirect coverage from other packages. That leaves package-local regressions around DSN construction, legacy-header rejection, and managed/read-only open flows easier to miss.

## What Changes

- add package-local regression tests for `internal/sqlitedriver`
- add package-local regression tests for `internal/dbopen`
- sync the main test-coverage spec for these low-level opening paths

## Impact

- stronger direct coverage around SQLite boot/open behavior
- earlier detection of legacy-header and read-only open regressions
- better confidence in production storage startup paths
