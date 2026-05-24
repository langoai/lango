## Why

`internal/session.NewEntStore` still carried SQLCipher-era semantics in code comments and behavior by issuing `PRAGMA key` when `WithPassphrase(...)` was supplied, even though the current runtime no longer supports SQLCipher page encryption for session storage.

## What Changes

- remove the stale `PRAGMA key` path from `internal/session.NewEntStore`
- document `WithPassphrase(...)` as a deprecated compatibility-only option
- add a regression test proving plaintext session stores still open when the deprecated option is provided
- sync the main ent-session-store and test-coverage specs

## Impact

- aligns session-store behavior with the current plaintext/payload-protection runtime model
- removes dead SQLCipher unlock semantics from a production constructor
- keeps the deprecated compatibility path explicit and test-backed
