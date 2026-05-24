## 1. Tests First

- [x] 1.1 Add failing ProvidersCheck test for `agent.provider` set with an empty providers map.

## 2. Implementation

- [x] 2.1 Fail when `agent.provider` has no matching providers map entry, even if the map is empty.
- [x] 2.2 Preserve the legacy `GOOGLE_API_KEY` fallback only when no provider config exists at all.

## 3. Verification

- [x] 3.1 Run focused doctor checks tests and OpenSpec validation.
- [x] 3.2 Run `go build ./...` and `go test ./...`.
- [x] 3.3 Run subagent review.
- [x] 3.4 Sync and archive the OpenSpec change.
