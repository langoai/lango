## 1. OpenSpec

- [x] 1.1 Add delta spec updates for runtime `SessionIsolation` behavior
- [x] 1.2 Capture summary-merge and discard rules in proposal/design

## 2. Runtime Isolation

- [x] 2.1 Extend the ADK session service to honor an isolated-agent set
- [x] 2.2 Route isolated sub-agent events into child session history instead of parent history
- [x] 2.3 Merge successful isolated child sessions back to parent using root-authored summaries
- [x] 2.4 Discard failed or rejected isolated child sessions

## 3. Defaults + Docs

- [x] 3.1 Mark built-in specialist agents with `session_isolation: true`
- [x] 3.2 Update docs/spec text so `SessionIsolation` is documented as runtime behavior

## 4. Verification

- [x] 4.1 Add or update tests for isolated/non-isolated routing, summary merge, and discard behavior
- [x] 4.2 Run `go build ./...`
- [x] 4.3 Run `go test ./...`
- [x] 4.4 Validate and archive the change
