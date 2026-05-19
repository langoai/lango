## 1. Planning And Baseline

- [x] 1.1 Measure the current non-Ent coverage baseline
- [x] 1.2 Identify top files by uncovered statement count
- [x] 1.3 Validate and commit this umbrella OpenSpec plan

## 2. Coverage Measurement Contract

- [x] 2.1 Add a repeatable coverage report command or script for non-generated Go code
- [x] 2.2 Exclude generated code by known generated paths and generated-file markers
- [x] 2.3 Add tests for the coverage filtering logic
- [x] 2.4 Report total coverage, covered statements, total statements, uncovered statements, and top uncovered files
- [x] 2.5 Add an executable threshold gate that fails when non-generated coverage is below 90%

## 3. Coverage Wave: Small Deterministic Packages

- [x] 3.1 Add focused tests for low-risk pure/helper packages with low coverage
- [x] 3.2 Add meaningful assertions for enum, parser, embed, and simple wrapper behavior
- [x] 3.3 Commit the wave separately

## 4. Coverage Wave: CLI/TUI State And Command Surfaces

- [x] 4.1 Add targeted tests for `internal/cli/tuicore/state_update.go`
- [x] 4.2 Add targeted tests for `internal/cli/settings/editor.go` and `internal/cli/settings/menu.go`
- [x] 4.3 Cover command construction, output seams, error paths, and state transitions
- [x] 4.4 Commit the wave separately

## 5. Coverage Wave: Storage And Knowledge

- [x] 5.1 Add focused tests for `internal/storage/facade.go`
- [x] 5.2 Add focused tests for `internal/storagebroker/server.go` and `internal/storagebroker/client.go`
- [x] 5.3 Add focused tests for `internal/knowledge/store.go`
- [x] 5.4 Cover CRUD, error handling, unavailable backends, and boundary conditions
- [x] 5.5 Commit the wave separately

## 6. Coverage Wave: Runtime And Workflow

- [x] 6.1 Add deterministic tests or seams for `internal/adk/agent.go`
- [x] 6.2 Add deterministic tests for `internal/workflow/engine.go`
- [x] 6.3 Add deterministic tests for `internal/turnrunner/runner.go`
- [x] 6.4 Add persistence boundary tests for `internal/turntrace/store.go`
- [x] 6.5 Cover workflow execution/status paths, turn-runner outcomes, and trace persistence boundaries
- [x] 6.6 Commit the wave separately

## 7. Coverage Wave: App Wiring And P2P

- [x] 7.1 Add focused tests for `internal/app/wiring_p2p.go`
- [x] 7.2 Add focused tests for `internal/app/wiring.go`
- [x] 7.3 Add focused tests for `internal/app/tools_meta.go`
- [x] 7.4 Add focused tests for `internal/p2p/handshake/handshake.go`
- [x] 7.5 Commit the wave separately

## 8. Coverage Wave: High-Impact Service Boundaries

- [x] 8.1 Add focused tests for remaining `internal/storagebroker/server.go` dispatch, payload protection, and error paths
- [x] 8.2 Add focused tests for remaining `internal/skill/importer.go` import validation, filesystem, and failure paths
- [x] 8.3 Add focused tests for remaining `internal/receipts/store.go` receipt lifecycle and query boundaries
- [x] 8.4 Commit the wave separately

## 9. Coverage Wave: App And CLI Execution Surfaces

- [x] 9.1 Add focused tests for remaining `internal/app/app.go`, `internal/app/modules.go`, and app wiring branches
- [x] 9.2 Add focused tests for `cmd/lango/main.go` command routing and error exits through seams
- [x] 9.3 Add focused tests for `internal/cli/chat/chat.go`, `internal/cli/settings/editor.go`, `internal/cli/sandbox/sandbox.go`, and `internal/cli/onboard/wizard.go`
- [x] 9.4 Commit the wave separately

## 10. Coverage Wave: Remaining P2P And Payment Surfaces

- [x] 10.1 Add focused tests for remaining `internal/p2p/handshake/handshake.go` stream and approval branches
- [x] 10.2 Add focused tests for high-uncovered P2P protocol, settlement, discovery, and payment paths that do not require external services
- [x] 10.3 Add focused tests for remaining `internal/app/tools_escrow.go` and payment tool wrappers
- [x] 10.4 Commit the wave separately

## 11. Coverage Wave: Cron, Broker, And App Helpers

- [x] 11.1 Add focused tests for `internal/cron/store.go` Ent-backed store and conversion helpers
- [x] 11.2 Add focused tests for remaining `internal/storagebroker/server.go` and `internal/storagebroker/client.go` RPC wrapper branches
- [x] 11.3 Add focused tests for app P2P helper surfaces that do not require external services
- [x] 11.4 Commit the wave separately

