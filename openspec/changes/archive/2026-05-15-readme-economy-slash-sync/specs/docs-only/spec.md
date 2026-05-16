## ADDED Requirements

### Requirement: README economy inventory uses slash-separated escrow wording
The README internal CLI inventory SHALL describe the escrow slice of the economy family as `escrow status/list/show/sentinel status`.

#### Scenario: Stale hyphen escrow shorthand stays removed
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe `escrow status/list/show/sentinel status` instead of `escrow status-list-show/sentinel status`
