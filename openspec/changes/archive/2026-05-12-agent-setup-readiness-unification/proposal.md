## Why

The workbench and onboarding test step both decide whether an agent profile is usable, but that logic had drifted into multiple local checks. That makes small readiness fixes risky because one surface can become stricter or looser than another.

## What Changes

- Centralize agent profile readiness evaluation in `internal/config`.
- Reuse that shared readiness contract in onboarding validation and workbench setup-recovery signals.
- Add regression coverage for remote-provider API-key requirements and the Ollama no-key exception.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-onboard`: Test Configuration now uses the shared agent readiness contract for provider/model/API-key validation.
- `mission-workbench-tui`: Header and setup guidance continue using one shared readiness contract for incomplete versus ready profiles.

## Impact

- Affected code: `internal/config/agent_readiness.go`, `internal/cli/onboard/test_step.go`, `internal/cli/cockpit/missioncontrol_projector.go`, `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/config/agent_readiness_test.go`