## 12. Coverage Wave: Smart Account, Wallet, And Git Bundle

- [x] 12.1 Add focused tests for `internal/wallet/spending.go` limiter behavior and store error paths
- [x] 12.2 Add focused tests for `internal/smartaccount/manager.go` manager helpers and encoding branches
- [x] 12.3 Add focused tests for `internal/p2p/gitbundle/protocol.go` protocol handler dispatch and error branches
- [x] 12.4 Add focused tests for app Smart Account tools and wiring helper branches
- [x] 12.5 Commit the wave separately

## 13. Coverage Wave: Filesystem, Test Utilities, And Knowledge Branches

- [x] 13.1 Add focused tests for `internal/tools/filesystem/filesystem.go` delete, exists, mkdir, copy, and path access branches
- [x] 13.2 Add focused tests for `internal/testutil/mock_session_store.go` CRUD, salt, end, counters, copy semantics, and error injection
- [x] 13.3 Add focused tests for `internal/knowledge/store.go` FTS5 result resolution, learning lookup branches, and low-risk search paths
- [x] 13.4 Commit the wave separately

## 14. Coverage Wave: App P2P, Meta Tools, And Knowledge Store

- [x] 14.1 Add focused tests for deterministic `internal/app/wiring_p2p.go` helper and wiring branches
- [x] 14.2 Add focused tests for deterministic `internal/app/tools_meta.go` tool handlers and helper branches
- [x] 14.3 Add focused tests for remaining low-risk `internal/knowledge/store.go` scoring, stats, external ref, and deletion branches
- [x] 14.4 Run subagent-driven review checkpoints for the wave
- [x] 14.5 Commit the wave separately

## 15. Coverage Wave: CLI Entrypoints, Chat State, Settings Navigation, And Config Redaction

- [x] 15.1 Add focused tests for deterministic `cmd/lango/main.go` helper and adapter branches
- [x] 15.2 Add focused tests for deterministic `internal/cli/chat/chat.go` state, accessor, session-mode, and key-helper branches
- [x] 15.3 Add focused tests for deterministic `internal/cli/configcmd/getset.go` redaction, scalar, reflection, formatting, and set-field branches
- [x] 15.4 Add focused tests for deterministic `internal/cli/settings/editor.go` dependency navigation and setup-flow branches
- [x] 15.5 Run subagent-driven review checkpoints for the wave
- [x] 15.6 Commit the wave separately

## 16. Coverage Wave: Gateway, Telegram, And App Adapters

- [x] 16.1 Add focused tests for deterministic `internal/gateway/server.go` setter, status, RPC, and companion branches
- [x] 16.2 Add focused tests for deterministic `internal/channels/telegram/telegram.go` channel identity, send, split, error formatting, and allow-list branches
- [x] 16.3 Add focused tests for deterministic `internal/app/app.go` publish, post-agent wiring, and lifecycle registration branches
- [x] 16.4 Run subagent-driven review checkpoints for the wave
- [x] 16.5 Commit the wave separately

## 17. Coverage Wave: App Wiring, Knowledge Store, And CLI/ADK Surfaces

