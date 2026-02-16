## Purpose

Define the CLI commands for inspecting, querying, and managing the knowledge graph store.

## Requirements

### Requirement: Graph status command
The system SHALL provide a `lango graph status` command that displays the graph store configuration and triple count. The command SHALL support a `--json` flag for machine-readable output.

#### Scenario: Graph disabled
- **WHEN** user runs `lango graph status` with graph.enabled=false
- **THEN** system displays that graph store is not enabled

#### Scenario: Graph enabled with JSON output
- **WHEN** user runs `lango graph status --json` with graph.enabled=true
- **THEN** system outputs JSON with enabled, backend, database_path, and triple_count fields

### Requirement: Graph query command
The system SHALL provide a `lango graph query` command that queries triples by subject, object, or subject+predicate. At least one of `--subject` or `--object` MUST be provided. The `--predicate` flag requires `--subject`. A `--limit` flag SHALL cap results. A `--json` flag SHALL enable JSON output.

#### Scenario: Query by subject
- **WHEN** user runs `lango graph query --subject "entity1"`
- **THEN** system displays matching triples in SUBJECT/PREDICATE/OBJECT tabwriter format

#### Scenario: No filter provided
- **WHEN** user runs `lango graph query` without --subject or --object
- **THEN** system returns an error indicating at least one filter is required

### Requirement: Graph stats command
The system SHALL provide a `lango graph stats` command that displays total triple count and per-predicate breakdown sorted by count descending. The command SHALL support a `--json` flag.

#### Scenario: Stats with data
- **WHEN** user runs `lango graph stats` with populated graph
- **THEN** system displays total triple count and PREDICATE/COUNT table

### Requirement: Graph clear command
The system SHALL provide a `lango graph clear` command that removes all triples from the graph store. The command SHALL prompt for confirmation unless `--force` is provided.

#### Scenario: Clear with confirmation
- **WHEN** user runs `lango graph clear` and confirms with "y"
- **THEN** system clears all triples and prints confirmation message

#### Scenario: Clear aborted
- **WHEN** user runs `lango graph clear` and does not confirm
- **THEN** system prints "Aborted." and makes no changes

#### Scenario: Force clear
- **WHEN** user runs `lango graph clear --force`
- **THEN** system clears all triples without prompting
