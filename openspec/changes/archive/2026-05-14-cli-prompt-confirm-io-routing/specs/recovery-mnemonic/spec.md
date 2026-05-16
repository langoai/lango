## MODIFIED Requirements

### Requirement: Recovery setup output routing
`lango security recovery setup` SHALL write its mnemonic banner, written-down confirmation prompt, confirmation-word prompt, and success message through the Cobra command output stream so wrappers and test harnesses can capture non-error output without intercepting process-global stdout.

#### Scenario: Recovery setup output writes to command output
- **WHEN** the user runs `lango security recovery setup`
- **THEN** the mnemonic banner, written-down confirmation prompt, confirmation-word prompt, and success confirmation write to the Cobra command output stream
