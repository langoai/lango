## 1. Authoritative Read

- [ ] 1.1 Switch run-state reads to RunLedger snapshots under the stage gate
- [ ] 1.2 Add tests that workflow/background/gateway read the same snapshot state

## 2. Resume Integration

- [ ] 2.1 Integrate resume candidate lookup with gateway session handling
- [ ] 2.2 Add explicit candidate selection and confirmation flow
- [ ] 2.3 Add tests for stale runs and multi-candidate selection

## 3. Command Context

- [ ] 3.1 Inject run summaries into agent command context
- [ ] 3.2 Keep full journal/state on-demand only
- [ ] 3.3 Add tests for context truncation resistance

## 4. Verification

- [ ] 4.1 Run `go build ./...`
- [ ] 4.2 Run gateway/agent/runledger tests
- [ ] 4.3 Run `go test ./...`
