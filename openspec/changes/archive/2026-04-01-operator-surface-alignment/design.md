## Context

The exec-safety-hardening batch added 6 runtime capabilities across `internal/observability`, `internal/agentrt`, `internal/provenance`, `internal/tools/exec`, `internal/bootstrap`, and `internal/appinit`. The code is merged, but the operator-facing documentation (`docs/features/` and `docs/cli/`) has not been updated. Operators cannot discover the new `/metrics/policy` endpoint, the `lango metrics policy` command, the recovery decision event flow, the config/hook provenance snapshot, or the startup instrumentation without documentation.

## Goals / Non-Goals

**Goals:**

- Document the `/metrics/policy` gateway endpoint and `lango metrics policy` CLI command accurately
- Document policy decision audit logging (the `PolicyDecisionEvent` -> audit recorder flow)
- Document `RecoveryDecisionEvent` and the exponential backoff / per-error-class retry limit behavior
- Document config fingerprint and hook registry snapshot in provenance checkpoints
- Update the gateway endpoint summary table in observability docs

**Non-Goals:**

- No Go code changes (this is a docs-only change)
- No documentation for resume-aware budget internals (session metadata keys are implementation detail)
- No documentation for exec phase 2 shell analysis internals (unwrap is transparent to operators)
- No documentation for startup instrumentation internals (PhaseTimingEntry/ModuleTimingEntry are logged, not user-facing)

## Decisions

1. **Policy metrics docs go into observability.md** -- The `/metrics/policy` endpoint is part of the observability gateway surface, so it belongs alongside the existing metrics/health/audit sections rather than in a separate file.

2. **Recovery events documented under new "Recovery Decision Events" section in observability.md** -- `RecoveryDecisionEvent` is an observability concern (event bus subscription + structured logging) rather than a provenance concern.

3. **Config/hook provenance documented in provenance.md** -- The config fingerprint and hook registry snapshot are recorded as checkpoint metadata at session start, which is squarely within provenance scope.

4. **`lango metrics policy` added to metrics.md** -- Follows the existing pattern where each `lango metrics <subcommand>` gets its own section in `docs/cli/metrics.md`.

5. **No new files created** -- All content fits naturally into existing doc files. This avoids nav/sidebar changes.

## Risks / Trade-offs

- [Risk] Documentation may drift from code if features change in future. -> Mitigation: Sections reference specific config keys and endpoint paths that are verified against code.
- [Risk] Recovery event documentation may over-expose internal retry logic. -> Mitigation: Document the observable interface (event name, fields) not the decision tree internals.
