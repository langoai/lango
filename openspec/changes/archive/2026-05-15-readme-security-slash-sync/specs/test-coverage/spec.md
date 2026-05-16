## ADDED Requirements

### Requirement: README security inventory slash-form guard stays executable
Repository-level regressions that reintroduce hyphen-compressed subfamily wording into the README internal security inventory SHALL be enforced by an executable test.

#### Scenario: Stale hyphen shorthand is rejected
- **WHEN** the README internal tree still documents the shipped security surface
- **THEN** an executable repository test SHALL fail if it falls back to `store-clear-status`, `setup-restore`, or `status-test-keys-wrap-detach`
