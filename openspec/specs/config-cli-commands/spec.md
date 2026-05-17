## Purpose

Define the `lango config` CLI subcommands for managing encrypted configuration profiles (list, import, export, validate, set, get, delete).
## Requirements
### Requirement: Config list command
The system SHALL provide a `lango config list` command that displays all profiles with name, active marker, version, and timestamps in a table format.

#### Scenario: List with profiles
- **WHEN** `lango config list` is run with existing profiles
- **THEN** a table is displayed with columns: NAME, ACTIVE, VERSION, CREATED, UPDATED

#### Scenario: List with no profiles
- **WHEN** no profiles exist
- **THEN** the message "No profiles found." is displayed

### Requirement: Config create command
The system SHALL provide a `lango config create <name>` command that creates a new profile with default configuration.

#### Scenario: Create new profile
- **WHEN** `lango config create staging` is run and "staging" does not exist
- **THEN** a profile named "staging" is created with default config

#### Scenario: Create duplicate profile
- **WHEN** `lango config create default` is run and "default" already exists
- **THEN** an error is returned: "profile \"default\" already exists"

#### Scenario: Config profile-management output uses the command writer
- **WHEN** `lango config list`, `create`, `use`, `delete`, or `import` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** `delete` confirmation SHALL read from `cmd.InOrStdin()`
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` / `cmd.InOrStdin()` SHALL capture the interaction

### Requirement: Config use command
The system SHALL provide a `lango config use <name>` command that switches the active profile.

#### Scenario: Switch active profile
- **WHEN** `lango config use production` is run
- **THEN** the "production" profile becomes active and all others are deactivated

### Requirement: Config delete command
The system SHALL provide a `lango config delete <name>` command with confirmation prompt. When `--force` is not supplied, the confirmation SHALL be driven through the shared confirmation helper using Cobra command input/output streams so wrappers and tests can capture the interaction without replacing process-global stdio.

#### Scenario: Delete with confirmation
- **WHEN** `lango config delete staging` is run without `--force`
- **THEN** a confirmation prompt is shown before deletion

#### Scenario: Delete with force flag
- **WHEN** `lango config delete staging --force` is run
- **THEN** the profile is deleted without confirmation

#### Scenario: Delete denied through command input
- **WHEN** `lango config delete staging` is run and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** the profile SHALL remain undeleted

### Requirement: Config import command
The system SHALL provide a `lango config import <file>` command that imports a JSON config file as an encrypted profile. The source JSON file SHALL be automatically deleted after successful import for security.

#### Scenario: Import JSON file
- **WHEN** `lango config import lango.json --profile migrated` is run
- **THEN** the JSON file is loaded, encrypted, and stored as profile "migrated" (set as active)
- **AND** the source JSON file is deleted after successful import
- **AND** the message "Source file deleted for security." is displayed

#### Scenario: Import with delete failure
- **WHEN** import succeeds but the source file cannot be deleted (e.g., permission denied)
- **THEN** a warning is logged but the command does not fail

### Requirement: Config export command
The system SHALL provide a `lango config export <name>` command that outputs decrypted config as JSON. Passphrase verification is required (handled implicitly by the bootstrap process).

#### Scenario: Export profile
- **WHEN** `lango config export default` is run
- **THEN** the passphrase is verified via bootstrap
- **AND** the decrypted config is printed to stdout as formatted JSON, with a WARNING on stderr

#### Scenario: Config export/validate output uses the command writer
- **WHEN** `lango config export` or `lango config validate` renders output
- **THEN** it SHALL write the full command output through the Cobra command output or error writers
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` SHALL capture the output

#### Scenario: Config JSON output remains decodable on the command writer
- **WHEN** `lango config export` or `lango config get <path> --output json` renders JSON output
- **THEN** the command writer SHALL receive valid pretty-printed JSON that can be decoded without stripping extra framing text

