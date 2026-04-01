## Context

The exec policy evaluator (Phase 2) handles shell wrapper unwrap and opaque pattern detection, but does not cover heredocs, process substitution, grouped subshells, shell functions, `xargs cmd`, `find -exec cmd`, or bare env prefix (`VAR=val cmd`). These are common shell constructs that can hide dangerous inner commands.

## Goals / Non-Goals

**Goals:**
- Detect heredocs, process substitution, grouped subshells, and shell functions as opaque patterns (observe verdict)
- Extract inner verb from `xargs cmd` and `find -exec cmd`, evaluate with existing guard; observe if extraction fails
- Strip bare env prefix `VAR=val cmd` and evaluate the effective command
- All existing tests must continue to pass

**Non-Goals:**
- Deep analysis of heredoc content or function bodies (intentionally opaque)
- Modifying CommandGuard or the existing catastrophic pattern system
- Supporting non-POSIX shell syntax

## Decisions

### Decision 1: AST-based detection for shell constructs
**Choice**: Use `mvdan.cc/sh/v3/syntax` AST node types (`Redirect` with heredoc, `ProcSubst`, `Subshell`, `Block`, `FuncDecl`) for detection in `opaque.go`.

**Rationale**: The AST parser already parses these constructs. Type-checking AST nodes is more reliable than regex. Falls back to string-based detection when AST parsing fails.

### Decision 2: Hybrid approach for xargs and find-exec
**Choice**: String-based extraction for `xargs cmd` and `find -exec cmd` inner verbs. If extraction succeeds, evaluate the inner verb through existing guard checks. If extraction fails, return observe.

**Rationale**: These are command-line argument patterns, not shell syntax. String parsing is simpler and sufficient. The inner verb is the security-relevant part.

### Decision 3: Env prefix in unwrap.go
**Choice**: When AST CallExpr has `Assigns` (variable assignments before command), skip them and evaluate the remaining CallExpr args as the command. This handles `VAR=val cmd` without explicit `env`.

**Rationale**: `mvdan.cc/sh/v3/syntax.CallExpr` already separates `Assigns` from `Args`. This is a clean AST-based approach.

### Decision 4: New ReasonCodes for constructs
**Choice**: Add `ReasonHeredoc`, `ReasonProcessSubst`, `ReasonGroupedSubshell`, `ReasonShellFunction`, `ReasonXargsExtract`, `ReasonFindExecExtract` reason codes.

**Rationale**: Machine-readable reason codes enable downstream monitoring and policy decisions.

### Decision 5: Integration point in Evaluate flow
**Choice**: Shell construct detection runs AFTER existing opaque pattern detection (Step 5) but shares the same observe verdict path. For xargs/find-exec, extraction happens in a new step between shell wrapper unwrap (Step 1) and the existing evaluation steps.

**Rationale**: Preserves existing evaluation order. Constructs that are fully opaque get observe. Constructs with extractable inner verbs get evaluated through existing guard.

## Risks / Trade-offs

- **[Risk] AST detection may miss edge cases** → String-based fallback for critical patterns (heredoc, process substitution)
- **[Risk] xargs/find-exec extraction is heuristic** → Observe fallback on extraction failure
- **[Risk] Env prefix could conflict with existing env wrapper handling** → Env prefix in unwrap.go handles bare assignments; existing env wrapper handling in unwrapShellWrapperAST handles explicit `env` command
