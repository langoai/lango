## ADDED Requirements

### Requirement: Settings command output routing

`lango settings` SHALL write its cancel message and post-save guidance through the Cobra command output stream so wrappers and test harnesses can capture non-TUI completion output without intercepting process-global stdout.

#### Scenario: Settings post-save guidance writes to command output
- **WHEN** settings saves successfully
- **THEN** the post-save guidance writes to the Cobra command output stream
