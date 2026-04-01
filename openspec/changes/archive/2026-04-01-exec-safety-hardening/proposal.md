## Why

The `exec`/`exec_bg` tool guards can be bypassed by wrapping commands in shell wrappers (`sh -c "kill 1234"`, `bash -c "lango security check"`). The kill-verb guard relies on `extractVerb` which returns `"sh"` instead of the inner verb, and the lango CLI guard uses prefix matching that fails when the command starts with `bash -c`. Additionally, the current guard runs inside the handler (after the approval middleware), so users are prompted for approval on commands that will be blocked anyway.

## What Changes

- Add a `PolicyEvaluator` that unwraps one level of shell wrappers (`sh -c`, `bash -c`, `/bin/sh -c`) and applies existing guard checks to the inner command
- Introduce a three-verdict system (`allow`, `observe`, `block`) with machine-readable reason codes, replacing the current binary allow/block string return
- Move policy enforcement from handler-level guards to an outermost middleware (`WithPolicy`), ensuring blocked commands never reach the approval prompt
- Add `PolicyDecisionEvent` to the event bus for observe/block verdicts (respecting `cfg.Hooks.EventPublishing` gate)
- Refactor `blockLangoExec` into `classifyLangoExec` returning structured `(message, ReasonCode)` to distinguish `ReasonLangoCLI` from `ReasonSkillImport`
- Detect opaque shell patterns (command substitution, unsafe variable expansion, `eval`, encoded pipes) and flag them as `observe` verdicts for monitoring

## Capabilities

### New Capabilities
- `exec-policy-evaluator`: Shell wrapper unwrap, three-verdict policy evaluation (allow/observe/block), reason codes, and outermost middleware integration

### Modified Capabilities
- `exec-command-guard`: Add shell wrapper unwrap detection and observe verdict for opaque patterns
- `tool-exec`: Change guard execution from handler-level to outermost middleware (`WithPolicy → WithApproval → ... → Handler`), update execution order semantics

## Impact

- **Code**: `internal/tools/exec/` (4 new files), `internal/eventbus/events.go`, `internal/app/tools.go`, `internal/app/app.go`
- **Middleware chain**: New outermost middleware changes execution order to `Policy → Approval → Hooks → Handler`
- **Prompts**: `TOOL_USAGE.md`, `SAFETY.md` updated to reflect shell wrapper enforcement
- **No breaking changes**: Tool names, parameter shapes, `BlockedResult` JSON shape, CLI output all preserved
- **No new dependencies**: String-based analysis, no shell parser library
