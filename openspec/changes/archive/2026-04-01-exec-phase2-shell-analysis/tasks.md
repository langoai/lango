## 1. AST-based unwrap implementation

- [x] 1.1 Add `mvdan.cc/sh/v3/syntax` import and `unwrapShellWrapperAST(file *syntax.File, depth int) (string, bool)` internal function in `unwrap.go`
- [x] 1.2 Implement AST walking: detect `CallExpr` with shell verb + `-c` flag, extract inner command
- [x] 1.3 Support login shell flags (`-lc`, `-ic`) — recognize combined flags where `c` is last
- [x] 1.4 Support env wrapper (`/usr/bin/env sh -c "cmd"`) — skip env verb and process remaining args
- [x] 1.5 Implement recursive unwrap: re-parse extracted inner command, recurse with depth+1, limit at depth 5
- [x] 1.6 Update `unwrapShellWrapper()` to call AST parser first, fallback to string-based on parse failure

## 2. Test coverage

- [x] 2.1 Add test cases for `sh -lc "kill 1234"` login shell unwrap
- [x] 2.2 Add test cases for `/usr/bin/env sh -c "echo hello"` env wrapper
- [x] 2.3 Add test cases for nested `sh -c "bash -c \"inner\""` recursive unwrap
- [x] 2.4 Add test case for depth limit exceeded (6 levels nested) returning original
- [x] 2.5 Verify all existing test cases still pass unchanged

## 3. Policy integration and downstream

- [x] 3.1 Add policy_test.go cases for newly supported wrapper forms (login shell, env wrapper, nested)
- [x] 3.2 Update `prompts/SAFETY.md` to mention newly supported wrapper forms
- [x] 3.3 Update `openspec/specs/exec-policy-evaluator/spec.md` with new scenarios

## 4. Verification

- [x] 4.1 Run `go build ./...` — verify no build errors
- [x] 4.2 Run `go test ./internal/tools/exec/... -v` — all tests pass
