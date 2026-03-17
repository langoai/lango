## ADDED Requirements

### Requirement: CommandGuard blocks access to protected data paths
The system SHALL block exec commands that reference any protected data path. Path matching SHALL normalize `~`, `$HOME`, and `${HOME}` to the actual home directory before checking. The guard SHALL be constructed with a list of protected paths resolved to absolute form at creation time.

#### Scenario: Block sqlite3 access to lango database
- **WHEN** agent executes `sqlite3 ~/.lango/lango.db`
- **THEN** the command is blocked with a message indicating the protected path and suggesting built-in tools

#### Scenario: Block cat access with $HOME variant
- **WHEN** agent executes `cat $HOME/.lango/keyfile`
- **THEN** the command is blocked with the same protection

#### Scenario: Block access in piped commands
- **WHEN** agent executes `cat ~/.lango/keyfile | base64`
- **THEN** the command is blocked because the protected path appears in the normalized command

#### Scenario: Allow access to non-protected paths
- **WHEN** agent executes `sqlite3 /tmp/test.db`
- **THEN** the command is allowed through

### Requirement: CommandGuard blocks process management commands
The system SHALL block commands where the first verb is `kill`, `pkill`, or `killall`. Verb extraction SHALL strip path prefixes (e.g., `/usr/bin/kill` → `kill`) and be case-insensitive.

#### Scenario: Block kill command
- **WHEN** agent executes `kill 1`
- **THEN** the command is blocked with a message suggesting exec_stop for background processes

#### Scenario: Allow kill as argument
- **WHEN** agent executes `grep kill log.txt`
- **THEN** the command is allowed because "kill" is not the verb

### Requirement: CommandGuard uses pre-built Replacer
The system SHALL pre-build a `strings.Replacer` at construction time for `$HOME`, `${HOME}`, and tilde-at-word-boundary patterns. The `normalizeCommand` method SHALL use this replacer for single-pass replacement.

#### Scenario: Efficient normalization
- **WHEN** a command contains both `$HOME` and `~/` references
- **THEN** the replacer handles all substitutions without multiple string scans
