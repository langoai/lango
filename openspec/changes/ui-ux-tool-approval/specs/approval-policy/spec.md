## ADDED Requirements

### Requirement: Example config includes approvalPolicy and exemptTools
The example `config.json` SHALL include `approvalPolicy` and `exemptTools` fields in the `security.interceptor` block to document the new configuration model.

#### Scenario: Example config fields
- **WHEN** a user inspects the example config.json
- **THEN** the `security.interceptor` block SHALL contain `"approvalPolicy": "dangerous"` and `"exemptTools": []`

### Requirement: README documents approvalPolicy
The README Security configuration table SHALL include `security.interceptor.approvalPolicy` with type `string`, default `dangerous`, and description of available policies. The legacy `security.interceptor.approvalRequired` row SHALL be marked as `(deprecated)`.

#### Scenario: README table entries
- **WHEN** a user reads the README Security section
- **THEN** the table SHALL list `approvalPolicy` (string, default "dangerous") and `exemptTools` ([]string) as configuration options
- **AND** `approvalRequired` SHALL be annotated with "(deprecated)"
