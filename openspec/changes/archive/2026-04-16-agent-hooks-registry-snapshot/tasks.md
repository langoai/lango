## 1. DefaultSaveableTools constant

- [x] 1.1 Add `DefaultSaveableTools` package-level var in `internal/toolchain/hook_knowledge.go` with conservative read-type tool names
- [x] 1.2 Add unit test validating `DefaultSaveableTools` is non-empty and contains no duplicates

## 2. Public BuildHookRegistry helper

- [x] 2.1 In `internal/app/app.go`, extract `buildHookRegistry` into public `BuildHookRegistry(cfg, bus, knowledgeSaver)` and have the private call site invoke it
- [x] 2.2 Wire `KnowledgeSaveHook` with `DefaultSaveableTools` into `BuildHookRegistry` when `cfg.Hooks.KnowledgeSave` is true
- [x] 2.3 Add unit test for `BuildHookRegistry`: given configs with all/some/none hooks enabled, assert expected hook names

## 3. Extend `lango agent hooks` CLI

- [x] 3.1 In `internal/cli/agent/hooks.go`, after loading config call `app.BuildHookRegistry(cfg, nil, nil)` to build the registry
- [x] 3.2 Define new output structs: `fullOutput`, `registryOutput`, `hookInfo` with Name, Priority, Phase, Wirable, Details
- [x] 3.3 Populate `hookInfo` from registry. For `KnowledgeSaveHook`, populate `Details` with `saveableTools` list
- [x] 3.4 Extend JSON output: `registry` field alongside existing fields (backward compatible)
- [x] 3.5 Extend text output: "Registered Hooks" section after existing config display
- [x] 3.6 Add unit tests for both JSON and text output structs

## 4. OpenSpec delta spec

- [x] 4.1 Delta spec for `cli-agent-tools-hooks` created in `specs/cli-agent-tools-hooks/spec.md` — covers registry snapshot, BuildHookRegistry, DefaultSaveableTools scenarios

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./...` passes (toolchain, app, cli/agent)
- [x] 5.3 Manual: requires passphrase (expected — config loader behavior), code path verified via unit tests
- [x] 5.4 Manual: JSON output struct verified via unit tests, backward-compatible fields preserved
