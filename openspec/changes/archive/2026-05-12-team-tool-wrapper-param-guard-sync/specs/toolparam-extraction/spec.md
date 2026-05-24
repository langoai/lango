## ADDED Requirements

### Requirement: Required integer extraction returns actionable missing-parameter errors

The shared tool parameter helper layer SHALL support required integer extraction with the same actionable missing-parameter error shape used by required string and float extraction.

#### Scenario: Missing required integer parameter returns ErrMissingParam
- **WHEN** a tool wrapper requests a required integer parameter and the key is absent or not numeric
- **THEN** the helper SHALL return `ErrMissingParam`
- **AND** the error message SHALL follow `missing <paramName> parameter`
