## ADDED Requirements
### Requirement: Companion connectivity docs stay truthful about the shipped model
Public docs and main specs SHALL not claim that automatic companion discovery ships when the current runtime uses gateway-backed companion connections.

#### Scenario: Stale companion discovery claims are rejected
- **WHEN** a maintainer updates companion connectivity docs or specs
- **THEN** they SHALL not claim automatic Bonjour/mDNS discovery or a legacy dedicated companion-address config key as current shipped behavior
