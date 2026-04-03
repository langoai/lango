## Context

The exec safety layer's blocked pattern list was minimal (9 entries) and lacked coverage for common attack vectors. The matching logic also allocated a new pattern slice on every `Pre()` invocation and had duplicated code between block and observe paths.

## Goals / Non-Goals

**Goals:**
- Cover all common privilege escalation, RCE pipeline, and reverse shell patterns
- Add observe-level patterns for ambiguous commands (legitimate but risky)
- Pre-compute compound patterns at construction time (zero per-call allocation on hot path)
- Single shared matching logic for both block and observe

**Non-Goals:**
- Regex-based pattern matching (stays with `strings.Contains` for simplicity)
- Shell parsing or deobfuscation (handled separately by `PolicyEvaluator`)
- Per-user or per-agent pattern overrides

## Decisions

1. **Category-organized patterns** — Group patterns by attack type (privilege escalation, RCE, reverse shell, block device, mass deletion) with inline comments. Rationale: readable, auditable, easy to extend.

2. **Compound patterns as struct** — `compoundPattern{parts []string}` requires ALL parts present. Rationale: `curl` alone is legitimate; only `curl` + `| sh` together is dangerous.

3. **Pre-computed lowered slices** — Both `blockedLowered` and compound patterns computed once in constructor, stored on struct. Rationale: `Pre()` is called on every tool invocation (hot path).

4. **Observe vs Block separation** — Observe patterns log but don't block. Rationale: `python -c "print('hello')"` is legitimate; blocking would break real workflows.

## Risks / Trade-offs

- [String matching is bypassable via encoding or variable expansion] → Accepted; `PolicyEvaluator` handles shell unwrapping separately. This layer is defense-in-depth.
- [Compound patterns with 2 parts may false-positive on unrelated commands containing both substrings] → Mitigated by choosing distinctive parts (e.g., `| sh` not just `sh`).
