## ADDED Requirements
### Requirement: Public config CLI completeness is enforced mechanically
README and public CLI docs SHALL continue to include the implemented `config get`, `config set`, and `config keys` commands.

#### Scenario: Missing config get/set/keys docs are rejected
- **WHEN** README or public CLI docs drop one of the implemented `config get`, `config set`, or `config keys` command references
- **THEN** an executable repository test SHALL fail
