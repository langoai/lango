## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Approval transcript summary truncation is Unicode-safe
- **WHEN** an approval transcript event summary is truncated for display
- **THEN** the truncation SHALL remain safe for multibyte characters instead of slicing raw bytes
