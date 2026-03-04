# CLI Graph Extended

## Purpose
Provides extended CLI commands for the graph store, including adding individual triples, exporting all triples, and importing triples from files.

## Requirements

### Requirement: Graph add command
The system SHALL provide a `lango graph add --subject <s> --predicate <p> --object <o> [--json]` command that adds a single triple to the graph store. The command SHALL use cfgLoader combined with initGraphStore() to initialize the graph backend. All three flags (subject, predicate, object) MUST be provided.

#### Scenario: Successful add
- **WHEN** user runs `lango graph add --subject "entity1" --predicate "related_to" --object "entity2"`
- **THEN** system adds the triple and prints "Triple added: entity1 --[related_to]--> entity2"

#### Scenario: Missing required flag
- **WHEN** user runs `lango graph add --subject "entity1"` without predicate or object
- **THEN** system returns an error indicating the missing required flags

#### Scenario: Graph disabled
- **WHEN** user runs `lango graph add` with graph.enabled set to false
- **THEN** system returns error "Graph store is not enabled"

### Requirement: Graph export command
The system SHALL provide a `lango graph export [--format json|csv] [--output <file>]` command that exports all triples from the graph store. The default format SHALL be JSON. When `--output` is provided, the command SHALL write to the specified file; otherwise it SHALL write to stdout. The command SHALL call AllTriples() on the graph.Store interface.

#### Scenario: Export to JSON stdout
- **WHEN** user runs `lango graph export`
- **THEN** system outputs all triples as a JSON array to stdout

#### Scenario: Export to CSV file
- **WHEN** user runs `lango graph export --format csv --output triples.csv`
- **THEN** system writes all triples as CSV (subject,predicate,object) to triples.csv

#### Scenario: Empty graph
- **WHEN** user runs `lango graph export` with no triples in the store
- **THEN** system outputs an empty JSON array "[]"

### Requirement: Graph import command
The system SHALL provide a `lango graph import <file> [--format json|csv]` command that imports triples from a file into the graph store. The default format SHALL be JSON. The command SHALL use AddTriples() for atomic batch insertion.

#### Scenario: Import from JSON file
- **WHEN** user runs `lango graph import triples.json`
- **THEN** system reads the file, parses triples, and adds them to the store, printing "Imported N triples"

#### Scenario: Import from CSV file
- **WHEN** user runs `lango graph import triples.csv --format csv`
- **THEN** system reads the CSV file with subject,predicate,object columns and imports the triples

#### Scenario: Invalid file format
- **WHEN** user runs `lango graph import malformed.json`
- **THEN** system returns error indicating the file could not be parsed

### Requirement: Graph extended commands registration
The add, export, and import subcommands SHALL be registered under the existing `lango graph` command group.

#### Scenario: Graph help lists new subcommands
- **WHEN** user runs `lango graph --help`
- **THEN** the help output includes add, export, and import alongside existing subcommands
