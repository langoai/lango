# Design

## Root Cause

`ProvidersCheck.Run` only reports a missing `agent.provider` reference when `len(cfg.Providers) > 0`. If `agent.provider` is non-empty and `cfg.Providers` is empty, the missing reference branch is skipped, the legacy fallback is skipped, and the check returns StatusPass with an empty configured provider list.

## Approach

- Remove the `len(cfg.Providers) > 0` guard from the `agent.provider` reference check.
- Keep the legacy environment fallback gated on both `len(cfg.Providers) == 0` and `cfg.Agent.Provider == ""`.
- Test the exact misconfiguration: `agent.provider` set, no providers map entry.

## Non-Goals

- Do not validate model availability or network connectivity.
- Do not modify onboarding or provider runtime resolution.
- Do not change CLI output rendering.
