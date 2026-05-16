## Purpose

Define the CLI commands for inspecting, querying, and managing the knowledge graph store.
## Requirements
### Requirement: Graph status command
The system SHALL provide a `lango graph status [--output table|json]` command that displays the graph store configuration and triple count.

#### Scenario: Graph disabled
- **WHEN** user runs `lango graph status` with graph.enabled=false
- **THEN** system displays that graph store is not enabled

#### Scenario: Graph enabled with JSON output
- **WHEN** user runs `lango graph status --output json` with graph.enabled=true
- **THEN** system outputs JSON with enabled, backend, database_path, and triple_count fields

#### Scenario: Graph status rejects unknown output before config load
- **WHEN** user runs `lango graph status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Graph status output uses the command writer
- **WHEN** `lango graph status` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Graph query command
The system SHALL provide a `lango graph query [--output table|json]` command that queries triples by subject, object, or subject+predicate. At least one of `--subject` or `--object` MUST be provided. The `--predicate` flag requires `--subject`. A `--limit` flag SHALL cap results.

#### Scenario: Query by subject
- **WHEN** user runs `lango graph query --subject "entity1"`
- **THEN** system displays matching triples in SUBJECT/PREDICATE/OBJECT tabwriter format

#### Scenario: Graph query output uses the command writer
- **WHEN** `lango graph query` renders text or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: No filter provided
- **WHEN** user runs `lango graph query` without --subject or --object
- **THEN** system returns an error indicating at least one filter is required

#### Scenario: Graph query rejects unknown output before config load
- **WHEN** user runs `lango graph query --subject "entity1" --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Graph stats command
The system SHALL provide a `lango graph stats [--output table|json]` command that displays total triple count and per-predicate breakdown sorted by count descending.

#### Scenario: Stats with data
- **WHEN** user runs `lango graph stats` with populated graph
- **THEN** system displays total triple count and PREDICATE/COUNT table

#### Scenario: Graph stats output uses the command writer
- **WHEN** `lango graph stats` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Graph stats reject unknown output before config load
- **WHEN** user runs `lango graph stats --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Graph clear command
The system SHALL provide a `lango graph clear` command that removes all triples from the graph store. The command SHALL prompt for confirmation unless `--force` is provided.

#### Scenario: Graph clear prompt uses command streams
- **WHEN** `lango graph clear` prompts for confirmation
- **THEN** it SHALL write the prompt through the Cobra command output writer
- **AND** it SHALL read the operator response through the Cobra command input reader
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` and `cmd.InOrStdin()` SHALL control the interaction

#### Scenario: Clear with confirmation
- **WHEN** user runs `lango graph clear` and confirms with "y"
- **THEN** system clears all triples and prints confirmation message

#### Scenario: Clear aborted
- **WHEN** user runs `lango graph clear` and does not confirm
- **THEN** system prints "Aborted." and makes no changes

#### Scenario: Force clear
- **WHEN** user runs `lango graph clear --force`
- **THEN** system clears all triples without prompting

### Requirement: Graph export command
The system SHALL provide a `lango graph export [--format json|csv]` command that streams all triples from the graph store to command output.

#### Scenario: Graph export output uses the command writer
- **WHEN** `lango graph export` renders JSON or CSV output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Graph export JSON output
- **WHEN** user runs `lango graph export --format json`
- **THEN** the command outputs a JSON array of triples

#### Scenario: Graph export CSV output
- **WHEN** user runs `lango graph export --format csv`
- **THEN** the command outputs CSV with a `subject,predicate,object` header row

### Requirement: Graph add command
The system SHALL provide a `lango graph add --subject <s> --predicate <p> --object <o> [--output table|json]` command that persists a single triple to the graph store.

#### Scenario: Graph add output uses the command writer
- **WHEN** `lango graph add` renders text or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Graph add JSON output
- **WHEN** user runs `lango graph add --subject <s> --predicate <p> --object <o> --output json`
- **THEN** the command outputs a JSON object representing the added triple

#### Scenario: Graph add rejects unknown output before config load
- **WHEN** user runs `lango graph add --subject <s> --predicate <p> --object <o> --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Graph import command
The system SHALL provide a `lango graph import <file> [--output table|json]` command that reads a JSON array of triples from disk and imports them into the graph store.

#### Scenario: Graph import output uses the command writer
- **WHEN** `lango graph import` renders text or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Graph import empty file
- **WHEN** user runs `lango graph import <file>` and the JSON file contains an empty array
- **THEN** the command displays "No triples to import." and exits successfully

#### Scenario: Graph import JSON summary
- **WHEN** user runs `lango graph import <file> --output json`
- **THEN** the command outputs a JSON object containing the imported triple count

#### Scenario: Graph import rejects unknown output before file parsing
- **WHEN** user runs `lango graph import <file> --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT read or parse the import file

### Requirement: AllTriples method on Store interface
The graph.Store interface SHALL include an `AllTriples(ctx context.Context) ([]Triple, error)` method that returns every triple in the store. This method is required to support the graph export command.

#### Scenario: AllTriples on populated store
- **WHEN** AllTriples() is called on a store containing N triples
- **THEN** the method returns a slice of exactly N Triple values with no error

#### Scenario: AllTriples on empty store
- **WHEN** AllTriples() is called on an empty store
- **THEN** the method returns an empty slice with no error

### Requirement: BoltDB AllTriples implementation
The BoltDB-backed Store implementation SHALL implement AllTriples() by scanning the SPO index bucket and returning all triples.

#### Scenario: Full scan
- **WHEN** AllTriples() is called on a BoltDB store with triples
- **THEN** the implementation iterates the SPO bucket, decodes all entries, and returns the complete list

### Requirement: Backward compatibility
The addition of AllTriples() to the Store interface SHALL NOT change the behavior of any existing Store methods. All existing tests for QueryBySubject, QueryByObject, QueryBySubjectPredicate, Count, PredicateStats, and ClearAll SHALL continue to pass.

#### Scenario: Existing tests pass
- **WHEN** `go test ./internal/graph/...` is run after the interface addition
- **THEN** all existing tests pass without modification

### Requirement: Graph clear confirmation uses shared command streams
`lango graph clear` SHALL drive its confirmation prompt through the shared confirmation helper using Cobra command input/output streams.

#### Scenario: Graph clear aborts on denial
- **WHEN** `lango graph clear` prompts for confirmation and the operator answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the graph store unchanged

#### Scenario: Graph clear prompt uses command streams
- **WHEN** `lango graph clear` prompts for confirmation
- **THEN** the warning line and `Continue? [y/N]: ` prompt SHALL be written through the Cobra command output stream
- **AND** the operator response SHALL be read through the Cobra command input stream

### Requirement: Graph clear treats EOF as denial
`lango graph clear` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Graph clear EOF aborts cleanly
- **WHEN** `lango graph clear` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the graph store unchanged
