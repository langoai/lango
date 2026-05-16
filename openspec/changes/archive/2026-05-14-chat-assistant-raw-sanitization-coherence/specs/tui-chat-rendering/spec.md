## MODIFIED Requirements

### Requirement: Assistant raw markdown reflow
Assistant transcript items SHALL preserve raw markdown for re-rendering when the viewport width changes.

#### Scenario: Stored assistant raw markdown strips control sequences
- **WHEN** assistant markdown input contains ANSI/OSC escape sequences
- **THEN** the stored raw markdown used for reflow SHALL strip those control sequences while preserving the remaining markdown/newline structure

### Requirement: DoneMsg three-rule processing
DoneMsg SHALL be processed with three rules in order:
1. If streamBuf is non-empty, finalize it as an assistant message.
2. Else if ResponseText is non-empty, add it via appendAssistant.
3. If outcome is not "success", add a compact status or error entry with deduplication.

#### Scenario: Duplicate error suppression compares sanitized assistant text
- **WHEN** DoneMsg arrives with a non-success outcome and ResponseText differs from the last assistant raw content only by stripped control sequences
- **THEN** the duplicate status/error entry SHALL still be skipped
