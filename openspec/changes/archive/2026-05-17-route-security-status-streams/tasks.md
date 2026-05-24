## 1. Tests

- [x] 1.1 Add a failing command-level test for status warnings on Cobra stderr.
- [x] 1.2 Add a failing helper test proving status secure-provider detection is injected.
- [x] 1.3 Add failing helper tests proving broker startup is injected and failure degrades.

## 2. Implementation

- [x] 2.1 Thread command stderr through `runStatusNonInteractive` and `readDBStatusNonInteractive`.
- [x] 2.2 Replace direct status secure-provider detection with a package seam.
- [x] 2.3 Replace direct status broker startup with a package seam.
- [x] 2.4 Preserve status stdout rendering and graceful degradation behavior.

## 3. Verification

- [x] 3.1 Validate the OpenSpec change in strict mode.
- [x] 3.2 Run focused security status tests.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync and archive the OpenSpec change.
- [x] 3.5 Commit this scoped unit separately.
