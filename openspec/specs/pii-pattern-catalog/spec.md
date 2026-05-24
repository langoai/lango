# pii-pattern-catalog Specification

## Purpose
Define the built-in PII pattern catalog, its categories, defaults, and validation behavior.
## Requirements
### Requirement: Builtin PII pattern catalog
The system SHALL provide a catalog of 13 builtin PII detection patterns organized into 4 categories: contact, identity, financial, and network.

#### Scenario: All builtin patterns have valid regex
- **WHEN** the builtin pattern catalog is initialized
- **THEN** every pattern SHALL compile as a valid Go regular expression

#### Scenario: Each pattern has a unique name
- **WHEN** the builtin pattern catalog is loaded
- **THEN** no two patterns SHALL have the same name

#### Scenario: Pattern lookup by name
- **WHEN** a caller looks up a builtin pattern by name (e.g., "email", "kr_rrn")
- **THEN** the system SHALL return the pattern definition and true if it exists, or false if not

### Requirement: PII pattern categories
Each PIIPatternDef SHALL have a Category field of type PIICategory with values: "contact", "identity", "financial", or "network".

#### Scenario: Contact category patterns
- **WHEN** patterns email, us_phone, kr_mobile, kr_landline, intl_phone are defined
- **THEN** each SHALL have Category "contact"

#### Scenario: Identity category patterns
- **WHEN** patterns kr_rrn, us_ssn, kr_driver, passport are defined
- **THEN** each SHALL have Category "identity"

#### Scenario: Financial category patterns
- **WHEN** patterns credit_card, kr_bank_account, iban are defined
- **THEN** each SHALL have Category "financial"

#### Scenario: Network category patterns
- **WHEN** pattern ipv4 is defined
- **THEN** it SHALL have Category "network"

### Requirement: Pattern enable/disable defaults
Each builtin pattern SHALL have an EnabledDefault field. Patterns with EnabledDefault=true SHALL be active by default. Patterns with EnabledDefault=false SHALL require explicit enablement.

#### Scenario: Default-enabled patterns
- **WHEN** no disabled patterns are configured
- **THEN** email, us_phone, kr_mobile, kr_landline, kr_rrn, us_ssn, and credit_card SHALL be active

#### Scenario: Default-disabled patterns
- **WHEN** no explicit enablement is configured
- **THEN** intl_phone, kr_driver, passport, kr_bank_account, iban, and ipv4 SHALL be inactive

### Requirement: Credit card Luhn validation
The credit_card pattern SHALL include a post-match Validate function that performs Luhn algorithm verification.

#### Scenario: Valid credit card passes Luhn
- **WHEN** a regex match "4111111111111111" is found
- **THEN** the Luhn validation SHALL return true

#### Scenario: Invalid credit card fails Luhn
- **WHEN** a regex match "4111111111111112" is found
- **THEN** the Luhn validation SHALL return false

### Requirement: Korean RRN pattern
The kr_rrn pattern SHALL match Korean resident registration numbers in format YYMMDD-GNNNNNN where G is 1-4.

#### Scenario: RRN with hyphen
- **WHEN** text contains "900101-1234567"
- **THEN** the pattern SHALL match

#### Scenario: RRN without hyphen
- **WHEN** text contains "9001011234567"
- **THEN** the pattern SHALL match

#### Scenario: Invalid gender digit
- **WHEN** text contains "900101-5234567"
- **THEN** the pattern SHALL NOT match
