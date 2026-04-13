## Why

When running golangci-lint v2.4.0 in CI, 90 issues were reported (errcheck:50, staticcheck:28, unused:11, ineffassign:1) causing CI to fail. No `.golangci.yml` configuration file existed so it ran with defaults, and ent auto-generated code was also included in lint targets, causing a large number of unnecessary issues to be reported.

## What Changes

- Add `.golangci.yml` v2 configuration with `generated: strict` exclusion and `std-error-handling` preset
- Fix ~50 errcheck violations: unchecked `defer Close()`, `json.Encode`, `tx.Rollback`, `fmt.Scanln`, etc.
- Fix ~28 staticcheck issues: QF1012 (WriteString+Sprintf→Fprintf), S1009 (redundant nil check), S1011 (append spread), SA1012 (nil context), SA9003 (empty branches), QF1003 (if/else→switch), ST1005 (error string case), S1017 (redundant HasSuffix before TrimSuffix)
- Remove 11 unused declarations: functions, struct fields, variables, imports
- Fix 1 ineffassign: dead assignment removal

## Capabilities

### New Capabilities
- `lint-configuration`: golangci-lint v2 configuration (`.golangci.yml`) with generated code exclusion and standard presets

### Modified Capabilities

(No spec-level behavior changes - all modifications are code quality improvements that don't alter functionality)

## Impact

- 20+ files modified across `internal/`, `cmd/lango/`
- No API or behavioral changes - purely code quality improvements
- CI pipeline will pass cleanly with zero lint issues
- New `.golangci.yml` establishes project-wide linting standards
