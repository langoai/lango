## 1. Tests First

- [x] 1.1 Add a failing archtest that forbids direct `bootstrap.Run` calls in production CLI packages outside `internal/cli/cliboot`.
- [x] 1.2 Add focused doctor/settings/onboard seam tests where needed to preserve command behavior while using shared loaders.
- [x] 1.3 Run focused tests and confirm the new archtest fails before implementation.

## 2. Implementation

- [x] 2.1 Change `doctor` to default to `cliboot.BootResult` through a no-argument test seam.
- [x] 2.2 Change `settings` to use a package-level `cliboot.BootResult` seam instead of direct `bootstrap.Run`.
- [x] 2.3 Change `onboard` to use a package-level `cliboot.BootResult` seam instead of direct `bootstrap.Run`.
- [x] 2.4 Remove now-unused direct bootstrap imports.

## 3. Verification

- [x] 3.1 Run focused tests for `internal/archtest`, `internal/cli/doctor`, `internal/cli/settings`, and `internal/cli/onboard`.
- [x] 3.2 Run `openspec validate centralize-cli-bootstrap-loaders --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Run subagent review.
- [x] 3.5 Sync and archive the OpenSpec change.