- [x] 17.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/tools_meta.go`, `internal/app/modules.go`, and related app wiring branches
- [x] 17.2 Add focused tests for deterministic `internal/knowledge/store.go` search, FTS5, payload-protection, deletion, and learning branches
- [x] 17.3 Add focused tests for deterministic `cmd/lango/main.go`, `internal/cli/chat/chat.go`, and `internal/adk` option/helper branches
- [x] 17.4 Run subagent-driven review checkpoints for the wave
- [x] 17.5 Commit the wave separately

## 18. Coverage Wave: Integrations, Ledgers, And Receipts

- [x] 18.1 Add focused tests for deterministic `internal/channels/discord/discord.go` channel identity, send, split, guild allow-list, and error formatting branches
- [x] 18.2 Add focused tests for deterministic `internal/p2p/gitbundle/bundle.go` diff, apply, hook, and safe-apply branches
- [x] 18.3 Add focused tests for deterministic `internal/runledger/writethrough.go`, `internal/mcp/connection.go`, and `internal/receipts/store.go` branch-heavy helpers
- [x] 18.4 Run subagent-driven review checkpoints for the wave
- [x] 18.5 Commit the wave separately

## 19. Coverage Wave: App, Knowledge, And User-Facing Entrypoints

- [x] 19.1 Add focused tests for deterministic `internal/app` metadata, module, wiring, and app helper branches that do not require live network services
- [x] 19.2 Add focused tests for deterministic `internal/knowledge/store.go` persistence, validation, and branch-heavy query helpers
- [x] 19.3 Add focused tests for deterministic `cmd/lango/main.go` and high-uncovered user-facing CLI helper branches without launching interactive UI loops
- [x] 19.4 Run subagent-driven review checkpoints for the wave
- [x] 19.5 Commit the wave separately

## 20. Coverage Wave: CLI TUI Pages, Settings, And ADK Branches

- [x] 20.1 Add focused tests for deterministic `internal/cli/cockpit/pages/missioncontrol.go` rendering, navigation, filter, and command-creation branches
- [x] 20.2 Add focused tests for deterministic `internal/cli/settings/editor.go` editor state, section navigation, validation, and save/cancel helper branches
- [x] 20.3 Add focused tests for deterministic `internal/adk/agent.go` run helper, collection, diagnostics, and error branch helpers without live provider calls
- [x] 20.4 Run subagent-driven review checkpoints for the wave
- [x] 20.5 Commit the wave separately

## 21. Coverage Wave: Knowledge Atomic Branches And App Wiring Hotspots

- [x] 21.1 Add focused tests for deterministic `internal/knowledge/store.go` atomic payload, FTS5 sync, and error helper branches without relying on generated code coverage
- [x] 21.2 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool handler, dispatcher, and receipt/error branches without live network services
- [x] 21.3 Add focused tests for deterministic `internal/app/modules.go`, `internal/app/app.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` module/wiring helper branches
- [x] 21.4 Run subagent-driven review checkpoints for the wave
- [x] 21.5 Commit the wave separately

## 22. Coverage Wave: CLI Entrypoint, Sandbox Probe, And Skill Importer Branches

- [x] 22.1 Add focused tests for deterministic `cmd/lango/main.go` chat/cockpit/workbench/config command branch helpers without launching interactive UI loops
- [x] 22.2 Add focused tests for deterministic `internal/cli/sandbox/sandbox.go` status collection, test/probe command helpers, and sandbox capability branches without requiring privileged sandbox execution
- [x] 22.3 Add focused tests for deterministic `internal/skill/importer.go` git/http import planning, URL/resource handling, and failure branches without live network access
- [x] 22.4 Run subagent-driven review checkpoints for the wave
- [x] 22.5 Commit the wave separately

## 23. Coverage Wave: Escrow Tools, Economy Wiring, And Chat Approval Branches

- [x] 23.1 Add focused tests for deterministic `internal/app/tools_escrow.go` escrow tool handler success, validation, and error branches without live chain calls
- [x] 23.2 Add focused tests for deterministic `internal/app/wiring_economy.go`, `internal/app/wiring.go`, and related app wiring helper branches without live providers or network services
- [x] 23.3 Add focused tests for deterministic `internal/cli/chat/chat.go` approval/key handling, pending state, and submission guard branches without launching the TUI program
- [x] 23.4 Run subagent-driven review checkpoints for the wave
- [x] 23.5 Commit the wave separately

## 24. Coverage Wave: App Meta, Module Initialization, P2P Wiring, And CLI Entrypoints

- [x] 24.1 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool composition, dependency-disabled, and handler edge branches without live network services
- [x] 24.2 Add focused tests for deterministic `internal/app/modules.go` module initialization, catalog, status, and disabled-dependency branches without live providers
- [x] 24.3 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/wiring.go`, and `cmd/lango/main.go` startup and wiring helper branches without binding external services
- [x] 24.4 Run subagent-driven review checkpoints for the wave
- [x] 24.5 Commit the wave separately

## 25. Coverage Wave: Contract Caller, A2A Card, Cover CLI, And ADK Context Helpers

- [x] 25.1 Add focused tests for deterministic `cmd/lango-cover/main.go` argument, module discovery, threshold, and error branches without external commands
- [x] 25.2 Add focused tests for deterministic `internal/a2a/server.go` agent-card construction, P2P/pricing mutation, route output, and skill extraction branches without network listeners
- [x] 25.3 Add focused tests for deterministic `internal/contract/caller.go` ABI, revert-reason, receipt timeout/context, and local error branches without live RPC services
- [x] 25.4 Add focused tests for deterministic `internal/adk/context_model.go` context item, token estimation, and retrieval merge helper branches without live providers
- [x] 25.5 Run subagent-driven review checkpoints for the wave
- [x] 25.6 Commit the wave separately

## 26. Coverage Wave: Mission Projector, Ontology Service, And App/Main Helper Branches

- [x] 26.1 Add focused tests for deterministic `internal/cli/cockpit/missioncontrol_projector.go` projection helper, status mapping, summary, and enrichment branches without launching the TUI
- [x] 26.2 Add focused tests for deterministic `internal/ontology/service.go` action executor, schema health, type usage, permission, and delegation error branches without external services
- [x] 26.3 Add focused tests for deterministic `internal/app` and `cmd/lango/main.go` helper branches that remain high-uncovered without binding listeners or launching interactive loops
- [x] 26.4 Run subagent-driven review checkpoints for the wave
- [x] 26.5 Commit the wave separately

