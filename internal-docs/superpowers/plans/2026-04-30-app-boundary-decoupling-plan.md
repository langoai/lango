# App Boundary Decoupling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `cmd/lango` the only production importer of `internal/app` while preserving CLI behavior and locking the boundary with OpenSpec deltas and archtest.

**Architecture:** Move app-independent hook registry construction into `internal/toolchain`, remove unused app concrete dependencies from cockpit, and move status dead-letter app bootstrap into `cmd/lango`. CLI packages receive function providers or exported narrow interfaces instead of importing `internal/app`.

**Tech Stack:** Go, Cobra CLI, OpenSpec, internal archtest via `go list -json`, existing `go build ./...` and `go test ./...` verification.

---

## Source Design

- Design spec: `internal-docs/superpowers/specs/2026-04-30-app-boundary-decoupling-design.md`
- OpenSpec workflow reference: `.claude/guides/openspec/workflows.md`
- Teammate roles reference: `.claude/rules/teammate.md`

## File Map

- Create `openspec/changes/app-boundary-decoupling/proposal.md`: documents the architecture boundary change.
- Create `openspec/changes/app-boundary-decoupling/design.md`: records implementation decisions from the approved design.
- Create `openspec/changes/app-boundary-decoupling/tasks.md`: tracks execution tasks and verification.
- Create `openspec/changes/app-boundary-decoupling/specs/cli-agent-tools-hooks/spec.md`: changes `BuildHookRegistry` ownership from `internal/app` to `internal/toolchain`.
- Create `openspec/changes/app-boundary-decoupling/specs/cockpit-status-page/spec.md`: changes the status page contract to a feature status provider function.
- Create `openspec/changes/app-boundary-decoupling/specs/feature-status/spec.md`: clarifies CLI/TUI consumption through `types.FeatureStatus` slices or providers, while `StatusCollector` remains app-layer aggregation.
- Create `openspec/changes/app-boundary-decoupling/specs/application-core/spec.md`: clarifies CLI packages must not import `internal/app`.
- Create `internal/toolchain/build_hook_registry.go`: exported `toolchain.BuildHookRegistry`.
- Move `internal/app/build_hook_registry_test.go` to `internal/toolchain/build_hook_registry_test.go`.
- Modify `internal/app/app.go`: remove `BuildHookRegistry` and private `buildHookRegistry`; call `toolchain.BuildHookRegistry` directly.
- Modify `internal/cli/agent/hooks.go`: call `toolchain.BuildHookRegistry`.
- Modify `internal/cli/agent/hooks_test.go`: keep package-local CLI JSON compatibility tests.
- Modify `internal/cli/cockpit/deps.go`: remove unused `*app.StatusCollector` field and `internal/app` import.
- Modify `cmd/lango/main.go`: remove `FeatureStatuses` from `cockpit.Deps`; pass a production dead-letter loader into `clistatus.NewStatusCmd`.
- Create `cmd/lango/dead_letter_status.go`: lazy app-backed status dead-letter loader for `cmd/lango`.
- Modify `internal/cli/status/status.go`: export dead-letter bridge API, remove `internal/app` import, accept injected `DeadLetterBridgeLoader`.
- Modify `internal/cli/status/status_test.go`: update tests to exported dead-letter types and injected loaders.
- Modify `internal/archtest/boundary_test.go`: add CLI-to-app import boundary rule.

## Commit Policy

Do not run `git commit` or `git push` while executing this plan. Each task includes a suggested commit message and staging scope for the user to apply manually after review.

## Task 1: OpenSpec Change Artifacts

**Files:**
- Create: `openspec/changes/app-boundary-decoupling/proposal.md`
- Create: `openspec/changes/app-boundary-decoupling/design.md`
- Create: `openspec/changes/app-boundary-decoupling/tasks.md`
- Create: `openspec/changes/app-boundary-decoupling/specs/cli-agent-tools-hooks/spec.md`
- Create: `openspec/changes/app-boundary-decoupling/specs/cockpit-status-page/spec.md`
- Create: `openspec/changes/app-boundary-decoupling/specs/feature-status/spec.md`
- Create: `openspec/changes/app-boundary-decoupling/specs/application-core/spec.md`

- [ ] **Step 1: Verify target main-spec requirement titles**

Run:

```bash
rg -n '### Requirement: (Public BuildHookRegistry helper|Feature status and metrics dashboard|StatusCollector aggregation|Application Bootstrap)' openspec/specs/cli-agent-tools-hooks/spec.md openspec/specs/cockpit-status-page/spec.md openspec/specs/feature-status/spec.md openspec/specs/application-core/spec.md
```

