## MODIFIED Requirements

### Requirement: Shell wrapper unwrap detects one level of sh/bash -c
The PolicyEvaluator SHALL detect and unwrap one level of shell wrapper commands. Supported wrapper verbs are `sh`, `bash`, `zsh`, `dash` (with optional path prefix). The unwrap SHALL extract only the **first argument** after the `-c` flag following POSIX semantics (`sh -c command_string [command_name [argument...]]`). For quoted arguments, it SHALL extract content between matching quotes. For unquoted arguments, it SHALL extract the first whitespace-delimited token. Unmatched quotes SHALL cause the unwrap to fail (return original command).

#### Scenario: Unwrap sh -c with double quotes
- **WHEN** `unwrapShellWrapper` receives `sh -c "kill 1234"`
- **THEN** it SHALL return inner command `kill 1234` and `unwrapped=true`

#### Scenario: Unwrap path-prefixed bash -c
- **WHEN** `unwrapShellWrapper` receives `/usr/bin/bash -c "echo hello"`
- **THEN** it SHALL return inner command `echo hello` and `unwrapped=true`

#### Scenario: Unwrap ignores positional arguments after quoted command
- **WHEN** `unwrapShellWrapper` receives `bash -c "kill 1234" ignored`
- **THEN** it SHALL return inner command `kill 1234` and `unwrapped=true`
- **AND** the `ignored` positional argument SHALL NOT be part of the inner command

#### Scenario: Unwrap ignores positional arguments after single-quoted command
- **WHEN** `unwrapShellWrapper` receives `sh -c 'lango cron' myname`
- **THEN** it SHALL return inner command `lango cron` and `unwrapped=true`

#### Scenario: Unquoted unwrap extracts first token only
- **WHEN** `unwrapShellWrapper` receives `bash -c echo foo bar`
- **THEN** it SHALL return inner command `echo` and `unwrapped=true`

#### Scenario: Unmatched quote returns original
- **WHEN** `unwrapShellWrapper` receives `sh -c "kill 1234`
- **THEN** it SHALL return the original command and `unwrapped=false`

#### Scenario: No unwrap for non-shell verb
- **WHEN** `unwrapShellWrapper` receives `python3 -c "print('hi')"`
- **THEN** it SHALL return the original command and `unwrapped=false`

#### Scenario: No unwrap for sh without -c flag
- **WHEN** `unwrapShellWrapper` receives `sh script.sh`
- **THEN** it SHALL return the original command and `unwrapped=false`

#### Scenario: No unwrap for login shell flag
- **WHEN** `unwrapShellWrapper` receives `sh -lc "cmd"`
- **THEN** it SHALL return the original command and `unwrapped=false`

#### Scenario: No unwrap for env wrapper
- **WHEN** `unwrapShellWrapper` receives `/usr/bin/env sh -c "cmd"`
- **THEN** it SHALL return the original command and `unwrapped=false`

## ADDED Requirements

### Requirement: PolicyEvaluator blocks catastrophic patterns before approval
The PolicyEvaluator SHALL check ALL commands against a catastrophic pattern set before opaque detection. This check runs as step 4 in the Evaluate flow, after CommandGuard (step 3) and before opaque detection (step 5). Commands matching any catastrophic pattern SHALL receive `VerdictBlock` with `ReasonCatastrophicPattern`.

#### Scenario: Block direct catastrophic command
- **WHEN** PolicyEvaluator evaluates `rm -rf /`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonCatastrophicPattern`

#### Scenario: Block opaque command containing catastrophic pattern
- **WHEN** PolicyEvaluator evaluates `eval "rm -rf /"`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonCatastrophicPattern`
- **AND** the opaque detection step SHALL NOT be reached

#### Scenario: Block command substitution with catastrophic pattern
- **WHEN** PolicyEvaluator evaluates `echo $(mkfs.ext4 /dev/sda)`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonCatastrophicPattern`

#### Scenario: Allow opaque non-catastrophic command
- **WHEN** PolicyEvaluator evaluates `eval "echo hello"`
- **THEN** the decision SHALL be `VerdictObserve` with `ReasonEvalVerb`

#### Scenario: Block user-configured blocked pattern
- **WHEN** PolicyEvaluator is configured with a user-defined blocked pattern `"drop table"`
- **AND** evaluates `exec "drop table users"`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonCatastrophicPattern`

### Requirement: NewPolicyEvaluator accepts functional options
`NewPolicyEvaluator` SHALL accept variadic `Option` functions after the required parameters. `WithCatastrophicPatterns(patterns)` SHALL set the catastrophic pattern set with dedupe and lowercase normalization matching `NewSecurityFilterHook` semantics.

#### Scenario: Constructor with catastrophic patterns option
- **WHEN** `NewPolicyEvaluator` is called with `WithCatastrophicPatterns(patterns)`
- **THEN** the evaluator SHALL have pre-lowercased, deduped catastrophic patterns

#### Scenario: Constructor without options preserves defaults
- **WHEN** `NewPolicyEvaluator` is called without options
- **THEN** the evaluator SHALL have no catastrophic patterns (empty slice)
- **AND** the catastrophic check step SHALL pass all commands through
