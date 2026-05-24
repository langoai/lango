## MODIFIED Requirements

### Requirement: `lango extension install` subcommand
The CLI SHALL provide `lango extension install <source> [--yes] [--output <format>]`. The command SHALL (a) print the inspect report, (b) unless `--yes` is set, prompt interactively for confirmation through the shared confirmation helper using Cobra command input/output streams, and (c) on confirm, install the pack as specified by the `extension-pack-core` install contract. `--yes` SHALL NOT suppress the inspect output. Exit codes match `inspect` plus: exit 3 on user-denied confirmation.

#### Scenario: Interactive install confirmed
- **WHEN** the user runs `lango extension install ./python-dev` and answers `y` at the prompt
- **THEN** the command SHALL exit 0 after a successful install

#### Scenario: Interactive install denied
- **WHEN** the user answers `n` or sends EOF at the prompt
- **THEN** the command SHALL exit 3 without writing any files
- **AND** a message stating "install cancelled by user" SHALL be printed

#### Scenario: --yes skips prompt but prints inspect
- **WHEN** the user runs `lango extension install --yes ./python-dev`
- **THEN** the command SHALL print the inspect report AND install the pack without prompting
- **AND** exit 0 on success

#### Scenario: Non-TTY stdin defaults to deny
- **WHEN** the user runs `lango extension install <pack>` without `--yes` and stdin is not a TTY
- **THEN** the command SHALL exit 3 with a message directing the user to pass `--yes` for scripted installs

### Requirement: `lango extension remove` subcommand
The CLI SHALL provide `lango extension remove <name> [--yes]` that removes a pack per the `extension-pack-core` removal contract. Without `--yes`, it SHALL print the list of files/directories that will be deleted, then prompt for confirmation through the shared confirmation helper using Cobra command input/output streams. Exit 0 on success; 1 if the pack is not installed; 3 on user-denied confirmation.

#### Scenario: Remove with confirmation
- **WHEN** the user runs `lango extension remove python-dev` and answers `y`
- **THEN** the command SHALL delete the pack and its `ext-python-dev/` skill subdir, then exit 0

#### Scenario: Remove with --yes
- **WHEN** the user runs `lango extension remove --yes python-dev`
- **THEN** the command SHALL skip the prompt and remove the pack
- **AND** the list of to-be-deleted paths SHALL still be printed to stdout

#### Scenario: Remove unknown pack
- **WHEN** the user runs `lango extension remove missing`
- **THEN** the command SHALL exit 1 with an error naming the pack
