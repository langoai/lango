## ADDED Requirements

### Requirement: CLI index includes dedicated core and status command sections
The public CLI index SHALL include dedicated sections for the implemented core and status command families so the index structure matches the existing `docs/cli/core.md` and `docs/cli/status.md` references.

#### Scenario: Implemented core and status sections stay discoverable
- **WHEN** a maintainer updates `docs/cli/index.md`
- **THEN** it SHALL include dedicated `Core Commands` and `Status Dashboard` sections covering the implemented `lango`/`cockpit`/`chat`/`serve`/`version`/`health`/`onboard`/`settings`/`doctor` entries and the `lango status` dead-letter command family
