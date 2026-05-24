## Why

`internal/session.NewEntStore` still performs Ent schema migration directly during constructor startup. After finding and fixing the same Atlas concurrency crash in `dbopen`, this constructor path remains another production-facing place where concurrent startup could reintroduce `concurrent map writes`.

## What Changes

- serialize `client.Schema.Create` inside `internal/session.NewEntStore`
- add a concurrent `NewEntStore` regression test
- sync the main test-coverage spec

## Impact

- removes another Atlas migration crash path from runtime-owned constructors
- keeps session-store startup behavior directly covered by package-local tests
- aligns the session store with the db-open safety fix
