## ADDED Requirements

### Requirement: Public workbench docs explain starter prompt context-awareness

Public docs that describe the standalone workbench quick-start flow SHALL explain that ready-profile starter prompts adapt to the detected workspace context.

#### Scenario: Docs describe context-aware prompt behavior
- **WHEN** README or CLI/TUI docs describe ready-profile starter prompts
- **THEN** they SHALL mention that the prompts adapt to the detected workdir or repository
- **AND** they SHALL mention the Go-specific structure guidance when a `go.mod` is present
