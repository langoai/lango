## ADDED Requirements

### Requirement: Orphaned FunctionResponse removal after sanitization
After the sanitization pipeline completes, the Gemini provider SHALL run `dropOrphanedFunctionResponses()` to remove FunctionResponse parts whose corresponding FunctionCall is not present in the content sequence. This handles cases where FunctionCalls were dropped by thought filtering or other upstream processing.

#### Scenario: Orphan response removed when FunctionCall was dropped
- **WHEN** the sanitized content contains a FunctionResponse with ID "call_thought" but no FunctionCall with that ID exists
- **THEN** `dropOrphanedFunctionResponses()` SHALL remove the orphaned FunctionResponse part

#### Scenario: Matched response preserved
- **WHEN** the sanitized content contains both a FunctionCall and FunctionResponse with matching ID "call_real"
- **THEN** `dropOrphanedFunctionResponses()` SHALL preserve the FunctionResponse

#### Scenario: Content block with only orphaned responses is removed entirely
- **WHEN** a Content block contains only FunctionResponse parts and all are orphaned
- **THEN** `dropOrphanedFunctionResponses()` SHALL remove the entire Content block from the sequence

#### Scenario: No FunctionCalls in sequence removes all FunctionResponses
- **WHEN** the content sequence contains no FunctionCall parts at all
- **THEN** all FunctionResponse parts SHALL be treated as orphaned and removed
