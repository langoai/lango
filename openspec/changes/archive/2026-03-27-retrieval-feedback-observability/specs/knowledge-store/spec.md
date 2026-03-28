## ADDED Requirements

### Requirement: ContextLayer String method
The `ContextLayer` type SHALL implement the `String()` method returning human-readable snake_case names for all 9 known layers. Unknown layer values SHALL return `"layer_N"` where N is the integer value.

#### Scenario: All known layers have string names
- **WHEN** `String()` is called on each of the 9 known `ContextLayer` values
- **THEN** each SHALL return a distinct non-empty snake_case string

#### Scenario: Unknown layer fallback
- **WHEN** `String()` is called on an unrecognized `ContextLayer` value
- **THEN** it SHALL return `"layer_N"` format without panicking