## 27. Coverage Wave: Meta Tools, Receipts Store, Ontology Tools, And App Wiring Hotspots

- [x] 27.1 Add focused tests for deterministic `internal/app/tools_meta.go` and `internal/app/tools_escrow.go` validation, dispatcher, receipt, and helper branches without live network or chain calls
- [x] 27.2 Add focused tests for deterministic `internal/receipts/store.go` lifecycle, query, filtering, and error branches using in-memory or temp-file state only
- [x] 27.3 Add focused tests for deterministic `internal/ontology/tools.go` tool builder, parameter validation, action/query, and error branches without external services
- [x] 27.4 Add focused tests for deterministic `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/app.go` helper branches without binding listeners or starting P2P services
- [x] 27.5 Run subagent-driven review checkpoints for the wave
- [x] 27.6 Commit the wave separately

## 28. Coverage Wave: Gateway, P2P Protocol, Gemini Provider, And Remaining App Entrypoints

- [x] 28.1 Add focused tests for deterministic `internal/gateway/server.go` response handlers, callbacks, router setup, and connection lifecycle branches without binding real listeners
- [x] 28.2 Add focused tests for deterministic `internal/p2p/protocol/remote_agent.go` request/response, timeout, encoding, and error branches with fake streams/hosts only
- [x] 28.3 Add focused tests for deterministic `internal/provider/gemini/gemini.go` provider identity, schema conversion, request assembly, model listing, and error branches without live Gemini calls
- [x] 28.4 Add focused tests for deterministic `internal/app/tools_p2p.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `cmd/lango/main.go` helper branches without starting P2P services or interactive loops
- [x] 28.5 Run subagent-driven review checkpoints for the wave
- [x] 28.6 Commit the wave separately

## 29. Coverage Wave: App Metadata, Entrypoints, Test Utilities, And Economy Hotspots

- [x] 29.1 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` branches that remain high-uncovered without live providers or listeners
- [x] 29.2 Add focused tests for deterministic `cmd/lango/main.go`, `internal/cli/settings/forms_economy.go`, and `internal/testutil/mock_cron.go` helper branches without launching interactive UI loops or relying on wall-clock scheduling
- [x] 29.3 Add focused tests for deterministic `internal/adk/agent.go`, `internal/bootstrap/phases.go`, and `internal/economy/tools.go` branch-heavy helpers without live model providers, chain RPC, or external services
- [x] 29.4 Run subagent-driven review checkpoints for the wave
- [x] 29.5 Commit the wave separately

## 30. Coverage Wave: Remaining App Hotspots, TUI Form, Smart Account, And Workflow CLI