Expected output:

```text
openspec/specs/cockpit-status-page/spec.md:<line>:### Requirement: Feature status and metrics dashboard
openspec/specs/cli-agent-tools-hooks/spec.md:<line>:### Requirement: Public BuildHookRegistry helper
openspec/specs/feature-status/spec.md:<line>:### Requirement: StatusCollector aggregation
openspec/specs/application-core/spec.md:<line>:### Requirement: Application Bootstrap
```

- [ ] **Step 2: Verify no additional OpenSpec contract references need deltas**

Run:

```bash
rg 'app\.BuildHookRegistry|app\.StatusCollector|deadLetterLoaderFromBoot' openspec --glob '!**/archive/**'
```

Expected: no matches and exit code 1. If matches appear, add a delta spec for the affected capability before creating the change artifacts.

- [ ] **Step 3: Create the OpenSpec change directories**

Run:

```bash
mkdir -p openspec/changes/app-boundary-decoupling/specs/cli-agent-tools-hooks
mkdir -p openspec/changes/app-boundary-decoupling/specs/cockpit-status-page
mkdir -p openspec/changes/app-boundary-decoupling/specs/feature-status
mkdir -p openspec/changes/app-boundary-decoupling/specs/application-core
```

Expected: directories exist under `openspec/changes/app-boundary-decoupling/`.

- [ ] **Step 4: Write `proposal.md`**

Create `openspec/changes/app-boundary-decoupling/proposal.md` with:

```markdown
# App Boundary Decoupling

## Problem

`internal/cli/status`, `internal/cli/cockpit`, and `internal/cli/agent` directly import `internal/app`. This makes `internal/app` behave like a shared utility package instead of a process composition root and increases the cost of future app decomposition.

## Proposed Change

Make `cmd/lango` the only production importer of `internal/app`. Move app-independent hook registry construction to `internal/toolchain`, remove unused app concrete dependencies from cockpit, inject status dead-letter runtime access from `cmd/lango`, and enforce the boundary with `internal/archtest`.

## User-Facing Impact

Command names, flags, output schemas, status rendering, cockpit status rendering, and `lango agent hooks` output remain compatible.
```

- [ ] **Step 5: Write `design.md`**

Create `openspec/changes/app-boundary-decoupling/design.md` with:

