## ADDED Requirements

### Requirement: Recovery setup output routing

`lango security recovery setup` SHALL write its mnemonic banner, confirmation-word prompt, and success message through the Cobra command output stream so wrappers and test harnesses can capture non-error output without intercepting process-global stdout.

#### Scenario: Recovery setup output writes to command output
- **WHEN** the user runs `lango security recovery setup`
- **THEN** the mnemonic banner, confirmation-word prompt, and success confirmation write to the Cobra command output stream

### Requirement: Recovery restore output routing

`lango security recovery restore` SHALL write its success confirmation through the Cobra command output stream so wrappers and test harnesses can capture completion output without intercepting process-global stdout.

#### Scenario: Recovery restore success writes to command output
- **WHEN** the user runs `lango security recovery restore` and enters the correct mnemonic
- **THEN** the success confirmation writes to the Cobra command output stream
