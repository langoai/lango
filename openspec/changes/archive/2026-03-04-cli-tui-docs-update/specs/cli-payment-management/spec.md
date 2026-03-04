## MODIFIED Requirements

### Requirement: X402 subcommand addition
The existing `lango payment` command group SHALL gain a new `x402` subcommand that displays X402 protocol configuration. This extends the payment CLI surface without modifying existing payment subcommands.

#### Scenario: Payment help includes x402
- **WHEN** user runs `lango payment --help`
- **THEN** the help output lists x402 alongside any existing payment subcommands (status, history)

### Requirement: X402 subcommand uses cfgLoader
The `payment x402` subcommand SHALL use cfgLoader to read X402 configuration from the config file. It SHALL NOT require bootLoader or database access.

#### Scenario: Config-only access
- **WHEN** user runs `lango payment x402`
- **THEN** the command loads configuration via cfgLoader and reads the payment.x402 config block

### Requirement: Existing payment commands unaffected
The addition of the x402 subcommand SHALL NOT change the behavior or registration of any existing payment subcommands.

#### Scenario: Existing commands still work
- **WHEN** user runs existing `lango payment status` command
- **THEN** the command behaves identically to before the x402 addition