#### Scenario: Config get rejects an unknown output format before loading config
- **WHEN** `lango config get <path> --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL reject the invocation before loading the active config

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

`config set` SHALL preserve the loaded profile's `ExplicitKeys` metadata when saving. When the set path is one of `config.ContextRelatedKeys()`, the saved explicit-key map SHALL include that path so later context auto-enable resolution treats the value as user-explicit.

Invalid dot-path errors from `config set` SHALL include actionable key discovery help. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. The command SHALL NOT save after an invalid path.

`config set` SHALL support map-backed dot paths whose map keys are dynamic user names. When a map entry is missing and the remaining path is valid for the map value type, the command SHALL create the map and entry before setting the requested value. When the final map value is a `string`, the final path segment SHALL be treated as the map key and the provided value SHALL be stored as the map value.

#### Scenario: Successful set closes DB client
- **WHEN** `config set agent.provider openai` succeeds
- **THEN** the DB client is closed after the command completes

#### Scenario: setConfigPath error closes DB client
- **WHEN** `config set invalid.key value` fails at setConfigPath
- **THEN** the cleanup function is still called via defer, closing the DB client

#### Scenario: Save error closes DB client
- **WHEN** save fails (e.g., validation error from PostLoad in Save)
- **THEN** the cleanup function is still called via defer, closing the DB client

#### Scenario: Loader failure does not leak resources
- **WHEN** the cfgLoader fails (bootstrap error)
- **THEN** cleanup is nil, defer is a no-op, no DB client exists to leak

#### Scenario: Config get/set/keys output uses the command writer
- **WHEN** `lango config get`, `lango config set`, or `lango config keys` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Existing explicit keys preserved
- **WHEN** `lango config set agent.provider openai` saves a profile whose `ExplicitKeys` includes `knowledge.enabled`
- **THEN** the saved profile SHALL still include `knowledge.enabled` in `ExplicitKeys`

#### Scenario: Context-related set marks explicit
- **WHEN** `lango config set knowledge.enabled false` saves a profile
- **THEN** the saved profile SHALL include `knowledge.enabled` in `ExplicitKeys`

#### Scenario: Invalid set path does not mutate explicit keys
- **WHEN** `lango config set invalid.key value` fails before save
- **THEN** the loaded explicit-key map SHALL NOT be mutated

#### Scenario: Config set suggests nearby key on invalid path
- **WHEN** `lango config set knowledge.enable false` is run
- **THEN** the command SHALL fail before saving
- **AND** the error SHALL include `knowledge.enabled`
- **AND** the error SHALL include `lango config keys knowledge`

#### Scenario: Config set leaf-extension path suggests valid leaf key
- **WHEN** `lango config set agent.provider.extra openai` is run
- **THEN** the command SHALL fail before saving
- **AND** the error SHALL include `agent.provider`
- **AND** the error SHALL include `lango config keys agent`

#### Scenario: Config set creates provider map entry
- **WHEN** `lango config set providers.openai.type openai` is run on a profile with no `providers.openai` entry
- **THEN** the saved profile SHALL include `providers.openai.type` set to `openai`

#### Scenario: Config set updates map-backed struct field
- **WHEN** `lango config set providers.openai.baseUrl http://localhost:11434/v1` is run on a profile with an existing `providers.openai` entry
- **THEN** the saved profile SHALL update only the `providers.openai.baseUrl` field

#### Scenario: Config set creates nested string map value
- **WHEN** `lango config set mcp.servers.docs.env.LOG_LEVEL debug` is run on a profile with no `mcp.servers.docs` entry
- **THEN** the saved profile SHALL include `mcp.servers.docs.env.LOG_LEVEL` set to `debug`

#### Scenario: Invalid map-backed set path does not save
- **WHEN** `lango config set providers.openai.notAField value` is run
- **THEN** the command SHALL fail before saving

### Requirement: Config get actionable key errors
The `config get <dot.path>` command SHALL return actionable key discovery help when the dot path cannot be resolved. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. When no nearby keys exist, the error SHALL still include a `lango config keys` discovery hint.

#### Scenario: Config get suggests nearby key on invalid path
- **WHEN** `lango config get agent.providr` is run
- **THEN** the command SHALL fail with an error containing `agent.provider`
- **AND** the error SHALL include `lango config keys agent`

#### Scenario: Config get unknown top-level path gives discovery hint
- **WHEN** `lango config get made.up.path` is run
- **THEN** the command SHALL fail with an error containing `lango config keys`

#### Scenario: Config get leaf-extension path suggests valid leaf key
- **WHEN** `lango config get agent.provider.extra` is run
- **THEN** the command SHALL fail with an error containing `agent.provider`
- **AND** the error SHALL include `lango config keys agent`

### Requirement: Config validate command
The system SHALL provide a `lango config validate` command that validates the active profile's configuration.

#### Scenario: Valid config
- **WHEN** the active profile's config passes validation
- **THEN** the message "Profile \"default\" configuration is valid." is displayed

### Requirement: Config profile commands in configcmd package
The config profile subcommands (list, create, use, delete, import, export, validate) SHALL be defined in the `configcmd` package via `configcmd.NewConfigCmd()`, not inline in main.go. The returned `*cobra.Command` SHALL include all 7 profile subcommands.

#### Scenario: NewConfigCmd provides all profile subcommands
- **WHEN** `configcmd.NewConfigCmd(bootLoader)` is called
- **THEN** the returned command SHALL include subcommands: list, create, use, delete, import, export, validate

#### Scenario: main.go delegates to configcmd
- **WHEN** the root command is assembled in main.go
- **THEN** config profile commands SHALL be added via configcmd.NewConfigCmd()
- **THEN** main.go SHALL NOT contain RunE implementations for profile commands

### Requirement: Config delete treats EOF as denial
`lango config delete <name>` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Config delete EOF aborts cleanly
- **WHEN** `lango config delete staging` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** the profile SHALL remain undeleted
