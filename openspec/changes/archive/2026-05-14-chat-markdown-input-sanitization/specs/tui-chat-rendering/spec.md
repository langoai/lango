## MODIFIED Requirements

### Requirement: Markdown rendering fails closed to plain text
The chat TUI SHALL preserve transcript stability when the markdown renderer errors or panics.

#### Scenario: Sanitized markdown input is used for rendering
- **WHEN** assistant markdown input contains ANSI/OSC escape sequences
- **THEN** the chat TUI SHALL strip those control sequences before rendering the markdown

#### Scenario: Sanitized markdown input is used for plain-text fallback
- **WHEN** assistant markdown rendering fails after receiving input that contained ANSI/OSC escape sequences
- **THEN** the fallback plain-text transcript content SHALL use the sanitized text rather than the raw control-sequence input
