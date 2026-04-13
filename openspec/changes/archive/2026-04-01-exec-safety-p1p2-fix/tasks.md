## 1. P1: Fix Shell Unwrap Positional Args

- [x] 1.1 Rewrite `unwrapShellWrapper` in `internal/tools/exec/unwrap.go` to extract only the first argument after `-c` (quoted content or first unquoted token)
- [x] 1.2 Add test cases to `internal/tools/exec/unwrap_test.go`: positional args after quoted command, unquoted first-token-only, unmatched quote
- [x] 1.3 Update `internal/tools/exec/policy_test.go` for `bash -c "kill 1234" ignored` → VerdictBlock

## 2. P2: Catastrophic Pattern Check

- [x] 2.1 Add `ReasonCatastrophicPattern` to `internal/tools/exec/policy.go`
- [x] 2.2 Add `Option` type, `WithCatastrophicPatterns` functional option, update `NewPolicyEvaluator` signature
- [x] 2.3 Add `matchesCatastrophicPattern` function and integrate as step 4 in `Evaluate` (before opaque detection)
- [x] 2.4 Add catastrophic pattern test cases to `internal/tools/exec/policy_test.go`

## 3. Wiring and Integration

- [x] 3.1 Update `internal/app/app.go` Phase B to pass `WithCatastrophicPatterns(merged defaults + user config)` to `NewPolicyEvaluator`
- [x] 3.2 Create `internal/app/policy_integration_test.go` with chain-order regression test (catastrophic → approval mock not called)

## 4. Build and Verify

- [x] 4.1 Run `go build ./...` and fix any compilation errors
- [x] 4.2 Run `go test ./...` and fix any test failures

## 5. Downstream Artifact Audit

- [x] 5.1 Update `openspec/specs/exec-policy-evaluator/spec.md` with catastrophic pattern step and ReasonCatastrophicPattern
- [x] 5.2 Verify prompts/README/skills unchanged (internal reason code only, no external surface change)
