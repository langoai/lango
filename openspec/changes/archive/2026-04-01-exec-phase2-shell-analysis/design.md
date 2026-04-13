## Context

The current `unwrapShellWrapper()` in `internal/tools/exec/unwrap.go` uses string-based field splitting and manual quote extraction. It cannot handle login shell flags (`sh -lc`), env wrappers (`/usr/bin/env sh -c`), nested wrappers, or complex quoting/escaping — all of which are trivial reformulations that bypass the policy evaluator.

## Goals / Non-Goals

**Goals:**
- Replace string-based internals with AST parser (`mvdan.cc/sh/v3/syntax`) for robust shell command analysis
- Support login shell flags (`-lc`), env wrappers, and nested wrappers
- Recursive unwrap with depth limit 5
- Maintain backward compatibility — same `unwrapShellWrapper(cmd)` public signature
- Graceful fallback to string-based parser on AST parse failure

**Non-Goals:**
- Changing PolicyEvaluator's step ordering or adding new verdict types
- Supporting non-POSIX shell syntaxes
- Modifying the `CommandGuard` or opaque pattern detection

## Decisions

### Decision 1: AST parser via `mvdan.cc/sh/v3/syntax`
**Choice**: Use `mvdan.cc/sh/v3/syntax.NewParser().Parse()` to parse commands into `*syntax.File` AST, then walk the AST to find shell wrapper patterns.

**Rationale**: `mvdan.cc/sh/v3` is the standard Go shell parser. It handles quoting, escaping, nested structures, and POSIX semantics correctly. Alternative (regex/string splitting) cannot handle these correctly.

### Decision 2: Recursive unwrap with internal helper
**Choice**: Add `unwrapShellWrapperAST(file *syntax.File, depth int) (string, bool)` as internal recursive function. Public `unwrapShellWrapper()` calls this with depth=0.

**Rationale**: Separates AST walking (recursive) from entry point (parse + fallback). Depth limit 5 prevents infinite recursion on adversarial input.

### Decision 3: Fallback to string-based parser
**Choice**: On `syntax.Parse()` failure, fall back to existing string-based logic.

**Rationale**: Some malformed commands may fail AST parsing but still be parseable by the simpler string approach. Defense in depth — never lose coverage.

### Decision 4: Env wrapper detection in AST
**Choice**: When AST shows `CallExpr` with verb `env`, skip the `env` word and process remaining args as if they were the command. This handles `/usr/bin/env sh -c "cmd"`.

**Rationale**: `env` is a common prefix that the string-based parser couldn't handle. The AST makes it trivial to detect.

### Decision 5: Login shell flag support
**Choice**: When walking AST args, recognize `-lc`, `-ic`, etc. as containing the `-c` flag (combined POSIX short flags where `c` is the last character, since `-c` takes an argument).

**Rationale**: `sh -lc "cmd"` is a common pattern. The `-c` flag consumes the next argument regardless of preceding flags.

## Risks / Trade-offs

- **[Risk] AST parser adds ~2ms per parse** → Acceptable; exec policy evaluation is not a hot path
- **[Risk] `mvdan.cc/sh/v3` may not parse all edge cases** → Fallback to string-based parser covers gaps
- **[Risk] Recursive unwrap could be computationally expensive for deeply nested commands** → Depth limit of 5 caps worst case
