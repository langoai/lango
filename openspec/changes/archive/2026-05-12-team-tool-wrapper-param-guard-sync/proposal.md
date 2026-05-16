## Why

The team and team+escrow tools declared several required parameters in their tool schemas, but parts of the wrapper layer still accepted missing values and only failed later with generic messages or defaults. That leaves the operator-facing contract weaker than the declared schema and weaker than the newer meta-tool wrapper guards elsewhere in the system.

## What Changes

- Add `toolparam.RequireInt` for wrapper-level required integer extraction.
- Tighten `team_form`, `team_form_with_budget`, and `team_complete_milestone` so they enforce their declared required parameters at the wrapper layer.
- Add regression coverage for the new missing-parameter and invalid-budget paths.
- Sync specs for the team wrapper guard contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `team-connectivity`: team workflow tools now preserve actionable wrapper-level missing-parameter errors for declared required inputs.
- `production-readiness`: wrapper-level request-guard coverage now includes the team workflow tool cluster.
- `toolparam-extraction`: integer required-parameter extraction is now part of the shared helper contract.

## Impact

- Affected code: `internal/toolparam/extract.go`, `internal/p2p/team/tools.go`, `internal/p2p/team/tools_escrow.go`
- Affected tests: `internal/toolparam/extract_test.go`, `internal/p2p/team/tools_agent_test.go`
- Affected specs: `openspec/specs/team-connectivity/spec.md`, `openspec/specs/production-readiness/spec.md`, `openspec/specs/toolparam-extraction/spec.md`