- [x] 30.1 Add focused tests for remaining deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/app.go` branches that still dominate uncovered statement count without live providers, listeners, or chain RPC
- [x] 30.2 Add focused tests for deterministic `internal/cli/tuicore/form.go`, `internal/cli/smartaccount/session.go`, and `internal/cli/workflow/workflow.go` helper branches without launching interactive UI loops
- [x] 30.3 Add focused tests for remaining deterministic `cmd/lango/main.go` and `internal/adk/agent.go` helper branches that are safe to exercise with fake stores, fake agents, and no external services
- [x] 30.4 Run subagent-driven review checkpoints for the wave
- [x] 30.5 Commit the wave separately

## 31. Coverage Wave: Status, Settings Knowledge, P2P Node, And ADK Support Helpers

- [x] 31.1 Add focused tests for deterministic `internal/cli/settings/forms_knowledge.go` form builders and field wiring without launching interactive UI loops
- [x] 31.2 Add focused tests for deterministic `internal/cli/status/status.go` dead-letter bridge, retry/detail, sanitization, and status helper branches using fake stores and command writers
- [x] 31.3 Add focused tests for deterministic `internal/p2p/node.go` constructor/accessor, key path, handler, and peer-found helper branches without external peers
- [x] 31.4 Add focused tests for deterministic `internal/adk` support helpers such as child session service, context helpers, PII adapter, plugin constructors, and tool adapters without live providers
- [x] 31.5 Run subagent-driven review checkpoints for the wave
- [x] 31.6 Commit the wave separately

## 32. Coverage Wave: Automation Settings Forms And Smart Account CLI Branches

- [x] 32.1 Add focused tests for deterministic `internal/cli/settings/forms_automation.go` form builders, field wiring, defaults, and validation branches without launching interactive UI loops
- [x] 32.2 Add focused tests for deterministic `internal/cli/smartaccount/policy.go` and remaining `internal/cli/smartaccount/session.go` command output, cleanup, validation, and error branches without live chain RPC
- [x] 32.3 Run subagent-driven review checkpoints for the wave
- [x] 32.4 Commit the wave separately

## 33. Coverage Wave: Session Store, Storage Broker, App Runtime Helpers

- [x] 33.1 Add focused tests for deterministic `internal/session/ent_store.go` constructor, client/DB accessor, list, timeout annotation, checksum, and migration branches using temp SQLite state only
- [x] 33.2 Add focused tests for deterministic `internal/storagebroker/server.go` dispatch, payload encode/decode, crypto wrapper, and unavailable dependency branches without binding real listeners
- [x] 33.3 Add focused tests for remaining deterministic `internal/app` and `cmd/lango/main.go` runtime helper branches without live providers, external listeners, P2P services, or interactive UI loops
- [x] 33.4 Run subagent-driven review checkpoints for the wave
- [x] 33.5 Commit the wave separately

## 34. Coverage Wave: App Meta, Network, And Wiring Branches

- [x] 34.1 Add focused tests for deterministic `internal/app/tools_meta.go` meta-tool construction and exportability/payment helper branches without live payments, chain RPC, or background workers
- [x] 34.2 Add focused tests for deterministic `internal/app/modules.go` network/intelligence module disabled and lightweight catalog branches without starting P2P, payment, or external services
- [x] 34.3 Add focused tests for deterministic `internal/app/wiring.go` and `internal/app/wiring_p2p.go` helper branches such as security/provenance/agent options and invalid or disabled P2P configuration paths
- [x] 34.4 Run subagent-driven review checkpoints for the wave
- [x] 34.5 Commit the wave separately

## 35. Coverage Wave: Knowledge Wiring, Smart Account Sessions, And ADK Helpers

- [x] 35.1 Add focused tests for deterministic `internal/app/wiring_knowledge.go` knowledge, FTS5 bulk-indexing, skill registry, and adapter helper branches without live providers or external services
- [x] 35.2 Add focused tests for deterministic `internal/cli/smartaccount/session.go` execution helper error paths, cleanup behavior, list status mapping, revoke branches, command validation, and table formatting without live chain RPC
- [x] 35.3 Add focused tests for deterministic `internal/adk` low-risk helper branches that remain high-uncovered without live providers, including context retrieval/formatting and session-service discard/cleanup helpers
- [x] 35.4 Run subagent-driven review checkpoints for the wave
- [x] 35.5 Commit the wave separately

## 36. Coverage Wave: Security Status, P2P Handshake, And CLI Helpers

- [x] 36.1 Add focused tests for deterministic `internal/cli/security/status.go` status helpers, envelope/identity rendering, config fallback, and non-interactive DB-status branches without interactive passphrase prompts
- [x] 36.2 Add focused tests for deterministic `internal/p2p/handshake/handshake.go` stream handler, protocol selection, timestamp, signer, and early failure branches without external peers
- [x] 36.3 Add focused tests for deterministic `cmd/lango/main.go` command startup/config helper branches without launching interactive TUI loops or binding external services
- [x] 36.4 Run subagent-driven review checkpoints for the wave
- [x] 36.5 Commit the wave separately

## 37. Coverage Wave: FTS5, Bootstrap Phases, And App Post-Agent Wiring

- [x] 37.1 Add focused tests for deterministic `internal/search/fts5.go` table lifecycle, insert/update/delete, bulk insert, search, DB accessor, and query sanitization branches using temp SQLite where FTS5 is available, plus fallback SQL-path/error/sanitization coverage when the local SQLite runtime lacks FTS5
- [x] 37.2 Add focused tests for deterministic `internal/bootstrap/phases.go` data-dir, encryption detection, credential acquisition, database-open, security-state, crypto-init, and profile-load branches without interactive prompts
- [x] 37.3 Add focused tests for deterministic `internal/app/app.go` and post-agent P2P wiring branches covering memory/turn callbacks, provenance options, security init, and inbound approval wiring without live providers or external peers
- [x] 37.4 Run subagent-driven review checkpoints for the wave
- [x] 37.5 Commit the wave separately

## 38. Coverage Wave: CLI Settings, Workflow, Config, And App Wiring Hotspots

- [x] 38.1 Add focused tests for deterministic `internal/cli/settings/forms_agent.go` form construction, defaults, validation, dependency wiring, and provider/model field branches without launching interactive UI loops
- [x] 38.2 Add focused tests for deterministic `internal/cli/workflow/workflow.go` command construction, output formatting, validation, and error branches using fake writers and no external services
- [x] 38.3 Add focused tests for deterministic `internal/cli/configcmd/getset.go` scalar/reflection formatting, redaction, invalid path, and set-field branches without interactive prompts
- [x] 38.4 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, `internal/app/wiring.go`, and `internal/app/wiring_p2p.go` helper branches that still dominate uncovered statement count without live providers, listeners, or external peers
- [x] 38.5 Run subagent-driven review checkpoints for the wave
- [x] 38.6 Commit the wave separately

## 39. Coverage Wave: Workspace Tools, Run Ledger, And Entrypoint Helpers

- [x] 39.1 Add focused tests for deterministic `internal/app/tools_workspace.go` create/join/leave/list/status helper branches without live P2P peers or external services
- [x] 39.2 Add focused tests for deterministic `internal/runledger/ent_store.go` journal, snapshot, cache, and error branches using isolated test storage
- [x] 39.3 Add focused tests for deterministic `cmd/lango/main.go` chat/cockpit/workbench/config helper branches through existing seams without launching interactive UI loops
- [x] 39.4 Run subagent-driven review checkpoints for the wave
- [x] 39.5 Commit the wave separately

## 40. Coverage Wave: USDC Settlement, App Modules, And Chat State

- [x] 40.1 Add focused tests for deterministic `internal/economy/escrow/usdc_settler.go` address resolution, lock, release/refund, signing, submission retry, and receipt confirmation branches using in-process RPC seams only
- [x] 40.2 Add focused tests for deterministic `internal/app/modules.go` foundation/intelligence/automation/network/extension module branches that do not start live external services
- [x] 40.3 Add focused tests for deterministic `internal/cli/chat/chat.go` state/update/render helper branches without launching Bubble Tea programs
- [x] 40.4 Run subagent-driven review checkpoints for the wave
- [x] 40.5 Commit the wave separately

## 41. Coverage Wave: Test Utilities, Smart Account Policy, And P2P Firewall

- [x] 41.1 Add focused tests for deterministic `internal/testutil/mock_graph.go` CRUD, query, traversal, copy semantics, error injection, counters, clear, and close branches
- [x] 41.2 Add focused tests for deterministic `internal/cli/smartaccount/policy.go` show/set command formatting, output mode validation, seam errors, cleanup, value formatting, and limit parsing branches without live chain RPC
- [x] 41.3 Add focused tests for deterministic `internal/p2p/firewall/firewall.go` ACL validation, rule mutation, rate limiting, reputation outcomes, sanitization, attestation, matching, and copy semantics without live network services
- [x] 41.4 Run subagent-driven review checkpoints for the wave
- [x] 41.5 Commit the wave separately

## 42. Coverage Wave: Receipts, Proactive Librarian, And MCP Connection

- [x] 42.1 Add focused tests for deterministic `internal/receipts/store.go` remaining transaction lifecycle, clone/copy, event trail, validation, and error branches without external services
- [x] 42.2 Add focused tests for deterministic `internal/librarian/proactive_buffer.go` default config, trigger processing, message/observation errors, threshold/cooldown/max-pending behavior, auto-save, dual-save, event publishing, and inquiry creation branches without background sleeps
- [x] 42.3 Add focused tests for deterministic `internal/mcp/connection.go` state/accessor/copy semantics, sandbox decision publishing, disconnect, health-check, timeout, and transport error branches without launching real MCP servers
- [x] 42.4 Run subagent-driven review checkpoints for the wave
- [x] 42.5 Commit the wave separately

## 43. Coverage Wave: App Entrypoints, Channels, Smart Account Wiring, And Turn Runner

- [x] 43.1 Add focused tests for deterministic `internal/app/channels.go` and `internal/app/wiring_smartaccount.go` helper branches without launching external channel services, chain RPC, or background workers
- [x] 43.2 Add focused tests for deterministic `internal/turnrunner/runner.go` lifecycle, event emission, error, cancellation, receipt, and callback branches using fake agents and stores only
- [x] 43.3 Add focused tests for deterministic `cmd/lango/main.go` command startup/config helper branches without launching interactive TUI loops or binding listeners
- [x] 43.4 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/modules.go`, and `internal/app/wiring_p2p.go` branches that remain high-uncovered without live providers, listeners, or external peers
- [x] 43.5 Run subagent-driven review checkpoints for the wave
- [x] 43.6 Commit the wave separately

