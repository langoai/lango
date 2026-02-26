## MODIFIED Requirements

### Requirement: Category Mapping
The system SHALL map LLM analysis type strings to valid `entknowledge.Category` enum values. The `mapCategory()` and `mapKnowledgeCategory()` functions SHALL return `(Category, error)` and SHALL return an error for any unrecognized type string instead of silently defaulting. Valid types SHALL include: `preference`, `fact`, `rule`, `definition`, `pattern`, `correction`.

#### Scenario: Valid type mapping
- **WHEN** a recognized type string (preference, fact, rule, definition, pattern, correction) is passed to `mapCategory()` or `mapKnowledgeCategory()`
- **THEN** the corresponding `entknowledge.Category` value SHALL be returned with a nil error

#### Scenario: Unrecognized type rejection
- **WHEN** an unrecognized type string is passed to `mapCategory()` or `mapKnowledgeCategory()`
- **THEN** an empty category and a non-nil error containing `"unrecognized knowledge type"` SHALL be returned

#### Scenario: Case sensitivity
- **WHEN** a type string with incorrect casing (e.g., `"FACT"`, `"Preference"`) is passed
- **THEN** the function SHALL return an error (types are case-sensitive)
