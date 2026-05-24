## MODIFIED Requirements

### Requirement: Markdown rendering fails closed to plain text
The chat TUI SHALL preserve transcript stability when the markdown renderer errors or panics.

#### Scenario: Renderer error falls back to plain text
- **WHEN** `renderMarkdown()` cannot render through Glamour because the renderer returns an error
- **THEN** it SHALL return the original content as plain text instead of failing the transcript

#### Scenario: Renderer panic falls back to plain text
- **WHEN** `renderMarkdown()` encounters a renderer panic
- **THEN** it SHALL recover and return the original content as plain text instead of crashing the TUI
