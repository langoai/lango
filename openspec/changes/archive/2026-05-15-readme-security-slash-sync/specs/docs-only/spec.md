## ADDED Requirements

### Requirement: README security inventory uses slash-separated subfamily wording
The README internal CLI inventory SHALL describe keyring, recovery, and KMS subfamilies with slash-separated wording rather than stale hyphen-compressed shorthand.

#### Scenario: Stale hyphen shorthand stays removed
- **WHEN** a maintainer updates the README internal tree security inventory
- **THEN** it SHALL use `keyring store/clear/status`, `recovery setup/restore`, and `kms status/test/keys/wrap/detach`