## 44. Coverage Wave: App Wiring, ADK/Knowledge, Deadletters, And Slack

- [x] 44.1 Add focused tests for deterministic `internal/app/wiring_p2p.go` and related `internal/app/app.go` helper branches without live listeners, peers, or P2P services
- [x] 44.2 Add focused tests for deterministic `internal/app/modules.go` and `internal/app/wiring.go` helper branches without live providers, background workers, or external services
- [x] 44.3 Add focused tests for deterministic `internal/adk/agent.go` and `internal/knowledge/store.go` remaining helper/error branches without live model providers or external services
- [x] 44.4 Add focused tests for deterministic `internal/cli/cockpit/pages/deadletters.go` and `internal/channels/slack/slack.go` branches without launching TUI loops or live Slack clients
- [x] 44.5 Run subagent-driven review checkpoints for the wave
- [x] 44.6 Commit the wave separately

## 45. Coverage Wave: ZKP, Storage Facade, Escrow Monitor, And App Helpers

- [x] 45.1 Add focused tests for deterministic `internal/p2p/zkp/zkp.go` branches without external provers, peers, or network services
- [x] 45.2 Add focused tests for deterministic `internal/storage/facade.go` resolution, fallback, and error branches without external storage services
- [x] 45.3 Add focused tests for deterministic `internal/economy/escrow/hub/monitor.go` event handling, replay, state, and error branches without live chain RPC
- [x] 45.4 Add focused tests for remaining deterministic `internal/app/tools_meta.go`, `internal/app/wiring_p2p.go`, `internal/app/modules.go`, and `internal/app/wiring.go` branches that do not require live providers, listeners, peers, or background workers
- [x] 45.5 Run subagent-driven review checkpoints for the wave
- [x] 45.6 Keep Ent-backed Wave 45 test bootstraps race-safe with the shared schema serialization seam and generated enttest template
- [x] 45.7 Commit the wave separately

