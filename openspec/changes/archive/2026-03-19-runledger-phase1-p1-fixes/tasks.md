## 1. Step Proposal Authorization

- [x] 1.1 Update `run_propose_step_result` to load the snapshot before appending
  `EventStepResultProposed`
- [x] 1.2 Reject proposals when the step does not exist
- [x] 1.3 Reject proposals when the caller agent does not match `step.OwnerAgent`
- [x] 1.4 Reject proposals when the step is not in `in_progress`
- [x] 1.5 Add/adjust tests for valid owner, wrong owner, unknown step, and wrong state

## 2. Retry-Safe Workspace Lifecycle

- [x] 2.1 Update workspace preparation to use retry-safe branch/path naming
- [x] 2.2 Add best-effort branch cleanup to workspace teardown
- [x] 2.3 Add a test that prepares the same step workspace twice successfully

## 3. Phase-Gated Runtime Semantics

- [x] 3.1 Update `runLedgerModule` comments to state Phase 1 intentionally does not
  enable workspace wiring
- [x] 3.2 Update RunLedger README/docs to describe workspace isolation as readiness-only
  until the future activation phase
- [x] 3.3 Update the main RunLedger spec delta with the explicit phase-gated wording

## 4. Verification

- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run `go test ./internal/runledger/...`
- [x] 4.3 Run `go test ./...`
