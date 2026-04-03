## Why

`DefaultBlockedPatterns()` in `hook_security.go` contained only 9 hardcoded patterns. Missing: privilege escalation (`sudo`, `su -`, `chmod +s`), remote code execution pipelines (`curl|sh`, `wget|bash`), reverse shell tools (`nc -l`, `ncat`, `socat`), and additional destructive operations. No observe-level patterns existed for legitimate but commonly-abused interpreter invocations.

## What Changes

- Expand blocked patterns from 9 to 20+ entries organized by category (privilege escalation, remote code execution, reverse shells, block device writes, mass deletion)
- Add `ObservePatterns` field and `DefaultObservePatterns()` for interpreter invocations (`python -c`, `perl -e`, `node -e`, `ruby -e`)
- Add compound pattern support — multi-part patterns (e.g., `curl` + `| sh`) pre-computed at construction time via `compoundPattern` struct
- Extract shared `matchPattern()` helper to eliminate code duplication between block and observe paths
- Wire observe patterns through hook middleware alongside existing block results

## Capabilities

### New Capabilities

- `exec-observe-patterns`: Observe-level logging for interpreter invocations that are legitimate but common obfuscation vectors
- `exec-compound-patterns`: Multi-part pattern matching for command pipelines (e.g., `curl ... | sh`)

### Modified Capabilities

- `exec-safety`: Expanded blocked pattern set covering privilege escalation, remote code execution, reverse shells, and block device writes

## Impact

- `internal/toolchain/hook_security.go` — pattern expansion, compound pattern struct, matchPattern helper
- `internal/toolchain/hook_security_test.go` — tests for all new pattern categories
- `internal/toolchain/hook_registry.go` — observe pattern propagation
- `internal/toolchain/hooks.go` — ObservePatterns field
- `internal/toolchain/hooks_test.go` — observe pattern tests
- `internal/toolchain/mw_hooks.go` — middleware observe result handling
- `internal/toolchain/mw_hooks_test.go` — middleware integration tests