## 46. Verification And Enforcement

## 46. Coverage Wave: Entrypoints, Payment, Storage Broker, And App Wiring

- [x] 46.1 Add focused tests for deterministic `cmd/lango/main.go` command entrypoint branches that avoid launching interactive TUI loops or real servers
- [x] 46.2 Add focused tests for deterministic `internal/payment/service.go` send, balance, retry, confirmation, and failure-recording branches without live RPC
- [x] 46.3 Add focused tests for deterministic `internal/storagebroker/server.go` request dispatch, config, payment, recall, workflow, and error branches without external brokers
- [x] 46.4 Add focused tests for deterministic `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/wiring_p2p.go`, and `internal/app/tools_meta.go` branches that can be covered without live providers, listeners, peers, or long-running workers
- [x] 46.5 Run subagent-driven review checkpoints for the wave
- [x] 46.6 Commit the wave separately

## 47. Coverage Wave: Gateway Auth, App P2P Tools, ADK Agent, And CLI Workflow

- [x] 47.1 Add focused tests for deterministic `internal/gateway/auth.go` authentication middleware branches without real gateway servers
- [x] 47.2 Add focused tests for deterministic `internal/app/tools_p2p.go` and remaining safe `internal/app/wiring_p2p.go` branches without live peers or listeners
- [x] 47.3 Add focused tests for deterministic `internal/adk/agent.go` execution, callback, and error branches without live providers
- [x] 47.4 Add focused tests for deterministic `internal/cli/workflow/workflow.go` command, validation, and output branches without external services
- [x] 47.5 Run subagent-driven review checkpoints for the wave
- [x] 47.6 Commit the wave separately

## 48. Coverage Wave: Knowledge, App Lifecycle, Provider, And RunLedger

- [x] 48.1 Add focused tests for deterministic `internal/knowledge/store.go` branches using in-memory SQLite/FTS fixtures without external services
- [x] 48.2 Add focused tests for deterministic `internal/app/app.go`, `internal/app/wiring.go`, and remaining safe app lifecycle branches without live listeners, peers, or providers
- [x] 48.3 Add focused tests for deterministic `internal/provider/anthropic/anthropic.go` request building, response mapping, and error handling without live Anthropic API calls
- [x] 48.4 Add focused tests for deterministic `internal/runledger/ent_store.go` and related run ledger storage branches using local Ent test stores
- [x] 48.5 Run subagent-driven review checkpoints for the wave
- [x] 48.6 Commit the wave separately

## 49. Coverage Wave: App Wiring, MCP Manager, And P2P Workspace

