## 1. Authoritative Read

- [x] 1.1 Switch run-state reads to RunLedger snapshots under the stage gate
- [x] 1.2 Add tests that workflow/background/gateway read the same snapshot state

## 2. Resume Integration

- [x] 2.1 Integrate resume candidate lookup with gateway session handling
- [x] 2.2 Add explicit candidate selection and confirmation flow
- [x] 2.3 Add tests for stale runs and multi-candidate selection

## 3. Command Context

- [x] 3.1 Inject run summaries into agent command context
- [x] 3.2 Keep full journal/state on-demand only
- [x] 3.3 Add tests for context truncation resistance

## 4. Verification

- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run gateway/agent/runledger tests
- [x] 4.3 Run `go test ./...`
