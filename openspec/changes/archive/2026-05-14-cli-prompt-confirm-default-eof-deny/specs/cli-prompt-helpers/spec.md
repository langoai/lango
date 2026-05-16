## ADDED Requirements

### Requirement: Default confirmation wrapper treats EOF as denial
The top-level `prompt.Confirm(...)` wrapper SHALL use the safer EOF-deny confirmation behavior by default.

#### Scenario: Default confirmation wrapper maps EOF to denial
- **WHEN** `prompt.Confirm(...)` reads EOF before an approval answer is received
- **THEN** it SHALL return `false` with no error
