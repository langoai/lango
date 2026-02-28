## 1. Component Lifecycle Registry (Phase 0)

- [x] 1.1 Create `internal/lifecycle/component.go` with Component interface, Priority constants, and ComponentEntry struct
- [x] 1.2 Create `internal/lifecycle/registry.go` with Registry (Register, StartAll with priority order, StopAll reverse, rollback on failure)
- [x] 1.3 Create `internal/lifecycle/adapter.go` with SimpleComponent, FuncComponent, and ErrorComponent adapters
- [x] 1.4 Create `internal/lifecycle/registry_test.go` with start order, reverse stop, rollback, empty registry, same-priority tests
- [x] 1.5 Create `internal/lifecycle/adapter_test.go` with tests for all adapter types
- [x] 1.6 Add `registry *lifecycle.Registry` field to App struct in `internal/app/types.go`
- [x] 1.7 Add `registerLifecycleComponents()` method and wire registry in `New()` in `internal/app/app.go`
- [x] 1.8 Refactor `App.Start()` to delegate to `registry.StartAll()`
- [x] 1.9 Refactor `App.Stop()` to delegate to `registry.StopAll()`

## 2. Tool Middleware Chain (Phase 2)

- [x] 2.1 Create `internal/toolchain/middleware.go` with Middleware type, Chain(), and ChainAll() functions
- [x] 2.2 Create `internal/toolchain/mw_learning.go` with WithLearning middleware
- [x] 2.3 Create `internal/toolchain/mw_approval.go` with WithApproval middleware (NeedsApproval, BuildApprovalSummary, Truncate)
- [x] 2.4 Create `internal/toolchain/mw_browser.go` with WithBrowserRecovery middleware
- [x] 2.5 Create `internal/toolchain/middleware_test.go` with chain order and composition tests
- [x] 2.6 Refactor `internal/app/tools.go` wrapping functions to delegate to toolchain package

## 3. Bootstrap Phase Pipeline (Phase 3)

- [x] 3.1 Create `internal/bootstrap/pipeline.go` with Phase struct, Pipeline, State, and Execute with reverse cleanup
- [x] 3.2 Create `internal/bootstrap/phases.go` with DefaultPhases() returning 7 bootstrap phases
- [x] 3.3 Create `internal/bootstrap/pipeline_test.go` with phase failure/rollback and state passing tests
- [x] 3.4 Refactor `bootstrap.Run()` from 230-line function to Pipeline.Execute() invocation (~3 lines)

## 4. AppBuilder Module System (Phase 1)

- [x] 4.1 Create `internal/appinit/module.go` with Module interface, Provides keys, Resolver interface, and ModuleResult
- [x] 4.2 Create `internal/appinit/topo_sort.go` with Kahn's algorithm topological sort and cycle detection
- [x] 4.3 Create `internal/appinit/builder.go` with Builder (AddModule/Build) and BuildResult aggregation
- [x] 4.4 Create `internal/appinit/topo_sort_test.go` with dependency ordering, cycle detection, and disabled module tests
- [x] 4.5 Create `internal/appinit/builder_test.go` with resolver passing, tool aggregation, and error propagation tests

## 5. Event Bus (Phase 4)

- [x] 5.1 Create `internal/eventbus/bus.go` with Bus (Subscribe/Publish), Event interface, and SubscribeTyped[T] generic helper
- [x] 5.2 Create `internal/eventbus/events.go` with ContentSavedEvent, TriplesExtractedEvent, TurnCompletedEvent, ReputationChangedEvent, MemoryGraphEvent
- [x] 5.3 Create `internal/eventbus/bus_test.go` with publish/subscribe order, multiple handlers, no-handler, and thread-safety tests

## 6. Integration Verification

- [x] 6.1 Verify `go build ./...` passes with all new packages
- [x] 6.2 Verify `go test ./...` passes for all new package tests
- [x] 6.3 Verify all existing tests in `internal/app/` continue to pass
