## ADDED Requirements

### Requirement: `lango extension inspect` subcommand
The CLI SHALL provide `lango extension inspect <source>` that accepts a local directory path or a git URL (with optional `#<ref>` suffix) and prints the inspect report to stdout. The command SHALL exit with code 0 on a readable, valid pack; 1 on user-facing error (invalid manifest, unreachable source, path-safety violation); 2 on internal error (I/O failure, malformed working copy). The command SHALL NOT write any file outside the system temp directory used for fetching.

#### Scenario: Inspect local directory
- **WHEN** the user runs `lango extension inspect ./python-dev` on a valid pack
- **THEN** the command SHALL exit 0 and print the inspect report

#### Scenario: Inspect git URL with ref
- **WHEN** the user runs `lango extension inspect https://example.com/pack.git#abc1234`
- **THEN** the command SHALL clone to a temp dir, produce the inspect report pinned to `abc1234`, and clean up the temp dir on exit

#### Scenario: Invalid manifest exits 1
- **WHEN** the source contains an `extension.yaml` that fails validation
- **THEN** the command SHALL exit 1 with an error naming the validation failure

### Requirement: `lango extension install` subcommand
The CLI SHALL provide `lango extension install <source> [--yes] [--output <format>]`. The command SHALL (a) print the inspect report, (b) unless `--yes` is set, prompt interactively for confirmation, and (c) on confirm, install the pack as specified by the `extension-pack-core` install contract. `--yes` SHALL NOT suppress the inspect output. Exit codes match `inspect` plus: exit 3 on user-denied confirmation.

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

### Requirement: `lango extension list` subcommand
The CLI SHALL provide `lango extension list [--output <format>]` that prints all packs discovered under `extensions.dir`. The default `--output table` SHALL print columns `NAME`, `VERSION`, `AUTHOR`, `INSTALLED`, `STATUS` (one of `ok`, `tampered`, `orphan`). `--output json` SHALL emit an array of records with stable field names. `--output plain` SHALL emit `<name>\t<version>\t<status>` one per line.

#### Scenario: Table output for two packs
- **WHEN** two packs are installed and the user runs `lango extension list`
- **THEN** the command SHALL print a header row and two data rows with the above columns

#### Scenario: JSON output has stable shape
- **WHEN** the user runs `lango extension list --output json`
- **THEN** the output SHALL be a JSON array where each element has `name`, `version`, `author`, `installed_at`, `source`, `status`, and `manifest_sha256` fields

#### Scenario: Empty list
- **WHEN** no packs are installed
- **THEN** `list` SHALL print a header row only (table) or `[]` (json) or nothing (plain) and exit 0

#### Scenario: Tampered status surfaced
- **WHEN** a pack's on-disk SHA-256 differs from the recorded manifest hash
- **THEN** its row SHALL show `STATUS: tampered` in table mode
- **AND** its json record SHALL carry `status: "tampered"`

### Requirement: `lango extension remove` subcommand
The CLI SHALL provide `lango extension remove <name> [--yes]` that removes a pack per the `extension-pack-core` removal contract. Without `--yes`, it SHALL prompt for confirmation and print the list of files/directories that will be deleted before prompting. Exit 0 on success; 1 if the pack is not installed; 3 on user-denied confirmation.

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

### Requirement: Consistent `--output` flag across subcommands
`inspect`, `list`, and `remove --dry-run` (if introduced later) SHALL accept `--output <table|json|plain>`. Default resolution: `table` when stdout is a TTY, `plain` otherwise. Unknown format values SHALL exit 2 with a usage error.

#### Scenario: Default on TTY
- **WHEN** stdout is a TTY and no `--output` is provided
- **THEN** the command SHALL render in `table` format

#### Scenario: Default off TTY
- **WHEN** stdout is a pipe and no `--output` is provided
- **THEN** the command SHALL render in `plain` format

#### Scenario: Unknown format rejected
- **WHEN** `--output yaml` is passed
- **THEN** the command SHALL exit 2 with a usage error naming the unsupported format

### Requirement: Help text discoverability
`lango extension --help` SHALL list the four subcommands with one-line descriptions. Each subcommand's `--help` SHALL include a usage line, flag descriptions, and at least one example. The top-level help SHALL also note the inspect-before-install trust model in one sentence.

#### Scenario: Top-level help
- **WHEN** the user runs `lango extension --help`
- **THEN** the output SHALL list `inspect`, `install`, `list`, `remove`
- **AND** SHALL include a one-sentence note on the inspect + confirm trust model

#### Scenario: Subcommand help has example
- **WHEN** the user runs `lango extension install --help`
- **THEN** the output SHALL include at least one `lango extension install <source>` example
