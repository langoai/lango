## Purpose

Define the CommandGuard system that blocks exec tool commands from accessing protected data paths and executing dangerous process management operations.

## Requirements

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

### Requirement: PolicyEvaluator unwraps shell wrappers before guard checks
The PolicyEvaluator SHALL detect one level of shell wrapper (`sh -c`, `bash -c`, `zsh -c`, `dash -c` with optional path prefix) and apply all guard checks to the inner command. Patterns not supported in Phase 1: `sh -lc`, `/usr/bin/env sh -c`, nested wrappers, escaped quotes.

#### Scenario: Block kill verb through shell wrapper
- **WHEN** PolicyEvaluator evaluates `sh -c "kill 1234"`
- **THEN** the inner command `kill 1234` is extracted and blocked by the process management guard

#### Scenario: Block lango CLI through shell wrapper
- **WHEN** PolicyEvaluator evaluates `bash -c "lango security check"`
- **THEN** the inner command is extracted and blocked by the lango CLI guard

#### Scenario: Allow clean command through shell wrapper
- **WHEN** PolicyEvaluator evaluates `sh -c "go build ./..."`
- **THEN** the inner command is extracted and allowed

### Requirement: PolicyEvaluator observes opaque shell patterns
The PolicyEvaluator SHALL detect opaque shell patterns (command substitution, unsafe variable expansion, eval verb, encoded pipes) and return an observe verdict. Observe verdicts allow execution but log the decision and publish an event.

#### Scenario: Observe command substitution
- **WHEN** PolicyEvaluator evaluates `ls $(cat /etc/passwd)`
- **THEN** the verdict is observe with reason `cmd_substitution`

#### Scenario: Safe variable expansion not flagged
- **WHEN** PolicyEvaluator evaluates `echo ${HOME}/bin`
- **THEN** the verdict is allow (HOME is in the safe variable set)

### Requirement: classifyLangoExec returns structured reason codes
The `classifyLangoExec` function SHALL return `(message string, reason ReasonCode)` distinguishing lango CLI invocation (`ReasonLangoCLI`) from skill import redirects (`ReasonSkillImport`).

#### Scenario: Skill import classified separately from lango CLI
- **WHEN** `classifyLangoExec` receives `git clone https://github.com/org/repo skill-name`
- **THEN** it returns a non-empty message and `ReasonSkillImport`

### Requirement: CommandGuard uses pre-built Replacer
The system SHALL pre-build a `strings.Replacer` at construction time for `$HOME`, `${HOME}`, and tilde-at-word-boundary patterns. The `normalizeCommand` method SHALL use this replacer for single-pass replacement.

#### Scenario: Efficient normalization
- **WHEN** a command contains both `$HOME` and `~/` references
- **THEN** the replacer handles all substitutions without multiple string scans
