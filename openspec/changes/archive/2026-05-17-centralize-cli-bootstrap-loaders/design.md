## Context

Most CLI commands already receive `cliboot.BootResult` or `cliboot.Config` from `cmd/lango/main.go`. The remaining direct callers are interactive/setup commands and doctor:

- `internal/cli/doctor` keeps a `doctorBootstrapRun` seam with the `bootstrap.Run` signature.
- `internal/cli/settings` and `internal/cli/onboard` call `bootstrap.Run` directly inside their run paths.

The last storage broker change made `cliboot.BootResult` the centralized place for `StartStorageBroker: true`. Keeping direct bootstrap calls outside that package reintroduces the same drift risk for future lifecycle options.

## Decision

Use `cliboot.BootResult` as the default bootstrap loader for the remaining CLI packages:

- `doctorBootstrapRun` becomes a no-argument `func() (*bootstrap.Result, error)` defaulting to `cliboot.BootResult`.
- `settings` and `onboard` get package-level boot loader seams defaulting to `cliboot.BootResult`, mirroring their existing command-run seams.
- Production CLI packages outside `internal/cli/cliboot` are forbidden from calling `bootstrap.Run` directly by an archtest.

This preserves unit-test injection while making the production path depend on one lifecycle owner.

## Alternatives Considered

- Keep direct calls but add option tests per package: rejected because it duplicates every future bootstrap option and does not remove the architectural drift.
- Pass loader functions from `cmd/lango/main.go` into settings/onboard/doctor constructors: rejected for now because these commands already own test seams and this would add constructor churn without improving the invariant beyond the archtest.

## Test Strategy

- Add a failing architecture test that scans non-test production files under `internal/cli` for `bootstrap.Run(` and only allows `internal/cli/cliboot`.
- Update doctor tests to use the new no-argument seam and verify bootstrap failures still surface.
- Add focused settings/onboard tests that replace their boot loader seams and verify the loader is invoked after the interactive guard.
- Run focused package tests, strict OpenSpec validation, full build, full tests, and subagent review.
