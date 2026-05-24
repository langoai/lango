## 1. Tests First

- [x] 1.1 Add failing `agent list` test for invalid configured user `AGENT.md`.
- [x] 1.2 Add failing `agent status` test for invalid configured user `AGENT.md`.

## 2. Implementation

- [x] 2.1 Surface user agent registry load errors from `agent list`.
- [x] 2.2 Surface embedded and user registry load errors from `agent status`.
- [x] 2.3 Preserve missing `agent.agentsDir` behavior.

## 3. Docs and Verification

- [x] 3.1 Update public agent CLI docs for invalid user agent file behavior.
- [x] 3.2 Run focused agent CLI tests and OpenSpec validation.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Review, sync, and archive the OpenSpec change.
