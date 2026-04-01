## MODIFIED Requirements

### Requirement: Shell wrapper unwrap detects one level of sh/bash -c
The PolicyEvaluator SHALL detect and unwrap shell wrapper commands using AST-based parsing via `mvdan.cc/sh/v3/syntax`. Supported wrapper verbs are `sh`, `bash`, `zsh`, `dash` (with optional path prefix). The unwrap SHALL support login shell flags (`-lc`, `-ic`), env wrappers (`/usr/bin/env sh -c`), and recursive unwrap of nested wrappers with a depth limit of 5. The unwrap SHALL extract only the **first argument** after the `-c` flag following POSIX semantics. For quoted arguments, it SHALL extract content between matching quotes. For unquoted arguments, it SHALL extract the first whitespace-delimited token. Unmatched quotes SHALL cause the unwrap to fail (return original command). On AST parse failure, the unwrap SHALL fall back to the string-based parser.

#### Scenario: Unwrap sh -c with double quotes
- **WHEN** `unwrapShellWrapper` receives `sh -c "kill 1234"`
- **THEN** it SHALL return inner command `kill 1234` and `unwrapped=true`

#### Scenario: Unwrap path-prefixed bash -c
- **WHEN** `unwrapShellWrapper` receives `/usr/bin/bash -c "echo hello"`
- **THEN** it SHALL return inner command `echo hello` and `unwrapped=true`

#### Scenario: Unwrap ignores positional arguments after quoted command
- **WHEN** `unwrapShellWrapper` receives `bash -c "kill 1234" ignored`
- **THEN** it SHALL return inner command `kill 1234` and `unwrapped=true`

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

#### Scenario: Unwrap login shell with -lc flag
- **WHEN** `unwrapShellWrapper` receives `sh -lc "kill 1234"`
- **THEN** it SHALL return inner command `kill 1234` and `unwrapped=true`

#### Scenario: Unwrap env wrapper
- **WHEN** `unwrapShellWrapper` receives `/usr/bin/env sh -c "echo hello"`
- **THEN** it SHALL return inner command `echo hello` and `unwrapped=true`

#### Scenario: Recursive unwrap of nested wrappers
- **WHEN** `unwrapShellWrapper` receives `sh -c "bash -c \"inner\""`
- **THEN** it SHALL return inner command `inner` and `unwrapped=true`

#### Scenario: Depth limit exceeded returns original
- **WHEN** `unwrapShellWrapper` receives a command with more than 5 levels of nested shell wrappers
- **THEN** it SHALL return the original command and `unwrapped=false`

#### Scenario: AST parse failure falls back to string parser
- **WHEN** `unwrapShellWrapper` receives a command that fails AST parsing but matches the string-based pattern
- **THEN** it SHALL return the string-based unwrap result