- [x] 49.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/wiring.go`, and `internal/app/modules.go` branches without live network peers
- [x] 49.2 Add focused tests for deterministic `internal/app/tools_meta.go`, `internal/app/tools_p2p.go`, and safe tool wrapper branches without external RPC calls
- [x] 49.3 Add focused tests for deterministic `internal/mcp/manager.go` lifecycle, registration, and error branches without external MCP servers
- [x] 49.4 Add focused tests for deterministic `internal/p2p/workspace/gossip.go` publish/subscribe/error branches using local fakes
- [x] 49.5 Run subagent-driven review checkpoints for the wave
- [x] 49.6 Commit the wave separately

## 50. Coverage Wave: ADK, Workflow, Contract, Payment, And Tool Surfaces

- [x] 50.1 Add focused tests for deterministic `internal/adk/context_model.go` and `internal/adk/tools.go` helper branches without live model providers
- [x] 50.2 Add focused tests for deterministic `internal/workflow/state.go` and `internal/workflow/tools.go` persistence/query/tool branches using local Ent test stores or fakes only
- [x] 50.3 Add focused tests for deterministic `internal/contract/caller.go` read/write helper branches and revert/receipt handling without live RPC services
- [x] 50.4 Add focused tests for deterministic `internal/tools/payment/payment.go`, `internal/tools/filesystem/tools.go`, `internal/tools/websearch/tools.go`, `internal/tools/webfetch/readability.go`, and `internal/x402` helper branches without external network calls
- [x] 50.5 Add focused tests for deterministic `cmd/lango/main.go`, `cmd/lango/dead_letter_status.go`, and `internal/turntrace/retention.go` helper branches without launching interactive TUI loops or long-running workers
- [x] 50.6 Run subagent-driven review checkpoints for the wave
- [x] 50.7 Commit the wave separately

## 51. Coverage Wave: App Hotspots, Gateway, CLI Workflow, And Bundler

- [x] 51.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, `internal/app/wiring.go`, and `internal/app/app.go` branches that remain high-uncovered without live providers, listeners, peers, or long-running workers
- [x] 51.2 Add focused tests for deterministic `internal/gateway/server.go` chat/RPC/startup branches using local fakes and no externally bound services
- [x] 51.3 Add focused tests for deterministic `internal/smartaccount/bundler/client.go` nonce, gas fee, encoding, response, and error branches using local HTTP/RPC fakes only
- [x] 51.4 Add focused tests for deterministic `cmd/lango/main.go` and `internal/cli/workflow/workflow.go` command/helper branches without launching interactive TUI loops
- [x] 51.5 Add focused tests for remaining deterministic `internal/knowledge/store.go` write, atomic, relevance, and FTS cleanup branches using local SQLite/FTS fixtures only
- [x] 51.6 Run subagent-driven review checkpoints for the wave
- [x] 51.7 Commit the wave separately

## 52. Coverage Wave: App Wiring, CLI Hotspots, Ontology, And X402

- [x] 52.1 Add focused tests for deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/tools_meta.go`, and `internal/app/wiring.go` branches that remain top uncovered without live providers, listeners, peers, or long-running workers
- [x] 52.2 Add focused tests for deterministic `internal/app/tools_p2p.go`, `internal/app/wiring_economy.go`, and `internal/app/wiring_knowledge.go` helper/error branches without live P2P peers, chain RPC, or external services
- [x] 52.3 Add focused tests for deterministic `internal/cli/chat/chat.go` and `internal/cli/configcmd/getset.go` helper branches without launching interactive TUI loops
- [x] 52.4 Add focused tests for deterministic `internal/ontology/service.go`, `internal/x402`, and adjacent payment/settlement helper branches without external network or chain services
- [x] 52.5 Run subagent-driven review checkpoints for the wave
- [x] 52.6 Commit the wave separately

## 53. Coverage Wave: App Residuals, Post-Adjudication Status, Browser, Smart Account, And Graph Store

- [x] 53.1 Add focused tests for remaining deterministic `internal/app/wiring_p2p.go`, `internal/app/modules.go`, `internal/app/wiring.go`, `internal/app/app.go`, and `internal/app/wiring_economy.go` branches plus the related `internal/economy/negotiation` callback lock-safety branch that can be exercised without live providers, listeners, peers, chain RPC, or long-running workers
- [x] 53.2 Add focused tests for deterministic `internal/postadjudicationstatus/service.go` and `internal/graph/bolt_store.go` status aggregation, receipt, filtering, snapshot, stats, scan, and clear branches using local stores only
- [x] 53.3 Add focused tests for deterministic `internal/tools/browser/browser.go` and `internal/cli/smartaccount/policy.go` session/error/helper branches without launching a real browser or live chain RPC
- [x] 53.4 Run subagent-driven review checkpoints for the wave
- [x] 53.5 Commit the wave separately

## 54. Verification And Enforcement

- [ ] 54.1 Run `go build ./...`
- [ ] 54.2 Run `go test ./...`
- [ ] 54.3 Run the non-generated coverage report and confirm coverage is at least 90%
- [ ] 54.4 Run the executable 90% coverage gate and confirm it passes
- [ ] 54.5 Run `git diff --check`
- [ ] 54.6 Run subagent-driven spec and code-quality review checkpoints
- [ ] 54.7 Run `openspec validate --all --strict`
- [ ] 54.8 Archive the completed OpenSpec change
