## Why

The current runtime no longer supports SQLCipher page-level encryption, but bootstrap/db-open comments still describe active `PRAGMA key` behavior and package-local tests do not directly pin the intended no-op semantics for deprecated encryption arguments on plaintext databases.

## What Changes

- update bootstrap/db-open comments to describe deprecated encryption arguments truthfully
- add package-local regression tests proving plaintext managed/read-only opens still succeed when those deprecated arguments are present
- sync the main db-encryption and test-coverage specs

## Impact

- reduces maintainer confusion around obsolete SQLCipher semantics
- makes deprecated compatibility behavior explicit and test-backed
- keeps DB-open behavior aligned with the current unsupported-legacy-DB model
