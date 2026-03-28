## ADDED Requirements

### Requirement: Knowledge retrieval result truncation
The system SHALL provide a `TruncateResult(result *RetrievalResult, budgetTokens int) *RetrievalResult` function that reduces a `RetrievalResult` to fit within a token budget. Truncation SHALL operate at the item level — removing lower-priority items from each layer until the total estimated tokens fit within the budget. The function SHALL NOT modify assembled text; it operates on `RetrievalResult` before `AssemblePrompt()` is called.

#### Scenario: Result fits within budget
- **WHEN** `TruncateResult` is called with a result whose total tokens are within budget
- **THEN** the result SHALL be returned unchanged

#### Scenario: Result exceeds budget
- **WHEN** `TruncateResult` is called with a result exceeding the budget
- **THEN** items SHALL be removed from the end of each layer (lowest priority first) until the total fits
- **AND** the layer structure and headings SHALL remain intact

#### Scenario: Zero budget means unlimited
- **WHEN** `TruncateResult` is called with `budgetTokens == 0`
- **THEN** the result SHALL be returned unchanged (0 = unlimited, legacy mode)

#### Scenario: Budget too small for any items
- **WHEN** `TruncateResult` is called with a budget smaller than any single item
- **THEN** the result SHALL be empty (zero items) but not nil
