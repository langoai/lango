## 1. Command Guard Core

- [x] 1.1 Create `internal/tools/exec/guard.go` with CommandGuard struct, NewCommandGuard, CheckCommand
- [x] 1.2 Implement normalizeCommand with pre-built strings.Replacer for $HOME/${HOME}/~ substitution
- [x] 1.3 Implement extractVerb for first-word extraction with path stripping and lowercasing
- [x] 1.4 Implement expandAndAbs for path resolution at construction time
- [x] 1.5 Create `internal/tools/exec/guard_test.go` with table-driven tests for all scenarios

## 2. SecurityFilterHook Default Patterns

- [x] 2.1 Add DefaultBlockedPatterns() returning catastrophic command patterns
- [x] 2.2 Update NewSecurityFilterHook to merge defaults with user patterns (case-insensitive dedup)
- [x] 2.3 Pre-lowercase all patterns at construction time, use blockedPatternsLower in Pre()
- [x] 2.4 Add tests for default patterns, merging, dedup, and blocking behavior

## 3. Exec Handler Integration

- [x] 3.1 Define BlockedResult struct with Blocked/Message fields and JSON tags
- [x] 3.2 Add blockProtectedPaths helper in tools.go delegating to CommandGuard
- [x] 3.3 Wire blockProtectedPaths into exec handler after blockLangoExec
- [x] 3.4 Wire blockProtectedPaths into exec_bg handler after blockLangoExec
- [x] 3.5 Replace all map[string]interface{}{"blocked":...} with &BlockedResult{...}
- [x] 3.6 Add tests for blockProtectedPaths with guard and nil guard cases

## 4. Always-On Security Hook

- [x] 4.1 Move SecurityFilterHook registration out of cfg.Hooks.Enabled conditional
- [x] 4.2 Keep AccessControl and EventPublishing hooks config-gated
- [x] 4.3 Pass cfg.Hooks.BlockedCommands to NewSecurityFilterHook for user pattern merge

## 5. DataRoot Path Enforcement

- [x] 5.1 Add DataRoot field to Config struct with mapstructure/json tags
- [x] 5.2 Add default DataRoot ("~/.lango") to DefaultConfig and viper defaults
- [x] 5.3 Implement expandTilde(path, home) with cached home dir parameter
- [x] 5.4 Implement NormalizePaths expanding ~ and resolving relative paths under DataRoot
- [x] 5.5 Implement ValidateDataPaths checking all data paths are under DataRoot
- [x] 5.6 Wire NormalizePaths and ValidateDataPaths into Load pipeline

## 6. Config Type Extensions

- [x] 6.1 Add AdditionalProtectedPaths to ExecToolConfig
- [x] 6.2 Wire DataRoot + AdditionalProtectedPaths into CommandGuard construction in app.New()
- [x] 6.3 Update buildTools/buildExecTools signatures to accept CommandGuard
