## MODIFIED Requirements

### Requirement: Secrets set command
The system SHALL provide a `lango security secrets set <name>` command that
stores an encrypted secret value either interactively (via passphrase prompt) or
non-interactively (via `--value-hex` flag). When `--value-hex` is provided, the
command SHALL hex-decode the input (stripping an optional `0x` prefix) and store
the raw bytes. When `--value-hex` is not provided, the command SHALL require an
interactive command input stream and prompt for the value. The name SHALL be a
positional argument.

#### Scenario: Interactive secret storage
- **WHEN** user runs `lango security secrets set api-key` in an interactive
  terminal without `--value-hex`
- **THEN** the command prompts for the secret value with hidden input, encrypts
  it, stores it, and displays a success message

#### Scenario: Non-interactive hex secret storage
- **WHEN** user runs `lango security secrets set wallet.privatekey --value-hex
  0xac0974...` in a non-interactive environment
- **THEN** the command SHALL hex-decode the value (stripping `0x` prefix), store
  the raw bytes encrypted, and print success

#### Scenario: Non-interactive without value-hex flag
- **WHEN** user runs `lango security secrets set api-key` with a non-interactive
  command input stream and without `--value-hex`
- **THEN** the command exits with an error suggesting `--value-hex` for
  non-interactive use

#### Scenario: Invalid hex value
- **WHEN** user runs `lango security secrets set mykey --value-hex "not-hex"`
- **THEN** the command SHALL return a hex decode error

#### Scenario: Secrets-set guard uses command input stream
- **WHEN** `lango security secrets set <name>` prompts interactively
- **THEN** its interactive guard SHALL validate `cmd.InOrStdin()` instead of
  reading process-global stdin directly
