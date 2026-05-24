## ADDED Requirements

### Requirement: KMS wrap and detach output routing

`lango security kms wrap` and `lango security kms detach` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture success confirmations and multi-slot guidance without intercepting process-global stdout.

#### Scenario: Wrap success writes to command output
- **WHEN** `lango security kms wrap --provider aws-kms --key-id <arn>` succeeds
- **THEN** the command writes the wrap success confirmation to the Cobra command output stream

#### Scenario: Detach success writes to command output
- **WHEN** `lango security kms detach` removes a KMS slot
- **THEN** the command writes the detach success confirmation to the Cobra command output stream

#### Scenario: Detach multi-slot guidance writes to command output
- **WHEN** `lango security kms detach` requires `--slot-id` because multiple hardware slots exist
- **THEN** the command writes the slot listing and guidance to the Cobra command output stream
