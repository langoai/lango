## Why

The policy evaluator detects shell wrappers and opaque patterns, but does not handle several common shell constructs: heredocs, process substitution, grouped subshells, shell function definitions, `xargs cmd`, and `find -exec cmd`. Additionally, `VAR=val cmd` (env prefix without explicit `env`) is not unwrapped. These gaps allow policy bypass or miss monitoring opportunities for potentially dangerous commands hidden inside these constructs.

## What Changes

- Detect heredocs, process substitution, grouped subshells, and shell function definitions as opaque patterns (observe verdict)
- Extract inner verb from `xargs cmd` and `find -exec cmd` constructs, evaluate with existing guard; fallback to observe if extraction fails
- Strip `VAR=val` env prefix from commands to evaluate the effective verb
- Integrate all new detections into the Evaluate flow

## Capabilities

### New Capabilities
- `opaque.go`: Detection of heredoc, process substitution, grouped subshell, shell function patterns
- `unwrap.go`: Env prefix stripping (`VAR=val cmd`), xargs/find-exec inner verb extraction

### Modified Capabilities
- `exec-policy-evaluator`: Integrate new construct handling in Evaluate flow; add new ReasonCodes for shell constructs

## Impact

- `internal/tools/exec/opaque.go` — add detection for heredoc, process substitution, grouped subshell, shell function, xargs, find-exec
- `internal/tools/exec/unwrap.go` — add env prefix handling (strip VAR=val), xargs/find-exec inner verb extraction
- `internal/tools/exec/policy.go` — integrate new construct handling in Evaluate flow
- `internal/tools/exec/policy_test.go` — add tests for each construct
- `internal/tools/exec/opaque_test.go` — add tests for new opaque patterns
- `internal/tools/exec/unwrap_test.go` — add tests for env prefix, xargs, find-exec extraction
- `prompts/SAFETY.md` — update shell wrapper section with new constructs
