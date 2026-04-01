## Purpose

Define the PolicyEvaluator system that evaluates exec tool commands with shell wrapper unwrapping, three-verdict policy decisions (allow/observe/block), opaque pattern detection, and outermost middleware integration.

## Requirements

### Requirement: PolicyEvaluator returns three-verdict decisions
The PolicyEvaluator SHALL evaluate shell commands and return a `PolicyDecision` with one of three verdicts: `VerdictAllow`, `VerdictObserve`, or `VerdictBlock`. Each decision SHALL include a `ReasonCode` and human-readable `Message`.

#### Scenario: Block kill verb through shell wrapper
- **WHEN** PolicyEvaluator evaluates `sh -c "kill 1234"`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonKillVerb`

#### Scenario: Block lango CLI through shell wrapper
- **WHEN** PolicyEvaluator evaluates `bash -c "lango security check"`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonLangoCLI`

#### Scenario: Block protected path (existing behavior preserved)
- **WHEN** PolicyEvaluator evaluates `sqlite3 ~/.lango/lango.db`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonProtectedPath`

#### Scenario: Observe opaque command substitution
- **WHEN** PolicyEvaluator evaluates `ls $(cat /etc/passwd)`
- **THEN** the decision SHALL be `VerdictObserve` with `ReasonCmdSubstitution`

#### Scenario: Allow clean command
- **WHEN** PolicyEvaluator evaluates `go build ./...`
- **THEN** the decision SHALL be `VerdictAllow` with `ReasonNone`

#### Scenario: Allow clean command through shell wrapper
- **WHEN** PolicyEvaluator evaluates `sh -c "go build ./..."`
- **THEN** the decision SHALL be `VerdictAllow` with `ReasonNone`

#### Scenario: Distinguish skill import from lango CLI
- **WHEN** PolicyEvaluator evaluates `bash -c "git clone https://github.com/org/repo skill-name"`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonSkillImport` (not `ReasonLangoCLI`)

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

### Requirement: Opaque pattern detection flags unsafe shell constructs
The PolicyEvaluator SHALL detect opaque shell patterns that prevent static analysis of the actual command intent. Detected patterns result in `VerdictObserve`.

#### Scenario: Detect command substitution with $(...)
- **WHEN** `detectOpaquePattern` receives `echo $(whoami)`
- **THEN** it SHALL return `ReasonCmdSubstitution`

#### Scenario: Detect command substitution with backticks
- **WHEN** `detectOpaquePattern` receives `` echo `whoami` ``
- **THEN** it SHALL return `ReasonCmdSubstitution`

#### Scenario: Detect unsafe variable expansion
- **WHEN** `detectOpaquePattern` receives `echo ${SECRET_TOKEN}`
- **THEN** it SHALL return `ReasonUnsafeVarExpand`

#### Scenario: Allow safe variable expansion
- **WHEN** `detectOpaquePattern` receives `echo ${HOME}/bin`
- **THEN** it SHALL return `ReasonNone` (HOME is in the safe variable set)

#### Scenario: Detect eval verb
- **WHEN** `detectOpaquePattern` receives `eval "dangerous stuff"`
- **THEN** it SHALL return `ReasonEvalVerb`

#### Scenario: Detect encoded pipe to shell
- **WHEN** `detectOpaquePattern` receives `base64 -d payload | bash`
- **THEN** it SHALL return `ReasonEncodedPipe`

### Requirement: WithPolicy middleware enforces policy before approval
The `WithPolicy` middleware SHALL be the outermost middleware in the tool chain, running before `WithApproval`. It SHALL only evaluate `exec` and `exec_bg` tools; all other tools pass through unchanged.

#### Scenario: Block verdict prevents approval prompt
- **WHEN** `WithPolicy` evaluates a command that results in `VerdictBlock`
- **THEN** it SHALL return `BlockedResult{Blocked: true, Message: msg}` without calling the next handler
- **AND** the approval provider SHALL NOT be invoked

#### Scenario: Observe verdict proceeds to approval
- **WHEN** `WithPolicy` evaluates a command that results in `VerdictObserve`
- **THEN** it SHALL log the decision and call the next handler (which includes approval)

#### Scenario: Allow verdict proceeds normally
- **WHEN** `WithPolicy` evaluates a command that results in `VerdictAllow`
- **THEN** it SHALL call the next handler without logging

#### Scenario: Non-exec tools pass through
- **WHEN** `WithPolicy` receives a tool invocation for `exec_status` or `exec_stop`
- **THEN** it SHALL call the next handler without any policy evaluation

### Requirement: PolicyDecisionEvent published for observe and block verdicts
The PolicyEvaluator SHALL publish a `PolicyDecisionEvent` to the event bus for `VerdictObserve` and `VerdictBlock` decisions. Events SHALL NOT be published for `VerdictAllow`. Publishing SHALL be skipped when the event bus is nil (respecting `cfg.Hooks.EventPublishing` gate).

#### Scenario: Block verdict publishes event
- **WHEN** PolicyEvaluator blocks a command and bus is non-nil
- **THEN** a `PolicyDecisionEvent` SHALL be published with verdict="block", the reason code, and command details

#### Scenario: Observe verdict publishes event
- **WHEN** PolicyEvaluator observes a command and bus is non-nil
- **THEN** a `PolicyDecisionEvent` SHALL be published with verdict="observe"

#### Scenario: No event when bus is nil
- **WHEN** PolicyEvaluator makes any decision and bus is nil
- **THEN** no event SHALL be published

### Requirement: PolicyEvaluator blocks catastrophic patterns before approval
The PolicyEvaluator SHALL check ALL commands against a catastrophic pattern set as step 4 in the Evaluate flow, after CommandGuard (step 3) and before opaque detection (step 5). This check runs for ALL commands regardless of opaque status. Commands matching any catastrophic pattern SHALL receive `VerdictBlock` with `ReasonCatastrophicPattern`.

#### Scenario: Block direct catastrophic command
- **WHEN** PolicyEvaluator evaluates `rm -rf /`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonCatastrophicPattern`

#### Scenario: Block opaque command containing catastrophic pattern
- **WHEN** PolicyEvaluator evaluates `eval "rm -rf /"`
- **THEN** the decision SHALL be `VerdictBlock` with `ReasonCatastrophicPattern`

#### Scenario: Allow opaque non-catastrophic command
- **WHEN** PolicyEvaluator evaluates `eval "echo hello"`
- **THEN** the decision SHALL be `VerdictObserve` with `ReasonEvalVerb`

### Requirement: NewPolicyEvaluator accepts functional options
`NewPolicyEvaluator(guard, classifier, bus, opts...)` SHALL accept variadic `Option` functions. `WithCatastrophicPatterns(patterns)` SHALL set the catastrophic pattern set with dedupe and lowercase normalization matching `NewSecurityFilterHook` semantics. The safe variable set SHALL be initialized from internal defaults.

#### Scenario: Constructor with catastrophic patterns option
- **WHEN** `NewPolicyEvaluator` is called with `WithCatastrophicPatterns(patterns)`
- **THEN** the evaluator SHALL have pre-lowercased, deduped catastrophic patterns

#### Scenario: Constructor without options
- **WHEN** `NewPolicyEvaluator` is called without options
- **THEN** the catastrophic pattern set SHALL be empty and the check SHALL pass all commands through

#### Scenario: Constructor provides default safe vars
- **WHEN** `NewPolicyEvaluator` is called with guard, classifier, and bus
- **THEN** the returned evaluator SHALL have the safe variable set pre-populated with 10 known-safe environment variables
