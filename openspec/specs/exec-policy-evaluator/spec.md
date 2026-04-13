## Purpose

Delta spec for env wrapper unwrap improvements in the policy evaluator.

## Requirements

### Requirement: Env wrapper unwrap handles flags and variable assignments

The shell wrapper unwrap SHALL skip env-specific arguments before reaching the shell verb. Supported env arguments:
- Standalone flags: `-i`, `-0`
- Flag with argument: `-u NAME`, `-C DIR`, `-S STRING`
- Terminator: `--`
- Variable assignments: `NAME=value` where NAME matches shell variable name pattern (alpha/underscore start, alphanumeric/underscore body)

Invalid assignments (paths like `./foo=bar`, flags like `--flag=val`) SHALL NOT be treated as env assignments.

#### Scenario: Env with variable assignment before shell verb
- **WHEN** unwrap receives `env FOO=1 sh -c "kill 1234"`
- **THEN** it SHALL return inner command `kill 1234` and `unwrapped=true`

#### Scenario: Env with standalone flag before shell verb
- **WHEN** unwrap receives `env -i bash -c "lango cron"`
- **THEN** it SHALL return inner command `lango cron` and `unwrapped=true`

#### Scenario: Env with flag-argument pair before shell verb
- **WHEN** unwrap receives `env -u SECRET sh -c "echo hi"`
- **THEN** it SHALL return inner command `echo hi` and `unwrapped=true`
- **AND** `SECRET` SHALL NOT be mistaken for the command verb

#### Scenario: Env with -S split-string flag
- **WHEN** unwrap receives `env -S "FOO=1 BAR=2" sh -c "echo test"`
- **THEN** it SHALL return inner command `echo test` and `unwrapped=true`

#### Scenario: Env with terminator before shell verb
- **WHEN** unwrap receives `env -- sh -c "kill 1"`
- **THEN** it SHALL return inner command `kill 1` and `unwrapped=true`

#### Scenario: Path-like string rejected as env assignment
- **WHEN** unwrap receives `env ./foo=bar`
- **THEN** it SHALL NOT treat `./foo=bar` as an env assignment
