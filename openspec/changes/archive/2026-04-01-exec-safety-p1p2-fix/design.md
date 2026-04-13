## Context

Two issues found post-implementation of exec-safety-hardening:
- P1: `unwrapShellWrapper` takes everything after `-c` instead of just the command string
- P2: Catastrophic commands bypass PolicyEvaluator and reach approval prompt

## Goals / Non-Goals

**Goals:**
- Fix unwrap to follow POSIX `sh -c command_string [command_name [args...]]` — extract only `command_string`
- Add catastrophic pattern check as independent step in Evaluate, running for ALL commands before opaque detection
- Ensure ALL commands that SecurityFilterHook would block are also blocked by PolicyEvaluator (before approval)

**Non-Goals:**
- Shell parser library (`mvdan.cc/sh`) — deferred to Exec Phase 2
- Changing prompt/README/external surface — `ReasonCatastrophicPattern` is internal only

## Decisions

### D1: Unwrap extracts first argument only (POSIX semantics)

After finding `-c`, extract only the command string:
- **Quoted**: Find matching close quote, extract content between quotes
- **Unquoted**: Take first whitespace-delimited token
- **Unmatched quote**: Return `(original, false)` — allow-without-unwrap. Bash would reject this as syntax error, so it's not a real bypass vector.

### D2: Catastrophic pattern check as independent step 4

**Before opaque detection, for ALL commands.** Not an observe→block escalation.

```
Evaluate flow:
1. Unwrap → 2. Classify lango CLI → 3. CommandGuard → 4. Catastrophic patterns → 5. Opaque → 6. Allow
```

This covers:
- `rm -rf /` — non-opaque, caught at step 4
- `eval "rm -rf /"` — would be opaque at step 5, but caught at step 4 first
- `eval "echo hello"` — passes step 4, caught at step 5 as observe

### D3: Merged pattern set from SecurityFilterHook

Patterns injected via `WithCatastrophicPatterns(patterns)` functional option. Caller (app.go) passes `DefaultBlockedPatterns() + cfg.Hooks.BlockedCommands` with dedupe + lowercase — same merge semantics as `NewSecurityFilterHook`.

### D4: `ReasonCatastrophicPattern` as dedicated reason code

Catastrophic check is a separate concern from opaque detection. Using a dedicated reason code avoids confusion with opaque triggers.

## Risks / Trade-offs

- **[Dual pattern matching]** → Catastrophic patterns checked twice (PolicyEvaluator + SecurityFilterHook). Mitigation: defense-in-depth, string operations negligible.
- **[Unquoted unwrap only gets first token]** → `sh -c echo hello` returns `("echo", true)` — bash executes `echo` with `hello` as `$0`, not as argument. This is technically correct but may lose context. Mitigation: The first token is what matters for verb detection.
