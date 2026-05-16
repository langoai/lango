## Why

Recent cleanup consolidated many CLI pretty-JSON writers into shared helpers, but nothing prevented future regressions from reintroducing package-local `SetIndent("", "  ")` blocks across `internal/cli`.

## What Changes

- add an executable repository guard that rejects duplicate pretty-JSON writer setup in CLI production code
- allow the shared `internal/cli/clihttp` helper to remain the single indentation setup point

## Impact

- cheaper detection of CLI JSON-formatting regressions
- stronger consistency around shared CLI JSON writer usage
