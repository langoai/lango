## Why

The exec-safety-hardening change introduced shell wrapper unwrap and a PolicyEvaluator middleware, but two bypass/UX issues remain:

1. **P1 — Positional args bypass**: `bash -c "kill 1234" ignored` passes extra positional arguments after the command string. The unwrap takes everything after `-c` as the inner command, so `stripQuotes` fails and the kill verb goes undetected.
2. **P2 — Catastrophic commands hit approval**: Direct catastrophic commands (`rm -rf /`) and opaque-but-catastrophic commands (`eval "rm -rf /"`) still reach the approval prompt before being blocked by `SecurityFilterHook`, because PolicyEvaluator doesn't check catastrophic patterns.

## What Changes

- Fix `unwrapShellWrapper` to extract only the first argument after `-c` (quoted string or first unquoted token), matching POSIX `sh -c command_string [command_name [argument...]]` semantics
- Add a catastrophic pattern check as an independent step in `PolicyEvaluator.Evaluate()`, running for ALL commands before opaque detection. Uses the same merged pattern set as `SecurityFilterHook` (defaults + user-configured `BlockedCommands`)
- Add `ReasonCatastrophicPattern` reason code
- Convert `NewPolicyEvaluator` to accept functional options (`WithCatastrophicPatterns`)
- Add chain-order regression test verifying blocked commands don't reach the approval provider

## Capabilities

### New Capabilities

(none — all changes are to existing capabilities)

### Modified Capabilities

- `exec-policy-evaluator`: Add catastrophic pattern check step (step 4) before opaque detection (step 5). Add `ReasonCatastrophicPattern`. Fix shell wrapper unwrap to respect POSIX positional argument semantics.

## Impact

- **Code**: `internal/tools/exec/unwrap.go`, `internal/tools/exec/policy.go`, `internal/app/app.go`, test files
- **No new dependencies**: Reuses `toolchain.DefaultBlockedPatterns()` + config
- **No breaking changes**: External tool names, JSON shapes, CLI output unchanged
- **Downstream**: `openspec/specs/exec-policy-evaluator/spec.md` updated. Prompts/README/skills unchanged (internal reason code only).
