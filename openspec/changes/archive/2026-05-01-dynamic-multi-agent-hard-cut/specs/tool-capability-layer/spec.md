## ADDED Requirements

### Requirement: Capability runtime rechecks grant state before blocked projection
Before persisting a `blocked_waiting_approval` projection for a built-in teammate run, the runtime SHALL re-read the latest grant and allowlist state. The check narrows the stale projection window but correctness is measured on the final observed run state rather than every intermediate transition.

#### Scenario: Final observed run state is clear after grant wins the race
- **WHEN** a grant becomes effective before the final observed state is read
- **THEN** the final observed teammate run state SHALL NOT remain stuck in `blocked_waiting_approval`
- **AND** the runtime regression tests SHALL be the concrete verification gate for that condition
