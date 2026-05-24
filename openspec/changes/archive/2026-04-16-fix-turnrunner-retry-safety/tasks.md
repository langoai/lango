# Tasks: fix-turnrunner-retry-safety

- [x] Fix 4: Add retryLoop label and break on parent.Done()
- [x] Fix 4: Override result with context_cancelled after loop
- [x] Fix 5: Add causeClass/attempt/backoffMs to recordRecovery payload
- [x] Verify: go test ./internal/turnrunner/... — ALL PASS
