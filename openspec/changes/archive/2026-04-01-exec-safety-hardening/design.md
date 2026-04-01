## Context

The `exec`/`exec_bg` tool guards protect against dangerous commands (kill verbs, protected path access, lango CLI invocation). Two structural gaps exist:

1. **Shell wrapper bypass**: `sh -c "kill 1234"` and `bash -c "lango security"` bypass verb extraction and prefix matching respectively.
2. **Approval-before-policy**: Guards run inside the handler, but `WithApproval` middleware wraps outside — so users are prompted for approval on commands that will be blocked anyway.

Current middleware execution order (`app.go:145-209`):
```
WithApproval → WithPrincipal → WithHooks → WithOutputManager → WithLearning → Handler(checkGuards → execute)
```

## Goals / Non-Goals

**Goals:**
- Close shell wrapper bypass for kill verbs and lango CLI invocation
- Enforce policy → approval → execution ordering via outermost middleware
- Add observe verdict for opaque command patterns (monitoring without blocking)
- Add structured reason codes for policy decisions
- Publish policy decision events through the existing eventbus (respecting config gates)

**Non-Goals:**
- Full shell parser (no `mvdan.cc/sh` dependency)
- Support for `sh -lc`, `/usr/bin/env sh -c`, nested wrappers, or escaped quotes inside wrappers
- New config flags or CLI surface changes
- Changes to `BlockedResult` JSON shape, tool names, or parameter shapes
- Session cost restore or startup instrumentation (deferred)

## Decisions

### D1: Policy check as outermost middleware (not handler guard or hook)

**Decision**: Create `WithPolicy` middleware in `internal/tools/exec/middleware.go`, applied after `WithApproval` in `app.go` (making it outermost = first to execute).

**Alternatives considered**:
- *Handler guard (status quo)*: Cannot run before approval. User gets approval prompt then block message.
- *PreToolHook*: Hooks run inside `WithHooks` middleware, which is inside `WithApproval`. Same ordering problem.
- *Separate middleware inside approval*: Would require changing `WithApproval` internals.

**Rationale**: Outermost middleware is the only position that runs before approval. `WithPolicy` returns `BlockedResult` (not error) for consistency with existing handler guard behavior.

### D2: Handler guards preserved as defense-in-depth (strict subset)

**Decision**: Keep existing `langoGuard` and `pathGuard` in `BuildTools` handlers unchanged.

**Responsibility split**:
- **Shared by both**: Deterministic lango CLI / skill-import classification (`classifyLangoExec` → `blockLangoExec` delegation). New deterministic rules in the classifier flow to both paths.
- **Middleware only**: Shell wrapper unwrap, opaque pattern detection, observe verdict. Never added to handler guards.

**Invariant**: Any command blocked by handler guards is also blocked by the middleware.

### D3: Classifier returns `(string, ReasonCode)` — no cross-package type

**Decision**: `PolicyEvaluator.langoClassifier` has signature `func(cmd string) (message string, reason ReasonCode)` using only primitive + exec-package types.

**Rationale**: `internal/tools/exec` cannot import `internal/app`. Using a closure with primitive return types avoids reverse dependency. `classifyLangoExec` lives in `app/tools.go` and is injected as closure in `app.go` Phase B.

### D4: PolicyEvaluator created in `app.go` Phase B

**Decision**: Create PolicyEvaluator in `app.go` Phase B using `fv.CmdGuard` and `fv.AutoAvail` from `foundationValues` (already resolved via `resolver.Resolve`).

**Rationale**: No changes to `foundationModule`, `foundationValues`, or module interfaces. All dependencies (`bus`, `cfg`, `fv`) are available in Phase B scope.

### D5: EventBus publishing respects `cfg.Hooks.EventPublishing` gate

**Decision**: Pass `bus` as nil when `cfg.Hooks.EventPublishing` is false. PolicyEvaluator checks nil before publishing.

**Rationale**: Existing `EventBusHook` registration uses the same gate (`app.go:437`). Direct bus injection that bypasses this gate would break config semantics.

### D6: String-based shell unwrap (no parser library)

**Decision**: Detect `sh -c`, `bash -c`, `zsh -c`, `dash -c` wrappers (with optional path prefix) using string operations. Extract inner command after `-c` flag, strip outer quotes.

**Rationale**: Phase 1 scope is limited to one-level unwrap of known patterns. A full parser adds dependency weight for a narrow use case. The opaque pattern detector catches commands that can't be statically analyzed.

### D7: Constructor initializes safeVars internally

**Decision**: `NewPolicyEvaluator(guard, classifier, bus)` initializes the safe variable set (`HOME`, `PATH`, `USER`, `PWD`, `SHELL`, `TERM`, `LANG`, `LC_ALL`, `LC_CTYPE`, `TMPDIR`) as an internal default. No caller-provided option.

## Risks / Trade-offs

- **[False positives on opaque detection]** → `$PATH` and `$HOME` are extremely common. Mitigation: 10-variable safe set covers the most common cases. Observe doesn't block, only logs.
- **[Shell wrapper bypass for unsupported patterns]** → `sh -lc`, `env sh -c` not detected. Mitigation: Phase 1 covers the most common bypass patterns. SecurityFilterHook provides additional coverage for catastrophic commands.
- **[Observe vs downstream blocking confusion]** → `eval "rm -rf /"` is Observe at evaluator level but blocked by SecurityFilterHook. Mitigation: Documented as "Observe = evaluator passes through; downstream layers may still block independently."
- **[Dual guard execution overhead]** → Middleware and handler both check commands. Mitigation: String operations are negligible. Defense-in-depth value outweighs the microsecond cost.
