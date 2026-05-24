## ADDED Requirements

### Requirement: Smart account module install output routing

`lango account module install` SHALL write the success summary through the Cobra command output stream so wrappers and test harnesses can capture installation confirmation without intercepting process-global stdout.

#### Scenario: Smart account module install success writes to command output
- **WHEN** `lango account module install <module-address> --type <module-type>` succeeds
- **THEN** the command writes the installation summary to the Cobra command output stream
