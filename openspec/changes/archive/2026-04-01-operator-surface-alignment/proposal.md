## Why

The exec-safety-hardening batch delivered 6 new runtime capabilities (policy observability, config/hook provenance, resume-aware budget, recovery hardening, AST-based shell analysis, startup instrumentation), but the operator-facing documentation has not been updated to reflect them. Operators cannot discover, configure, or troubleshoot these features without accurate docs and CLI reference.

## What Changes

- Add policy metrics section to observability docs (`/metrics/policy` endpoint, `lango metrics policy` CLI, audit logging for policy decisions)
- Add recovery decision event documentation to observability docs (`RecoveryDecisionEvent`, exponential backoff, per-error-class retry limits)
- Add config/hook provenance snapshot section to provenance docs (config fingerprint + hook registry snapshot at session start)
- Add `lango metrics policy` command reference to CLI metrics docs
- Update gateway endpoint tables to include `/metrics/policy`

## Capabilities

### New Capabilities

_(none -- this change is docs-only and introduces no new code capabilities)_

### Modified Capabilities

- `observability`: Adding documentation for policy metrics endpoint, policy decision audit logging, and recovery decision events
- `session-provenance`: Adding documentation for config fingerprint and hook registry snapshot at session start

## Impact

- `docs/features/observability.md` -- new sections for policy metrics and recovery events
- `docs/features/provenance.md` -- new section for config/hook provenance snapshots
- `docs/cli/metrics.md` -- new `lango metrics policy` command reference
- No Go code changes, no API changes, no dependency changes
