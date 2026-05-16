# pii-detector-interface Specification

## Purpose
Define the common PII detector interface and the baseline regex-based detection contract.
## Requirements
### Requirement: PIIDetector interface
The system SHALL define a PIIDetector interface with a single method `Detect(text string) []PIIMatch` that returns all PII matches found in the given text.

#### Scenario: Interface compliance
- **WHEN** RegexDetector, CompositeDetector, and PresidioDetector are implemented
- **THEN** each SHALL satisfy the PIIDetector interface at compile time

### Requirement: RegexDetector
RegexDetector SHALL compile builtin and custom patterns and detect PII by running all patterns against input text.

#### Scenario: Detect email
- **WHEN** RegexDetector is configured with RedactEmail=true and text contains "test@example.com"
- **THEN** Detect SHALL return a match with PatternName="email" and Score=1.0

#### Scenario: Detect Korean mobile
- **WHEN** RegexDetector is configured with default enabled patterns and text contains "010-1234-5678"
- **THEN** Detect SHALL return a match with PatternName="kr_mobile"

#### Scenario: Disabled builtin patterns are skipped
- **WHEN** DisabledBuiltins contains "email" and text contains "test@example.com"
- **THEN** Detect SHALL return no matches for email

#### Scenario: Custom named patterns
- **WHEN** CustomPatterns contains {"employee_id": `\bEMP-\d{6}\b`} and text contains "EMP-123456"
- **THEN** Detect SHALL return a match with PatternName="employee_id"

#### Scenario: Legacy custom regex
- **WHEN** CustomRegex contains a valid pattern and text matches it
- **THEN** Detect SHALL return a match

#### Scenario: Legacy email/phone toggle
- **WHEN** RedactEmail=false is configured
- **THEN** the email builtin pattern SHALL not be included

### Requirement: CompositeDetector
CompositeDetector SHALL chain multiple PIIDetectors and deduplicate overlapping matches.

#### Scenario: Chain multiple detectors
- **WHEN** two detectors each find different PII in the same text
- **THEN** CompositeDetector SHALL return matches from both detectors

#### Scenario: Deduplicate overlapping matches
- **WHEN** two detectors find overlapping matches at the same position
- **THEN** CompositeDetector SHALL keep only the higher-score match

#### Scenario: Empty input
- **WHEN** input text is empty
- **THEN** CompositeDetector SHALL return no matches

### Requirement: Match position tracking
Each PIIMatch SHALL include Start and End byte offsets into the original text.

#### Scenario: Accurate position
- **WHEN** text is "Email: user@test.com here" and email is detected
- **THEN** text[Start:End] SHALL equal "user@test.com"
