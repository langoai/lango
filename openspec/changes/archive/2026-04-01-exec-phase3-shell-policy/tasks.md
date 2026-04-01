## 1. Opaque pattern detection for new constructs

- [x] 1.1 Add `ReasonHeredoc`, `ReasonProcessSubst`, `ReasonGroupedSubshell`, `ReasonShellFunction` reason codes in `policy.go`
- [x] 1.2 Add AST-based `detectShellConstruct(cmd string) ReasonCode` in `opaque.go` that parses the command and checks for heredoc (`Redirect` with `<<`/`<<<`), process substitution (`ProcSubst`), grouped subshell (`Subshell`), brace group (`Block`), and shell function declarations (`FuncDecl`)
- [x] 1.3 Add string-based fallback detection in `detectShellConstruct` for heredoc (`<<`), process substitution (`<(`, `>(`), and function definition (`() {`) patterns when AST parsing fails
- [x] 1.4 Add opaque_test.go tests for each new construct pattern

## 2. Inner verb extraction for xargs and find-exec

- [x] 2.1 Add `extractXargsVerb(cmd string) (string, bool)` in `unwrap.go` — extract verb from `xargs [-flags] cmd [args]` pattern
- [x] 2.2 Add `extractFindExecVerb(cmd string) (string, bool)` in `unwrap.go` — extract verb from `find ... -exec cmd {} \;` or `-exec cmd {} +` pattern
- [x] 2.3 Add unwrap_test.go tests for xargs and find-exec verb extraction

## 3. Env prefix handling in unwrap

- [x] 3.1 Add `unwrapEnvPrefix(cmd string) (string, bool)` in `unwrap.go` — strip `VAR=val` prefixes from command, return remaining command
- [x] 3.2 Integrate env prefix unwrap into `unwrapShellWrapperAST` by checking `CallExpr.Assigns` field (already handled: AST separates Assigns from Args, so shell wrapper detection sees the verb correctly)
- [x] 3.3 Add unwrap_test.go tests for env prefix stripping

## 4. Policy integration

- [x] 4.1 Integrate `detectShellConstruct` into `Evaluate` flow — call after existing opaque detection (Step 5), return observe verdict with appropriate reason
- [x] 4.2 Integrate xargs/find-exec extraction into `Evaluate` flow — extract inner verb, run through guard checks; observe if extraction fails
- [x] 4.3 Integrate env prefix unwrap into `Evaluate` flow — strip env prefix before guard evaluation
- [x] 4.4 Add policy_test.go tests for all new constructs through full Evaluate path

## 5. Documentation and verification

- [x] 5.1 Update `prompts/SAFETY.md` to document new construct handling
- [x] 5.2 Run `go build ./...` — verify no build errors
- [x] 5.3 Run `go test ./internal/tools/exec/... -v` — all tests pass