```markdown
# Design

## Boundary

`cmd/lango` owns process assembly and may import `internal/app`. Packages under `internal/cli/**` must not import `internal/app`.

## Hook Registry

`BuildHookRegistry` moves from `internal/app` to exported `internal/toolchain.BuildHookRegistry`. Runtime bootstrap and CLI snapshot mode both call the same toolchain function.

## Cockpit Status

Cockpit status page uses `func() []types.FeatureStatus` as its provider contract. The unused `cockpit.Deps.FeatureStatuses *app.StatusCollector` field is removed.

## Status Dead-Letter Loader

`internal/cli/status` exports a narrow dead-letter bridge API and accepts a `DeadLetterBridgeLoader`. `cmd/lango` owns the lazy app-backed loader in `cmd/lango/dead_letter_status.go`.

## Enforcement

`internal/archtest` fails when production packages under `internal/cli/**` import `github.com/langoai/lango/internal/app`.
```

- [ ] **Step 6: Write `tasks.md`**

Create `openspec/changes/app-boundary-decoupling/tasks.md` with:

```markdown
# Tasks

- [ ] Move `BuildHookRegistry` to `internal/toolchain`
- [ ] Remove cockpit's unused app concrete dependency
- [ ] Export status dead-letter bridge API and inject the production loader from `cmd/lango`
- [ ] Add CLI-to-app boundary archtest
- [ ] Update affected specs
- [ ] Run `go test ./internal/archtest/...`
- [ ] Run `go build ./...`
- [ ] Run `go test ./...`
```

- [ ] **Step 7: Write `cli-agent-tools-hooks` delta spec**

Create `openspec/changes/app-boundary-decoupling/specs/cli-agent-tools-hooks/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: Public BuildHookRegistry helper
The `internal/toolchain` package SHALL export a `BuildHookRegistry(cfg *config.Config, bus *eventbus.Bus, knowledgeSaver toolchain.KnowledgeSaver, catalog *toolcatalog.Catalog) *toolchain.HookRegistry` function that produces the same hook registry as the runtime app builder. When `bus` is nil, EventBus hooks are omitted. When `knowledgeSaver` is nil, `KnowledgeSaveHook` is still registered for snapshot inspection but its `Post` method safely no-ops. When `catalog` is non-nil, `SaveableTools` is derived from catalog; otherwise it falls back to `DefaultSaveableTools`.

#### Scenario: CLI uses BuildHookRegistry without full bootstrap
- **WHEN** the `agent hooks` CLI command loads config and calls `toolchain.BuildHookRegistry(cfg, nil, nil, nil)`
- **THEN** the returned registry contains all config-derivable hooks
- **AND** no database connection, crypto initialization, app bootstrap, or event bus is required

#### Scenario: Runtime path provides KnowledgeSaver
- **WHEN** the app builder calls `toolchain.BuildHookRegistry` during full bootstrap
- **THEN** the `KnowledgeSaver` from the knowledge subsystem is passed through to `KnowledgeSaveHook`
```

- [ ] **Step 8: Write `cockpit-status-page` delta spec**

Create `openspec/changes/app-boundary-decoupling/specs/cockpit-status-page/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: Feature status and metrics dashboard
StatusPage SHALL display feature flags from a `func() []types.FeatureStatus` status provider, token usage and tool execution stats from `MetricsCollector.Snapshot()`, and provider/model info from Config. StatusPage MUST NOT import or depend on `internal/app`.

#### Scenario: Feature flags display
- **WHEN** StatusPage is active and the provider returns feature statuses
- **THEN** it SHALL render each feature with enabled/disabled badge
```

- [ ] **Step 9: Write `feature-status` delta spec**

Create `openspec/changes/app-boundary-decoupling/specs/feature-status/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: StatusCollector aggregation
The app layer SHALL provide a `StatusCollector` that collects `FeatureStatus` from wiring functions. It SHALL expose `All()` to list all statuses and `SilentDisabledCount()` to count features that are disabled with a non-empty reason. CLI and TUI packages SHALL consume feature statuses through `types.FeatureStatus` slices or provider functions and MUST NOT import `internal/app` to access `StatusCollector`.

#### Scenario: CLI/TUI feature status consumption
- **WHEN** CLI or TUI code needs feature statuses
- **THEN** the status data is provided as `[]types.FeatureStatus` or `func() []types.FeatureStatus`
- **AND** the CLI or TUI package does not import `internal/app`
```

- [ ] **Step 10: Write `application-core` delta spec**

Create `openspec/changes/app-boundary-decoupling/specs/application-core/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: Application Bootstrap
The system SHALL initialize all core components through a centralized application entry point (`internal/app`) assembled by `cmd/lango`. CLI packages under `internal/cli/**` SHALL NOT import `internal/app`; they receive narrow interfaces, function providers, DTOs, or app-independent helpers from `cmd/lango`.

#### Scenario: CLI package does not import app
- **WHEN** production imports are scanned
- **THEN** packages under `internal/cli/**` do not import `github.com/langoai/lango/internal/app`
```

- [ ] **Step 11: Verify OpenSpec artifacts exist**

Run:

```bash
find openspec/changes/app-boundary-decoupling -type f | sort
```

Expected output includes the seven files created in this task.

- [ ] **Step 12: Validate OpenSpec change**

Run:

```bash
openspec validate app-boundary-decoupling --strict
```

Expected: PASS.

- [ ] **Step 13: Review staging scope and suggested commit message**

Run:

```bash
git status --short openspec/changes/app-boundary-decoupling
```

Expected: only files under `openspec/changes/app-boundary-decoupling` are listed.

Suggested commit message for the user:

```text
spec: add app boundary decoupling change
```

## Task 2: Move Hook Registry Construction to Toolchain

**Files:**
- Create: `internal/toolchain/build_hook_registry.go`
- Move: `internal/app/build_hook_registry_test.go` -> `internal/toolchain/build_hook_registry_test.go`
- Modify: `internal/app/app.go`
- Modify: `internal/cli/agent/hooks.go`
- Modify: `internal/cli/agent/hooks_test.go`
- Test: `internal/toolchain/build_hook_registry_test.go`
- Test: `internal/cli/agent/hooks_test.go`

- [ ] **Step 1: Verify hook registry symbols and call sites**

Run:

```bash
rg -n 'func BuildHookRegistry|func buildHookRegistry|BuildHookRegistry\(|func buildRegistryOutput|func DefaultConfig' internal/app internal/cli/agent internal/toolchain internal/config --glob '*.go'
```

Expected output includes:

```text
internal/app/app.go:<line>:func BuildHookRegistry(...)
internal/app/app.go:<line>:func buildHookRegistry(...)
internal/app/app.go:<line>:hookRegistry := buildHookRegistry(...)
internal/cli/agent/hooks.go:<line>:registry := app.BuildHookRegistry(...)
internal/cli/agent/hooks.go:<line>:func buildRegistryOutput(...)
internal/config/loader.go:<line>:func DefaultConfig() *Config
```

- [ ] **Step 2: Add the toolchain builder test by moving the existing app test**

Move the file:

```bash
mv internal/app/build_hook_registry_test.go internal/toolchain/build_hook_registry_test.go
```

Edit the first line of `internal/toolchain/build_hook_registry_test.go`:

```go
package toolchain
```

Remove this import because the test is now in the package that defines the hook types:

```go
"github.com/langoai/lango/internal/toolchain"
```

Expected: tests still reference `BuildHookRegistry` directly.

- [ ] **Step 3: Run the moved test and verify it fails before implementation**

Run:

```bash
go test ./internal/toolchain -run 'TestBuildHookRegistry' -count=1
```

Expected: FAIL with `undefined: BuildHookRegistry`.

- [ ] **Step 4: Create `internal/toolchain/build_hook_registry.go`**

Create `internal/toolchain/build_hook_registry.go` with:

```go
package toolchain

import (
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/toolcatalog"
)

// BuildHookRegistry constructs the tool execution hook registry from config.
// Pass nil for bus when running outside a full app; EventBus hooks will be
// omitted. Pass nil for knowledgeSaver when the hook only needs to be
// inspected, not executed. When catalog is non-nil, SaveableTools is derived
// from tool metadata; otherwise it falls back to DefaultSaveableTools.
func BuildHookRegistry(cfg *config.Config, bus *eventbus.Bus, knowledgeSaver KnowledgeSaver, catalog *toolcatalog.Catalog) *HookRegistry {
	hookRegistry := NewHookRegistry()
	hookRegistry.RegisterPre(NewSecurityFilterHook(cfg.Hooks.BlockedCommands))
	if cfg.Hooks.AccessControl {
		hookRegistry.RegisterPre(NewAgentAccessControlHook(nil))
	}
	if (cfg.Hooks.Enabled || cfg.Agent.MultiAgent) && cfg.Hooks.EventPublishing && bus != nil {
		ebHook := NewEventBusHook(bus)
		hookRegistry.RegisterPre(ebHook)
		hookRegistry.RegisterPost(ebHook)
	}
	if cfg.Hooks.KnowledgeSave {
		saveableTools := DefaultSaveableTools
		if catalog != nil {
			if derived := catalog.SaveableToolNames(); len(derived) > 0 {
				saveableTools = derived
			}
		}
		hookRegistry.RegisterPost(NewKnowledgeSaveHook(knowledgeSaver, saveableTools))
	}
	return hookRegistry
}
```

- [ ] **Step 5: Update `internal/app/app.go` imports and call site**

In `internal/app/app.go`, keep the existing `toolchain` import. Replace:

```go
hookRegistry := buildHookRegistry(cfg, bus, knowledgeSaver, catalog)
```

with:

```go
hookRegistry := toolchain.BuildHookRegistry(cfg, bus, knowledgeSaver, catalog)
```

Delete the exported `BuildHookRegistry` function and the private `buildHookRegistry` wrapper from `internal/app/app.go`.

- [ ] **Step 6: Update `internal/cli/agent/hooks.go`**

Remove:

```go
"github.com/langoai/lango/internal/app"
```

Replace:

```go
registry := app.BuildHookRegistry(cfg, nil, nil, nil)
```

with:

```go
registry := toolchain.BuildHookRegistry(cfg, nil, nil, nil)
```

- [ ] **Step 7: Add CLI JSON compatibility coverage**

Append this test to `internal/cli/agent/hooks_test.go`:

```go
func TestAgentHooksJSONSnapshotShape(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Hooks.Enabled = true
	cfg.Hooks.SecurityFilter = true
	cfg.Hooks.AccessControl = true
	cfg.Hooks.KnowledgeSave = true
	cfg.Hooks.BlockedCommands = []string{"rm -rf /"}

	registry := toolchain.BuildHookRegistry(cfg, nil, nil, nil)

	out := fullOutput{
		hooksConfigOutput: hooksConfigOutput{
			Enabled:         cfg.Hooks.Enabled,
			SecurityFilter:  cfg.Hooks.SecurityFilter,
			AccessControl:   cfg.Hooks.AccessControl,
			EventPublishing: cfg.Hooks.EventPublishing,
			KnowledgeSave:   cfg.Hooks.KnowledgeSave,
			BlockedCommands: cfg.Hooks.BlockedCommands,
		},
		Registry: buildRegistryOutput(registry, cfg.Hooks),
	}

	data, err := json.Marshal(out)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, true, decoded["enabled"])
	assert.Equal(t, true, decoded["security_filter"])
	assert.Equal(t, true, decoded["access_control"])
	assert.Equal(t, true, decoded["knowledge_save"])
	assert.Contains(t, decoded, "registry")
}
```

Add `encoding/json` to the imports in `internal/cli/agent/hooks_test.go`.

- [ ] **Step 8: Run focused tests**

Run:

```bash
go test ./internal/toolchain ./internal/app ./internal/cli/agent -run 'TestBuildHookRegistry|TestAgentHooksJSONSnapshotShape|TestBuildRegistryOutput|TestPrintJSON' -count=1
```

Expected: PASS.

- [ ] **Step 9: Review staging scope and suggested commit message**

Run:

```bash
git status --short internal/toolchain/build_hook_registry.go internal/toolchain/build_hook_registry_test.go internal/app/build_hook_registry_test.go internal/app/app.go internal/cli/agent/hooks.go internal/cli/agent/hooks_test.go
```

Expected: only hook-registry related files are listed.

Suggested commit message for the user:

```text
refactor: move hook registry builder to toolchain
```

## Task 3: Remove Cockpit App Concrete Dependency

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `cmd/lango/main.go`
- Test: `internal/cli/cockpit/deps_test.go`
- Test: `internal/cli/cockpit/pages/status_test.go`

- [ ] **Step 1: Verify current cockpit import violation**

Run:

```bash
rg 'github.com/langoai/lango/internal/app["/]' internal/cli/cockpit --glob '*.go' --glob '!*_test.go'
```

Expected: one match in `internal/cli/cockpit/deps.go`.

- [ ] **Step 2: Remove the unused field from `internal/cli/cockpit/deps.go`**

Remove the import:

```go
"github.com/langoai/lango/internal/app"
```

Remove this field from `Deps`:

```go
FeatureStatuses   *app.StatusCollector
```

- [ ] **Step 3: Update `cmd/lango/main.go` cockpit wiring**

Remove this field from the `cockpit.Deps` literal:

```go
FeatureStatuses:   application.FeatureStatuses,
```

Keep the existing status page provider:

```go
var statusProvider func() []types.FeatureStatus
if application.FeatureStatuses != nil {
	statusProvider = application.FeatureStatuses.All
}
model.RegisterPage(cockpit.PageStatus,
	pages.NewStatusPage(statusProvider, application.MetricsCollector, cfg))
```

Do not modify the existing outer guard:

```go
if application.MetricsCollector != nil || application.FeatureStatuses != nil {
	// existing status page registration stays inside this guard
}
```

- [ ] **Step 4: Run focused cockpit tests**

Run:

```bash
go test ./internal/cli/cockpit/... -count=1
```

Expected: PASS.

- [ ] **Step 5: Verify cockpit import is gone**

Run:

```bash
rg 'github.com/langoai/lango/internal/app["/]' internal/cli/cockpit --glob '*.go' --glob '!*_test.go'
```

Expected: no matches and exit code 1.

- [ ] **Step 6: Review staging scope and suggested commit message**

Run:

```bash
git status --short internal/cli/cockpit/deps.go cmd/lango/main.go
```

Expected: only cockpit dependency removal files are listed.

Suggested commit message for the user:

```text
refactor: remove cockpit app dependency
```

## Task 4: Inject Status Dead-Letter Loader from cmd/lango

**Files:**
- Create: `cmd/lango/dead_letter_status.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`
- Test: `internal/cli/status/status_test.go`
- Test: `cmd/lango/main_test.go`

- [ ] **Step 1: Verify status dead-letter symbols and ripple surface**

Run:

```bash
rg -n 'type deadLetterBridge|type deadLetterBridgeLoader|type deadLetterListOptions|type deadLetterListPage|type fakeDeadLetterBridge|toolCatalogDeadLetterBridge|func executeCommand|NewStatusCmd\(|cliboot\.BootResult' internal/cli/status cmd/lango --glob '*.go'
```

Expected output includes:

```text
internal/cli/status/status.go:<line>:type deadLetterBridge interface
internal/cli/status/status.go:<line>:type deadLetterBridgeLoader func() ...
internal/cli/status/status.go:<line>:type deadLetterListOptions struct
internal/cli/status/status.go:<line>:type deadLetterListPage struct
internal/cli/status/status.go:<line>:type toolCatalogDeadLetterBridge struct
internal/cli/status/status_test.go:<line>:type fakeDeadLetterBridge struct
internal/cli/status/status_test.go:<line>:func executeCommand(...)
cmd/lango/main.go:<line>:statusCmd := clistatus.NewStatusCmd(cliboot.BootResult)
```

If additional implementations of `deadLetterBridge` appear, add `Ready() bool` to each test fake or implementation in the same task.

- [ ] **Step 2: Export status dead-letter types**

In `internal/cli/status/status.go`, replace:

```go
type deadLetterBridge interface {
	List(context.Context, deadLetterListOptions) (deadLetterListPage, error)
	Detail(context.Context, string) (postadjudicationstatus.TransactionStatus, error)
	Retry(context.Context, string) error
}

type deadLetterBridgeLoader func() (deadLetterBridge, func(), error)
```

with:

```go
type DeadLetterBridge interface {
	Ready() bool
	List(context.Context, DeadLetterListOptions) (DeadLetterListPage, error)
	Detail(context.Context, string) (postadjudicationstatus.TransactionStatus, error)
	Retry(context.Context, string) error
}

type DeadLetterBridgeLoader func() (DeadLetterBridge, func(), error)
```

Rename `deadLetterListOptions` to `DeadLetterListOptions` and `deadLetterListPage` to `DeadLetterListPage` throughout `internal/cli/status/status.go`.

The rename must include type definitions, method signatures, parameter types, return types, zero-value returns, and helper function signatures such as `toolCatalogDeadLetterBridge.List`, `aggregateDeadLetterSummary`, and `summaryTotal`.

- [ ] **Step 3: Export the tool catalog bridge constructor and readiness method**

In `internal/cli/status/status.go`, rename:

```go
type toolCatalogDeadLetterBridge struct {
	catalog *toolcatalog.Catalog
}
```

to:

```go
type toolCatalogDeadLetterBridge struct {
	catalog *toolcatalog.Catalog
}

func NewToolCatalogDeadLetterBridge(catalog *toolcatalog.Catalog) DeadLetterBridge {
	return &toolCatalogDeadLetterBridge{catalog: catalog}
}
```

Rename:

```go
func (b *toolCatalogDeadLetterBridge) ready() bool
```

to:

```go
func (b *toolCatalogDeadLetterBridge) Ready() bool
```

- [ ] **Step 4: Change `NewStatusCmd` signature and nil loader behavior**

In `internal/cli/status/status.go`, change:

```go
func NewStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
```

to:

```go
func NewStatusCmd(bootLoader func() (*bootstrap.Result, error), deadLetterLoader DeadLetterBridgeLoader) *cobra.Command {
```

Add this helper near `NewStatusCmd`:

```go
func unavailableDeadLetterLoader() DeadLetterBridgeLoader {
	return func() (DeadLetterBridge, func(), error) {
		return nil, nil, fmt.Errorf("dead-letter status tools are not available")
	}
}
```

At the start of `NewStatusCmd`, after local var declarations, add:

```go
if deadLetterLoader == nil {
	deadLetterLoader = unavailableDeadLetterLoader()
}
```

Replace all uses of `deadLetterLoaderFromBoot(bootLoader)` with `deadLetterLoader`.

- [ ] **Step 5: Remove app bootstrap from status package**

Delete `deadLetterLoaderFromBoot` from `internal/cli/status/status.go`.

Remove this import from `internal/cli/status/status.go`:

```go
"github.com/langoai/lango/internal/app"
```

Keep the `internal/bootstrap` import because the base status command still accepts `bootLoader`.

- [ ] **Step 6: Create `cmd/lango/dead_letter_status.go`**

Create `cmd/lango/dead_letter_status.go` with:

```go
package main

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	clistatus "github.com/langoai/lango/internal/cli/status"
)

func deadLetterStatusLoaderFromBoot(bootLoader func() (*bootstrap.Result, error)) clistatus.DeadLetterBridgeLoader {
	return func() (clistatus.DeadLetterBridge, func(), error) {
		boot, err := bootLoader()
		if err != nil {
			return nil, nil, fmt.Errorf("bootstrap: %w", err)
		}
		application, err := app.New(boot, app.WithLocalChat())
		if err != nil {
			_ = boot.Close()
			return nil, nil, fmt.Errorf("build app: %w", err)
		}
		cleanup := func() {
			_ = application.Stop(context.Background())
			_ = boot.Close()
		}
		bridge := clistatus.NewToolCatalogDeadLetterBridge(application.ToolCatalog)
		if !bridge.Ready() {
			cleanup()
			return nil, nil, fmt.Errorf("dead-letter status tools are not available")
		}
		return bridge, cleanup, nil
	}
}
```

- [ ] **Step 7: Wire the production loader in `cmd/lango/main.go`**

Replace:

```go
statusCmd := clistatus.NewStatusCmd(cliboot.BootResult)
```

with:

```go
statusCmd := clistatus.NewStatusCmd(cliboot.BootResult, deadLetterStatusLoaderFromBoot(cliboot.BootResult))
```

- [ ] **Step 8: Update status tests for exported names**

In `internal/cli/status/status_test.go`, rename:

```go
deadLetterListPage
deadLetterListOptions
deadLetterBridge
```

to:

```go
DeadLetterListPage
DeadLetterListOptions
DeadLetterBridge
```

Add this method to `fakeDeadLetterBridge`:

```go
func (f *fakeDeadLetterBridge) Ready() bool {
	return true
}
```

Update the wiring test:

```go
func TestNewStatusCmd_WiresDeadLetterSummaryCommand(t *testing.T) {
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, errors.New("should not bootstrap for wiring test")
	}, nil)

	names := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.Contains(t, names, "dead-letter-summary")
}
```

- [ ] **Step 9: Add a nil loader behavior test**

Append this test to `internal/cli/status/status_test.go`:

```go
func TestNewStatusCmd_NilDeadLetterLoaderReturnsUnavailable(t *testing.T) {
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, errors.New("should not bootstrap for dead-letter loader")
	}, nil)

	_, err := executeCommand(t, cmd, "dead-letter-summary")

	require.Error(t, err)
	assert.ErrorContains(t, err, "dead-letter status tools are not available")
}
```

- [ ] **Step 10: Update all status command call sites**

Run:

```bash
rg -n 'NewStatusCmd\(' cmd internal --glob '*.go'
```

Expected after edits:

```text
cmd/lango/main.go:<line>:	statusCmd := clistatus.NewStatusCmd(cliboot.BootResult, deadLetterStatusLoaderFromBoot(cliboot.BootResult))
internal/cli/status/status.go:<line>:func NewStatusCmd(bootLoader func() (*bootstrap.Result, error), deadLetterLoader DeadLetterBridgeLoader) *cobra.Command {
```

Test files may also call `NewStatusCmd` with two arguments.

- [ ] **Step 11: Run focused status and cmd tests**

Run:

```bash
go test ./internal/cli/status ./cmd/lango -count=1
```

Expected: PASS.

- [ ] **Step 12: Verify status package no longer imports app**

Run:

```bash
rg 'github.com/langoai/lango/internal/app["/]' internal/cli/status --glob '*.go' --glob '!*_test.go'
```

Expected: no matches and exit code 1.

- [ ] **Step 13: Review staging scope and suggested commit message**

Run:

```bash
git status --short cmd/lango/dead_letter_status.go cmd/lango/main.go cmd/lango/main_test.go internal/cli/status/status.go internal/cli/status/status_test.go
```

Expected: only status loader injection files are listed.

Suggested commit message for the user:

```text
refactor: inject status dead letter loader
```

## Task 5: Add CLI-to-App Boundary Archtest

**Files:**
- Modify: `internal/archtest/boundary_test.go`
- Test: `internal/archtest/boundary_test.go`

- [ ] **Step 1: Add the boundary rule**

In `internal/archtest/boundary_test.go`, append this rule to `rules`:

```go
{
	name: "cli packages must not import app",
	sourceMatch: func(importPath string) bool {
		return importPath == modulePath+"internal/cli" ||
			strings.HasPrefix(importPath, modulePath+"internal/cli/")
	},
	forbiddenMatch: func(dep string) bool {
		prefix := modulePath + "internal/app"
		return dep == prefix || strings.HasPrefix(dep, prefix+"/")
	},
},
```

- [ ] **Step 2: Run archtest**

Run:

```bash
go test ./internal/archtest/... -count=1
```

Expected: PASS.

- [ ] **Step 3: Verify repo-wide production app import matches**

Run:

```bash
rg 'github.com/langoai/lango/internal/app["/]' --glob '*.go' --glob '!*_test.go'
```

Expected output includes only files under `cmd/lango/`, specifically:

```text
cmd/lango/dead_letter_status.go:...
cmd/lango/main.go:...
```

- [ ] **Step 4: Review staging scope and suggested commit message**

Run:

```bash
git status --short internal/archtest/boundary_test.go
```

Expected: only `internal/archtest/boundary_test.go` is listed.

Suggested commit message for the user:

```text
test: enforce cli app boundary
```

## Task 6: Final Verification and OpenSpec Completion

**Files:**
- Modify: `openspec/changes/app-boundary-decoupling/tasks.md`
- Modify after OpenSpec archive/sync: `openspec/specs/cli-agent-tools-hooks/spec.md`
- Modify after OpenSpec archive/sync: `openspec/specs/cockpit-status-page/spec.md`
- Modify after OpenSpec archive/sync: `openspec/specs/feature-status/spec.md`
- Modify after OpenSpec archive/sync: `openspec/specs/application-core/spec.md`

- [ ] **Step 1: Mark OpenSpec tasks complete**

Edit `openspec/changes/app-boundary-decoupling/tasks.md` so every task checkbox is checked:

```markdown
# Tasks

- [x] Move `BuildHookRegistry` to `internal/toolchain`
- [x] Remove cockpit's unused app concrete dependency
- [x] Export status dead-letter bridge API and inject the production loader from `cmd/lango`
- [x] Add CLI-to-app boundary archtest
- [x] Update affected specs
- [x] Run `go test ./internal/archtest/...`
- [x] Run `go build ./...`
- [x] Run `go test ./...`
```

- [ ] **Step 2: Run acceptance grep**

Run:

```bash
rg 'github.com/langoai/lango/internal/app["/]' --glob '*.go' --glob '!*_test.go'
```

Expected: only files under `cmd/lango/`.

- [ ] **Step 3: Run OpenSpec-focused validation**

Run:

```bash
openspec validate app-boundary-decoupling --strict
```

Expected: PASS.

- [ ] **Step 4: Run focused Go tests**

Run:

```bash
go test ./internal/toolchain ./internal/app ./internal/cli/agent ./internal/cli/cockpit/... ./internal/cli/status ./cmd/lango ./internal/archtest/... -count=1
```

Expected: PASS.

- [ ] **Step 5: Run full build**

Run:

```bash
go build ./...
```

Expected: PASS.

- [ ] **Step 6: Run full test suite and record exit code**

Run:

```bash
go test ./...
```

Expected: PASS with exit code 0. If a failure appears, do not mark this task complete. Capture the exact package, test name, command exit code, and error output. Use `git diff --name-only` to identify the files changed since the previous review checkpoint, then either fix the failing task's files or ask for approval before reverting that task's file set.

- [ ] **Step 7: Review final OpenSpec sync scope and suggested commit message**

Run:

```bash
git status --short openspec/changes/app-boundary-decoupling openspec/specs/cli-agent-tools-hooks/spec.md openspec/specs/cockpit-status-page/spec.md openspec/specs/feature-status/spec.md openspec/specs/application-core/spec.md
```

Expected: only app-boundary OpenSpec files are listed.

Suggested commit message for the user if files changed:

```text
spec: complete app boundary decoupling
```

- [ ] **Step 8: Archive the OpenSpec change**

Before archiving, review what will be applied to the main specs:

```bash
openspec diff app-boundary-decoupling
```

Expected: diff is limited to `cli-agent-tools-hooks`, `cockpit-status-page`, `feature-status`, and `application-core`. If the local OpenSpec CLI does not support `diff`, run:

```bash
git diff -- openspec/specs/cli-agent-tools-hooks/spec.md openspec/specs/cockpit-status-page/spec.md openspec/specs/feature-status/spec.md openspec/specs/application-core/spec.md
```

Expected: diff is limited to the approved app boundary contract changes.

Run:

```bash
openspec archive app-boundary-decoupling --yes
```

Expected: change archives and affected specs are synced.

- [ ] **Step 9: Review archive output scope and suggested commit message**

Run:

```bash
git status --short
```

Expected: archive output is limited to `openspec` files for `app-boundary-decoupling`.

Suggested commit message for the user if archive changed files:

```text
spec: archive app boundary decoupling
```

## Self-Review

- Spec coverage: Task 1 covers OpenSpec deltas for `cli-agent-tools-hooks`, `cockpit-status-page`, `feature-status`, and `application-core`. Task 2 implements hook registry ownership. Task 3 implements cockpit status provider decoupling. Task 4 implements status dead-letter injection. Task 5 implements archtest enforcement. Task 6 covers verification and archive.
- Placeholder scan: no placeholder markers remain, and all former `git commit` execution steps have been converted to manual staging-scope review plus suggested commit messages.
- Type consistency: `DeadLetterBridge`, `DeadLetterBridgeLoader`, `DeadLetterListOptions`, `DeadLetterListPage`, and `toolchain.BuildHookRegistry` names are used consistently across tasks.
- Precondition checks: Task 1 verifies main-spec requirement titles, Task 2 verifies hook registry symbols and call sites, and Task 4 verifies status bridge ripple surface before edits.
- Failure handling: Task 6 requires exit-code capture for `go test ./...` failures and uses `git diff --name-only` to identify the changed file set before fixing or requesting approval to revert.
