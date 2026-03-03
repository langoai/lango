## 1. Agent Registry — Types & Parser

- [x] 1.1 Create `internal/agentregistry/agent.go` with AgentDefinition, AgentSource, AgentMeta types
- [x] 1.2 Create `internal/agentregistry/parser.go` with ParseAgentMD (YAML frontmatter + markdown body)
- [x] 1.3 Create `internal/agentregistry/parser_test.go` with roundtrip and edge case tests

## 2. Agent Registry — Registry & Stores

- [x] 2.1 Create `internal/agentregistry/registry.go` with Registry (LoadFromStore, Active, All, Specs)
- [x] 2.2 Create `internal/agentregistry/file_store.go` with FileStore (Load from directory)
- [x] 2.3 Create `internal/agentregistry/embed.go` with EmbeddedStore (embed.FS for defaults)
- [x] 2.4 Create `internal/agentregistry/options.go` with Store interface and functional options
- [x] 2.5 Create `internal/agentregistry/registry_test.go` with override priority and Active tests
- [x] 2.6 Create `internal/agentregistry/file_store_test.go` with FileStore loading tests
- [x] 2.7 Create `internal/agentregistry/embed_test.go` with EmbeddedStore loading tests

## 3. Built-in Agents as AGENT.md Defaults

- [x] 3.1 Create `internal/agentregistry/defaults/operator/AGENT.md`
- [x] 3.2 Create `internal/agentregistry/defaults/navigator/AGENT.md`
- [x] 3.3 Create `internal/agentregistry/defaults/vault/AGENT.md`
- [x] 3.4 Create `internal/agentregistry/defaults/librarian/AGENT.md`
- [x] 3.5 Create `internal/agentregistry/defaults/automator/AGENT.md`
- [x] 3.6 Create `internal/agentregistry/defaults/planner/AGENT.md`
- [x] 3.7 Create `internal/agentregistry/defaults/chronicler/AGENT.md`

## 4. Dynamic Tool Partitioning

- [x] 4.1 Add `DynamicToolSet` and `PartitionToolsDynamic()` to `internal/orchestration/tools.go`
- [x] 4.2 Preserve existing `PartitionTools()` as backward-compatible wrapper
- [x] 4.3 Add `BuiltinSpecs()` export function
- [x] 4.4 Add tests for dynamic partitioning in `orchestrator_test.go`

## 5. BuildAgentTree Dynamic Specs Support

- [x] 5.1 Add `Config.Specs []AgentSpec` field to orchestration Config
- [x] 5.2 Update `BuildAgentTree` to use `cfg.Specs` when provided
- [x] 5.3 Add `Capabilities` field to `routingEntry`
- [x] 5.4 Add capabilities to orchestrator instruction routing table

## 6. Agent Context Propagation

- [x] 6.1 Create `internal/ctxkeys/ctxkeys.go` with WithAgentName/AgentNameFromContext
- [x] 6.2 Create `internal/ctxkeys/ctxkeys_test.go` with context key tests
- [x] 6.3 Integrate agent name injection in ADK tool adapter (`internal/adk/tools.go`)

## 7. Tool Execution Hooks — Types & Registry

- [x] 7.1 Create `internal/toolchain/hooks.go` with HookContext, PreToolHook, PostToolHook interfaces, PreHookResult
- [x] 7.2 Create `internal/toolchain/hook_registry.go` with priority-based HookRegistry
- [x] 7.3 Create `internal/toolchain/hooks_test.go` with table-driven tests

## 8. Hook Middleware Bridge

- [x] 8.1 Create `internal/toolchain/mw_hooks.go` with WithHooks() middleware
- [x] 8.2 Create `internal/toolchain/mw_hooks_test.go` with integration tests

## 9. Built-in Hook Implementations

- [x] 9.1 Create `internal/toolchain/hook_security.go` with SecurityFilterHook (priority 10)
- [x] 9.2 Create `internal/toolchain/hook_access.go` with AgentAccessControlHook (priority 20)
- [x] 9.3 Create `internal/toolchain/hook_eventbus.go` with EventBusHook (priority 50)
- [x] 9.4 Create `internal/toolchain/hook_knowledge.go` with KnowledgeSaveHook (priority 100)
- [x] 9.5 Create tests for all built-in hooks

