## MODIFIED Requirements

### Requirement: `lango extension inspect` subcommand

The CLI SHALL provide `lango extension inspect <source>` that accepts a local directory path or a git URL (with optional `#<ref>` suffix) and prints the inspect report to stdout. The command SHALL exit with code 0 on a readable, valid pack; 1 on user-facing error (invalid manifest, unreachable source, path-safety violation); 2 on internal error (I/O failure, malformed working copy). The command SHALL NOT write any file outside the system temp directory used for fetching.

#### Scenario: Invalid manifest exits 1 through root-owned process exit
- **WHEN** the source contains an `extension.yaml` that fails validation
- **THEN** the command SHALL return a structured CLI error carrying exit code 1 to the root entrypoint
- **AND** the root entrypoint SHALL terminate with exit code 1

### Requirement: `lango extension install` subcommand

The CLI SHALL provide `lango extension install <source> [--yes] [--output <format>]`. The command SHALL (a) print the inspect report, (b) unless `--yes` is set, prompt interactively for confirmation through the shared confirmation helper using Cobra command input/output streams, and (c) on confirm, install the pack as specified by the `extension-pack-core` install contract. `--yes` SHALL NOT suppress the inspect output. Exit codes match `inspect` plus: exit 3 on user-denied confirmation. Extension command packages SHALL return structured CLI errors for non-zero exit outcomes rather than calling `os.Exit` directly.

#### Scenario: Interactive install denied
- **WHEN** the user answers `n` or sends EOF at the prompt
- **THEN** the command SHALL return a structured CLI error carrying exit code 3 without writing any files
- **AND** a message stating "install cancelled by user" SHALL be printed

### Requirement: `lango extension remove` subcommand

The CLI SHALL provide `lango extension remove <name> [--yes]` that removes a pack per the `extension-pack-core` removal contract. Without `--yes`, it SHALL print the list of files/directories that will be deleted, then prompt for confirmation through the shared confirmation helper using Cobra command input/output streams. Exit 0 on success; 1 if the pack is not installed; 3 on user-denied confirmation. Extension command packages SHALL return structured CLI errors for non-zero exit outcomes rather than calling `os.Exit` directly.

#### Scenario: Remove unknown pack
- **WHEN** the user runs `lango extension remove missing`
- **THEN** the command SHALL return a structured CLI error carrying exit code 1 with an error naming the pack
