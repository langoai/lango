## Why

The architecture data-flow page still referenced a deleted `internal/librarian/buffer.go` path even though the current implementation lives in `internal/librarian/proactive_buffer.go`. That kind of broken code-path reference erodes trust in the docs and slows down contributors who follow the architecture map.

## What Changes

- sync the architecture data-flow page to the current librarian proactive buffer path
- add an executable docs guard so the stale broken path cannot silently return

## Impact

- architecture docs better match the current code layout
- less contributor confusion when tracing the proactive librarian pipeline
- stronger regression protection for broken public code-path references