## 10. Sub-Session & Context Isolation

- [x] 10.1 Create `internal/session/child.go` with ChildSession and ChildSessionConfig types
- [x] 10.2 Create `internal/session/child_store.go` with ChildSessionStore interface
- [x] 10.3 Create `internal/session/child_test.go` with child session tests
- [x] 10.4 Create `internal/adk/child_session_service.go` with ChildSessionServiceAdapter
- [x] 10.5 Create `internal/adk/summarizer.go` with StructuredSummarizer and LLMSummarizer
- [x] 10.6 Create `internal/adk/child_session_test.go` with adapter tests

## 11. Agent Memory

- [x] 11.1 Create `internal/agentmemory/types.go` with MemoryEntry, MemoryScope, MemoryKind types
- [x] 11.2 Create `internal/agentmemory/store.go` with Store interface
- [x] 11.3 Create `internal/agentmemory/mem_store.go` with in-memory MemStore implementation
- [x] 11.4 Create `internal/agentmemory/mem_store_test.go` with store operation tests
- [x] 11.5 Create `internal/app/tools_agentmemory.go` with memory_agent_save/recall/forget tools

## 12. P2P Agent Pool

- [x] 12.1 Create `internal/p2p/agentpool/pool.go` with Pool, Agent, Selector, HealthChecker types
- [x] 12.2 Create `internal/p2p/agentpool/pool_test.go` with pool and selector tests
- [x] 12.3 Create `internal/p2p/agentpool/provider.go` with DynamicAgentProvider and PoolProvider
- [x] 12.4 Create `internal/p2p/agentpool/provider_test.go` with provider tests

## 13. P2P Team Coordination

- [x] 13.1 Create `internal/p2p/team/team.go` with Team, Member, TeamState, MemberRole types
- [x] 13.2 Create `internal/p2p/team/coordinator.go` with Coordinator (FormTeam, DelegateTask, CollectResults, DisbandTeam)
- [x] 13.3 Create `internal/p2p/team/conflict.go` with conflict resolution strategies
- [x] 13.4 Create `internal/p2p/team/coordinator_test.go` with coordinator tests
- [x] 13.5 Create `internal/p2p/team/team_test.go` with team type tests

## 14. P2P Team Payment

- [x] 14.1 Create `internal/p2p/team/payment.go` with NegotiatePayment and PaymentAgreement
- [x] 14.2 Create `internal/p2p/team/payment_test.go` with trust-based payment mode tests

## 15. P2P Events & Protocol Messages

- [x] 15.1 Add team events to `internal/eventbus/team_events.go`
- [x] 15.2 Create `internal/eventbus/team_events_test.go` with event type tests
- [x] 15.3 Create `internal/p2p/protocol/team_messages.go` with team protocol messages
- [x] 15.4 Create `internal/p2p/protocol/team_messages_test.go` with message type tests

## 16. App Wiring — Registry, Hooks, P2P Integration

- [x] 16.1 Add `AgentConfig.AgentsDir` and `HooksConfig` to `internal/config/types.go`
- [x] 16.2 Wire agent registry in `internal/app/wiring.go` (initAgent)
- [x] 16.3 Wire P2P agent pool and team coordinator in `internal/app/wiring_p2p.go`
- [x] 16.4 Add P2P fields to App struct in `internal/app/types.go`
- [x] 16.5 Update `internal/app/app.go` to pass p2pComponents to initAgent

## 17. P2P Dynamic Agent Provider

- [x] 17.1 Wire DynamicAgents provider to orchestration Config
- [x] 17.2 Add P2P agent routing table integration in BuildAgentTree

## 18. CLI Updates

- [x] 18.1 Update `internal/cli/agent/list.go` with registry-aware agent loading
- [x] 18.2 Update `internal/cli/agent/status.go` with registry info, P2P, and hooks status

## 19. Build & Test Verification

- [x] 19.1 Run `go build ./...` — passes
- [x] 19.2 Run `go test ./...` — all 76 test packages pass
