## ADDED Requirements
### Requirement: Companion connectivity docs guard stays executable
Repository-level regressions that reintroduce stale automatic companion-discovery claims into public docs or main specs SHALL be enforced by an executable test.

#### Scenario: Stale companion discovery claims are rejected
- **WHEN** companion connectivity docs or specs reintroduce automatic Bonjour/mDNS discovery claims or a legacy dedicated companion-address config key
- **THEN** an executable repository test SHALL fail
