## Context

Codex review exposed that several P2P security features were implemented but not correctly wired at the application layer. The controls existed in isolation but were bypassed at runtime due to context key mismatches, missing initialization calls, and incorrect execution ordering.

## Goals / Non-Goals

**Goals:**
- Ensure every P2P security control is enforced at runtime (no dead code paths)
- Fix all 9 issues identified across two Codex review rounds
- Maintain backward compatibility for existing configurations

**Non-Goals:**
- Adding new security features beyond what was already designed
- Changing the P2P protocol or handshake flow
- Modifying the sandbox runtime probe chain

## Decisions

1. **Unify P2P context marker** — Remove local `ctxKeyP2P` from filesystem package, use `ctxkeys.IsP2PRequest()` everywhere. This is the canonical P2P context key set by the handler. Alternative: bridge both keys → rejected because it creates two sources of truth.

2. **DNS resolution in URL validator** — Call `net.LookupIP()` for non-IP hostnames and check all resolved IPs against private ranges. If DNS fails, allow the request (browser will fail on its own). Alternative: block on DNS failure → rejected because it breaks legitimate URLs with temporary DNS issues.

3. **Post-navigation re-validation** — Add `CurrentURL()` method to browser tool, re-validate after Navigate in P2P context. If redirect leads to blocked URL, navigate to `about:blank` and return error. Alternative: intercept redirects at browser level → rejected because rod doesn't expose redirect hooks.

4. **Safety gate before payment gate** — In `handleToolInvokePaid`, check safety level before payment processing. This prevents charging for tools that will be denied.

5. **requireContainer fail-closed** — When `NewContainerExecutor` fails and `RequireContainer` is true, leave sandbox executor nil instead of falling back to subprocess. The handler's existing nil-check will reject P2P tool calls.

6. **ParseSafetyLevel fallback** — Check the boolean return from `ParseSafetyLevel`. On invalid/empty values, default to `SafetyLevelModerate` instead of silently accepting `Dangerous`.

## Risks / Trade-offs

- **[Risk] DNS resolution adds latency to P2P browser requests** → Acceptable: only affects P2P context, and DNS is typically cached.
- **[Risk] Post-navigation re-validation has a TOCTOU window** → Mitigated: the window is very small (between Navigate and Eval), and this is defense-in-depth on top of the initial check.
- **[Trade-off] requireContainer=true with no Docker → all P2P tools blocked** → Intentional: this is the documented fail-closed behavior operators opted into.
