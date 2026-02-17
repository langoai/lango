## Purpose

The Tool Safety Level capability provides an intrinsic risk classification system for agent tools. Each tool carries a SafetyLevel (Safe, Moderate, Dangerous) that determines its default approval requirements. The zero value is treated as Dangerous for fail-safe behavior.

## Requirements

### Requirement: SafetyLevel enum type
The system SHALL define a `SafetyLevel` integer enum type with three levels: `SafetyLevelSafe` (1), `SafetyLevelModerate` (2), `SafetyLevelDangerous` (3), starting at `iota + 1`.

#### Scenario: Enum values
- **WHEN** SafetyLevel constants are defined
- **THEN** SafetyLevelSafe SHALL equal 1, SafetyLevelModerate SHALL equal 2, SafetyLevelDangerous SHALL equal 3

### Requirement: Zero-value fail-safe behavior
The system SHALL treat the zero value of SafetyLevel (unset) as Dangerous. Any tool that does not explicitly set its SafetyLevel SHALL be treated as Dangerous.

#### Scenario: Unset SafetyLevel
- **WHEN** a Tool has SafetyLevel zero value (0)
- **THEN** `IsDangerous()` SHALL return true
- **AND** `String()` SHALL return "dangerous"

### Requirement: String representation
The system SHALL provide a `String()` method on SafetyLevel that returns "safe", "moderate", or "dangerous". Unknown values SHALL return "dangerous".

#### Scenario: Known levels
- **WHEN** `SafetyLevelSafe.String()` is called
- **THEN** it SHALL return "safe"

#### Scenario: Unknown level
- **WHEN** `SafetyLevel(99).String()` is called
- **THEN** it SHALL return "dangerous"

### Requirement: IsDangerous helper
The system SHALL provide an `IsDangerous()` method that returns true for SafetyLevelDangerous and for the zero value.

#### Scenario: Safe tool
- **WHEN** `SafetyLevelSafe.IsDangerous()` is called
- **THEN** it SHALL return false

#### Scenario: Dangerous tool
- **WHEN** `SafetyLevelDangerous.IsDangerous()` is called
- **THEN** it SHALL return true

### Requirement: SafetyLevel field on Tool struct
The system SHALL add a `SafetyLevel` field to the `agent.Tool` struct.

#### Scenario: Tool with SafetyLevel
- **WHEN** a Tool is created with `SafetyLevel: SafetyLevelDangerous`
- **THEN** the SafetyLevel field SHALL be accessible on the Tool

### Requirement: All built-in tools have SafetyLevel assigned
Every built-in tool SHALL have an explicit SafetyLevel assignment. The classifications SHALL be:
- **Dangerous**: exec, exec_bg, exec_stop, fs_write, fs_edit, fs_delete, browser_navigate, browser_action, crypto_encrypt, crypto_decrypt, crypto_sign, secrets_store, secrets_get, secrets_delete, payment_send
- **Moderate**: fs_mkdir, save_knowledge, save_learning, create_skill
- **Safe**: exec_status, fs_read, fs_list, browser_screenshot, crypto_hash, crypto_keys, secrets_list, search_knowledge, search_learnings, list_skills, payment_balance, payment_history, payment_limits, payment_wallet_info

#### Scenario: Exec tool classification
- **WHEN** the exec tool is created
- **THEN** its SafetyLevel SHALL be SafetyLevelDangerous

#### Scenario: Read tool classification
- **WHEN** the fs_read tool is created
- **THEN** its SafetyLevel SHALL be SafetyLevelSafe

### Requirement: SafetyLevel preservation through wrappers
The system SHALL preserve the SafetyLevel field when wrapping tools with `wrapWithLearning` or `wrapWithApproval`.

#### Scenario: Learning wrapper preserves level
- **WHEN** a tool with SafetyLevelDangerous is wrapped with wrapWithLearning
- **THEN** the resulting tool SHALL have SafetyLevelDangerous

### Requirement: Tool approval summary for payment_send
The `buildApprovalSummary` function SHALL include a case for `payment_send` that formats the summary as "Send {amount} USDC to {to} ({purpose})".

#### Scenario: Payment send approval message
- **WHEN** approval is requested for `payment_send` with amount "1.50", to "0x1234...5678", and purpose "API access"
- **THEN** the summary is "Send 1.50 USDC to 0x1234...567... (API access)"
