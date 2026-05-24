## ADDED Requirements

### Requirement: Onboard command output routing

`lango onboard` SHALL write its preset banner, cancel message, and post-save guidance through the Cobra command output stream so wrappers and test harnesses can capture non-TUI completion output without intercepting process-global stdout.

#### Scenario: Onboard next-steps guidance writes to command output
- **WHEN** onboard completes successfully
- **THEN** the post-save guidance writes to the Cobra command output stream
