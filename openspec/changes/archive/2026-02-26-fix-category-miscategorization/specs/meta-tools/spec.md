## MODIFIED Requirements

### Requirement: save_knowledge Tool Category Validation
The `save_knowledge` tool SHALL validate the category parameter using `entknowledge.CategoryValidator()` before persisting. The tool SHALL accept the following categories: `rule`, `definition`, `preference`, `fact`, `pattern`, `correction`. Invalid categories SHALL return an error to the caller.

#### Scenario: Valid category accepted
- **WHEN** the `save_knowledge` tool is called with a valid category (rule, definition, preference, fact, pattern, correction)
- **THEN** the knowledge entry SHALL be saved successfully

#### Scenario: Invalid category rejected
- **WHEN** the `save_knowledge` tool is called with an unrecognized category
- **THEN** the tool SHALL return an error indicating the invalid category without saving

#### Scenario: Tool schema includes all categories
- **WHEN** the tool parameters are inspected
- **THEN** the `category` enum SHALL include all six valid values: rule, definition, preference, fact, pattern, correction
