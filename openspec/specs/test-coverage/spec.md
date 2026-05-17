# Spec: Test Coverage

## Overview
Comprehensive test suite covering core business logic, security tools, and existing package enhancements for the Lango project.

## Purpose

Capability spec for test-coverage. See requirements below for scope and behavior contracts.
## Requirements

### REQ-1: Knowledge Package Tests
The `internal/knowledge/` package SHALL have test coverage for all Store CRUD operations and the ContextRetriever.

#### Scenario: Knowledge CRUD and retrieval paths are covered
- **WHEN** the knowledge package test suite runs
- **THEN** it SHALL cover entry create, upsert, get, search, delete, rate limiting, context retrieval, keyword extraction, and prompt assembly

### REQ-2: Learning Package Tests
The `internal/learning/` package SHALL have test coverage for error pattern analysis and the learning engine.

#### Scenario: Learning analysis and engine paths are covered
- **WHEN** the learning package test suite runs
- **THEN** it SHALL cover error-pattern extraction, categorization, wrapped deadline detection, parameter summaries, audit logging, learning creation, confidence boosts, known-fix retrieval, and user-correction recording

### REQ-3: Skill Package Tests
The `internal/skill/` package SHALL have test coverage for building, executing, and managing skills.

#### Scenario: Skill build, execution, and registry flows are covered
- **WHEN** the skill package test suite runs
- **THEN** it SHALL cover composite, script, and template skill construction, script safety validation, execution paths, registry validation, and active-skill tool exposure

### REQ-4: Security Tool Tests
The `internal/tools/crypto/` and `internal/tools/secrets/` packages SHALL have test coverage.

#### Scenario: Crypto and secrets tool paths are covered
- **WHEN** the security-tool test suite runs
- **THEN** it SHALL cover hashing, encrypt/decrypt round trips, signing, key listing, secret lifecycle operations, and secret upsert behavior

### REQ-5: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Existing package tests gain targeted enhancements
- **WHEN** the enhanced regression suite runs
- **THEN** it SHALL cover session-store CRUD and TTL behavior, anthropic model listing, openai unavailable-server handling, and app startup failure modes

### REQ-6: Channel Mock Thread Safety
Channel test mock types SHALL use mutex synchronization to protect shared slices from concurrent access by handler goroutines and test assertions.

#### Scenario: Channel mock slices are protected and copied safely
- **WHEN** channel handler goroutines append to mock slices while tests read them
- **THEN** mutex synchronization SHALL serialize concurrent access
- **AND** helper accessors SHALL return defensive copies of the slice data

### Requirement: Shared CLI loader helpers remain available
The shared `internal/testutil` package SHALL continue to provide lightweight config-loader and bootstrap-loader helpers for CLI regression tests after the global stdout interception harness is removed.

#### Scenario: CLI config loader tests still compile
- **WHEN** CLI regression tests construct commands that accept `func() (*config.Config, error)` dependencies
- **THEN** `internal/testutil` SHALL provide helper loaders that return a supplied config or error

#### Scenario: CLI bootstrap loader tests still compile
- **WHEN** CLI regression tests construct commands that accept `func() (*bootstrap.Result, error)` dependencies
- **THEN** `internal/testutil` SHALL provide helper loaders that return a supplied bootstrap result or error
- **AND** those helpers SHALL NOT require the removed global stdout interception harness

### Requirement: Workflow run schedule regressions remain deterministic
Workflow run command regressions that depend on package-global execution seams SHALL avoid parallel execution patterns that can leak stubbed state across tests.

#### Scenario: Direct-execution seam override does not affect sibling tests
- **WHEN** a workflow run regression temporarily replaces the package-global direct-execution seam
- **THEN** sibling workflow run regressions SHALL not observe that override unexpectedly
- **AND** repository-wide test runs SHALL remain deterministic

### Requirement: Exec warning seam regressions remain deterministic
Tests that temporarily replace the package-global `execWarningWriter` SHALL avoid parallel execution so suite results do not depend on test scheduling.

#### Scenario: Warning writer seam test does not race
- **WHEN** the exec warning regression temporarily replaces `execWarningWriter`
- **THEN** sibling tests SHALL not be able to observe that replacement concurrently

### Requirement: Collaboration runtime seam regressions remain deterministic
Tests that temporarily replace the package-global collaboration runtime `eventTime` seam SHALL avoid parallel execution so sibling tests cannot observe the override unexpectedly.

#### Scenario: Event-time seam override does not leak
- **WHEN** a collaboration runtime regression overrides `eventTime`
- **THEN** repository-wide test runs SHALL not depend on whether sibling tests were scheduled at the same time

### Requirement: Mission-control projector time-seam regressions remain deterministic
Tests that temporarily replace the mission-control projector `nowFn` seam SHALL avoid parallel execution so sibling tests cannot observe the override unexpectedly.

#### Scenario: nowFn seam override does not leak
- **WHEN** a mission-control projector regression overrides `nowFn`
- **THEN** repository-wide test runs SHALL not depend on whether sibling tests were scheduled at the same time

### Requirement: Mission-control projector time-sensitive regressions stay explicitly serialized
Mission-control projector regressions that freeze `projector.nowFn` SHALL remain explicitly documented as non-parallel so future maintenance does not accidentally reintroduce scheduler-dependent flakes.

#### Scenario: nowFn-freezing tests document serialization intent
- **WHEN** a maintainer reads the projector tests that freeze `nowFn`
- **THEN** the file SHALL make clear that those tests intentionally avoid parallel execution because they assert time-sensitive projection behavior

### Requirement: Main spec hygiene guards stay executable
Repository-level quality regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Main spec purpose placeholders are rejected
- **WHEN** a main OpenSpec spec reintroduces archive-generated placeholder purpose text
- **THEN** an executable repository test SHALL fail

#### Scenario: Public doc shared-confirm punctuation regressions are rejected
- **WHEN** a public doc or README reintroduces stale shared confirmation examples without the colon separator
- **THEN** an executable repository test SHALL fail

### Requirement: Public doc shared-confirm punctuation guards stay executable
Repository-level docs-quality regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Public doc shared-confirm punctuation regressions are rejected
- **WHEN** a public doc or README reintroduces stale shared confirmation examples without the colon separator
- **THEN** an executable repository test SHALL fail

### Requirement: Public config CLI docs guards stay executable
Repository-level config-doc regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Public config CLI example regressions are rejected
- **WHEN** a public doc or README reintroduces stale `config get --format json` examples or profile-less config export/import examples
- **THEN** an executable repository test SHALL fail

#### Scenario: Public config CLI completeness regressions are rejected
- **WHEN** a public doc or README drops one of the implemented `config get`, `config set`, or `config keys` command references
- **THEN** an executable repository test SHALL fail

### Requirement: CLI index docs guards stay executable
Repository-level CLI index regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: CLI index completeness and table-structure regressions are rejected
- **WHEN** public CLI quick reference docs drop implemented operator commands or reintroduce prose that splits the Agent & Memory table
- **THEN** an executable repository test SHALL fail

### Requirement: Production context-placeholder guards stay executable
Repository-level production-code hygiene regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Production context.TODO regressions are rejected
- **WHEN** a non-test Go file under `cmd/` or `internal/` reintroduces `context.TODO()`
- **THEN** an executable repository test SHALL fail

### Requirement: CLI harness hygiene guards stay executable
Repository-level CLI test-harness regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: CLI harness regressions are rejected
- **WHEN** a CLI test reintroduces process-global stdio replacement or legacy shared exec helpers
- **THEN** an executable repository test SHALL fail

### Requirement: Repository test-harness guards stay executable
Repository-level test-harness regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Repository test harness regressions are rejected
- **WHEN** a repository test reintroduces global stdio reassignment or legacy shared exec helpers
- **THEN** an executable repository test SHALL fail

### Requirement: CLI production stream guards stay executable
Repository-level CLI production stream regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: CLI production stream regressions are rejected
- **WHEN** CLI production code reintroduces raw print calls or forbidden direct standard-stream references
- **THEN** an executable repository test SHALL fail

### Requirement: TUI runtime status coverage stays executable
Repository-level regressions in chat slash-command runtime status rendering SHALL be enforced by executable tests.

#### Scenario: TUI status MCP runtime coverage stays executable
- **WHEN** the chat slash-command status surface can receive an active MCP runtime snapshot
- **THEN** executable chat tests SHALL fail if `/status` still labels MCP as configured but inactive in TUI mode
- **AND** executable chat tests SHALL fail if configured-only MCP is indistinguishable from active MCP

### Requirement: Focused chat setup readiness coverage stays executable
Repository-level regressions in focused chat first-run readiness behavior SHALL be enforced by executable tests.

#### Scenario: Focused chat setup readiness coverage blocks regressions
- **WHEN** focused chat is constructed with an incomplete default config
- **THEN** executable chat tests SHALL fail if the shell renders a ready/send state
- **AND** executable chat tests SHALL fail if normal input reaches the turn runner before setup is ready
- **AND** executable chat tests SHALL fail if slash commands are unavailable before setup is ready

### Requirement: Docs guard covers bare root non-interactive contract
Executable docs quality coverage SHALL fail when public CLI docs omit the bare-root non-interactive help fallback contract.

#### Scenario: Public docs guard checks bare root fallback
- **WHEN** docs quality tests run
- **THEN** README, `docs/cli/index.md`, and `docs/cli/core.md` SHALL be checked for the interactive bare-root launch contract
- **AND** they SHALL be checked for the non-interactive help fallback contract
- **AND** they SHALL be checked for the distinction from `lango cockpit` and `lango chat` non-interactive behavior

### Requirement: Background CLI boundary guards stay executable
Repository-level regressions in background CLI boundary messaging SHALL be enforced by executable tests.

#### Scenario: Root bg boundary and docs guards reject misleading references
- **WHEN** root CLI bg commands are wired without an in-process manager
- **THEN** executable tests SHALL fail if the error implies `lango serve` alone makes standalone `lango bg` work
- **AND** executable docs guards SHALL fail if public docs list `lango bg` commands without the in-memory/root-CLI boundary caveat

#### Scenario: Background automation docs guard rejects read-only CLI wording
- **WHEN** docs quality tests run
- **THEN** `docs/automation/background.md` SHALL fail the test suite if it describes all `lango bg` CLI commands as read-only
- **AND** it SHALL be checked for wording that distinguishes inspect-only commands from `lango bg cancel <id>` requesting cancellation in the target gateway process

#### Scenario: Root bg gateway client coverage stays executable
- **WHEN** root CLI bg commands are wired
- **THEN** executable tests SHALL fail if root `lango bg` falls back to the obsolete standalone in-memory boundary stub
- **AND** executable tests SHALL fail if `--addr` is ignored by the gateway-backed bg client

#### Scenario: Background gateway route coverage stays executable
- **WHEN** background gateway route tests run
- **THEN** they SHALL fail if list/status/result/cancel routes are not registered
- **AND** they SHALL fail if unavailable background managers do not return `503`
- **AND** they SHALL fail if task not-found responses are not reported as non-2xx errors

#### Scenario: Background automation docs guard checks gateway wording
- **WHEN** docs quality tests run
- **THEN** README, `docs/cli/index.md`, and `docs/automation/background.md` SHALL fail the test suite if they describe root `lango bg` as disconnected from gateway management after the gateway-backed client is implemented
- **AND** they SHALL be checked for the in-memory restart caveat, `--addr` override guidance, and auth-enabled gateway rejection caveat

### Requirement: P2P connect timeout coverage stays executable
Executable tests SHALL cover `lango p2p connect` command-context cancellation, timeout selection, and cleanup on connect failure.

#### Scenario: P2P connect context coverage blocks regressions
- **WHEN** P2P connect tests run
- **THEN** they SHALL fail if connect uses `context.Background()` instead of a command-derived context
- **AND** they SHALL fail if configured positive `p2p.handshakeTimeout` is not used
- **AND** they SHALL fail if a shorter parent command deadline is reported as the configured timeout
- **AND** they SHALL fail if an earlier configured timeout is reported as a later parent command deadline
- **AND** they SHALL fail if the 30 second fallback is not used for invalid timeout values
- **AND** they SHALL fail if cleanup is skipped after connect failure

### Requirement: P2P connect docs guard stays executable
Executable docs quality coverage SHALL fail when public P2P CLI docs omit the bounded connect timeout and cancellation contract.

#### Scenario: P2P docs guard checks connect timeout contract
- **WHEN** docs quality tests run
- **THEN** `docs/cli/p2p.md` SHALL be checked for the `p2p.handshakeTimeout` connect timeout contract
- **AND** it SHALL be checked for the 30 second fallback
- **AND** it SHALL be checked for command cancellation behavior

### Requirement: P2P CLI node startup context coverage stays executable
Executable tests SHALL cover command-context propagation into shared P2P CLI node startup.

#### Scenario: P2P CLI startup context tests block regressions
- **WHEN** P2P CLI tests run
- **THEN** they SHALL fail if representative status, peers, discover, identity, session, connect, or disconnect paths detach ephemeral node startup from `cmd.Context()`
- **AND** they SHALL fail if `initP2PDeps` starts `internal/p2p.Node` without passing the caller context
- **AND** they SHALL fail if ephemeral P2P CLI cleanup returns before startup worker goroutines registered with the startup wait group finish

### Requirement: P2P CLI docs guard covers startup cancellation
Executable docs quality coverage SHALL fail when public P2P CLI docs omit the command-scoped ephemeral node startup cancellation contract.

#### Scenario: P2P docs guard checks startup cancellation contract
- **WHEN** docs quality tests run
- **THEN** `docs/cli/p2p.md` SHALL be checked for the command-scoped ephemeral node startup cancellation contract

### Requirement: CLI pretty-JSON writer guards stay executable
Repository-level CLI pretty-JSON writer regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Duplicate CLI pretty-JSON writer setups are rejected
- **WHEN** CLI production code reintroduces direct pretty-JSON indentation setup outside the shared CLI JSON helper
- **THEN** an executable repository test SHALL fail

### Requirement: CLI output-flag contract guards stay executable
Repository-level CLI output-flag regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Boolean `--output` flags are rejected
- **WHEN** CLI production code reintroduces a boolean `--output` flag declaration
- **THEN** an executable repository test SHALL fail

### Requirement: Contract CLI docs output-contract guards stay executable
Repository-level contract CLI docs regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Stale contract output docs are rejected
- **WHEN** public contract CLI docs or the main contract interaction spec reintroduce boolean `--output` docs or bare `--output` examples without an explicit format
- **THEN** an executable repository test SHALL fail

### Requirement: Migrated CLI JSON-flag guards stay executable
Repository-level regressions that reintroduce boolean `--json` flags into CLI families already migrated to explicit output-format contracts SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Migrated command families reject boolean `--json` regressions
- **WHEN** production CLI code under migrated families reintroduces a boolean `--json` flag declaration
- **THEN** an executable repository test SHALL fail

#### Scenario: Migrated payment CLI stays covered
- **WHEN** the payment CLI family has already migrated to `--output table|json`
- **THEN** the executable migrated-family guard SHALL continue to cover `internal/cli/payment`

#### Scenario: Migrated graph and security CLI stay covered
- **WHEN** the graph and security inspection CLI families have already migrated to `--output table|json`
- **THEN** the executable migrated-family guard SHALL continue to cover `internal/cli/graph` and `internal/cli/security`

#### Scenario: Migrated agent inspection subset stays covered
- **WHEN** the agent inspection subset (`status`, `list`, `tools`, `hooks`) has already migrated to `--output table|json`
- **THEN** an executable repository test SHALL continue to reject boolean `--json` flag regressions in those specific files

#### Scenario: Migrated P2P operator family stays covered
- **WHEN** the P2P operator family (`status`, `peers`, `identity`, `discover`, `firewall list`, `pricing`, `reputation`, `session list`, `team list`, `team status`, `zkp status`, `zkp circuits`, `workspace create`, `workspace list`, `workspace status`, `git log`) has already migrated to `--output table|json`
- **THEN** an executable repository test SHALL continue to reject boolean `--json` flag regressions in those specific files

### Requirement: Migrated CLI docs output-contract guards stay executable
Repository-level docs regressions that reintroduce stale `--json` UX for migrated CLI families SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Migrated command-family docs reject stale `--json` regressions
- **WHEN** public docs or main specs for migrated command families reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail

#### Scenario: Migrated payment CLI docs stay covered
- **WHEN** the payment CLI family has already migrated to `--output table|json`
- **THEN** the executable migrated-family docs guard SHALL continue to cover payment CLI docs and specs

#### Scenario: Migrated graph and security CLI docs stay covered
- **WHEN** the graph and security inspection CLI families have already migrated to `--output table|json`
- **THEN** the executable migrated-family docs guard SHALL continue to cover graph and security docs and specs

#### Scenario: Migrated agent inspection docs stay covered
- **WHEN** the agent inspection subset (`status`, `list`, `tools`, `hooks`) has already migrated to `--output table|json`
- **THEN** the executable migrated-family docs guard SHALL continue to cover those public docs and the main agent inspection spec

#### Scenario: Migrated P2P operator docs stay covered
- **WHEN** the P2P operator family (`status`, `peers`, `identity`, `discover`, `firewall list`, `pricing`, `reputation`, `session list`, `team list`, `team status`, `zkp status`, `zkp circuits`, `workspace create`, `workspace list`, `workspace status`, `git log`) has already migrated to `--output table|json`
- **THEN** an executable repository test SHALL continue to reject stale `--json` docs and spec regressions for that subset

### Requirement: Sandbox worker exit-code regressions stay executable
Executable tests SHALL cover sandbox worker exit-code behavior without intercepting `os.Exit`.

#### Scenario: Worker protocol exit-code paths are covered
- **WHEN** sandbox worker tests run
- **THEN** they SHALL exercise malformed input, unregistered tool, tool error, and successful tool paths
- **AND** they SHALL assert returned exit codes and JSON results without process-global exit interception

### Requirement: Cmd entrypoint stream guards stay executable
Repository-level top-level entrypoint stream regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Cmd entrypoint stream regressions are rejected
- **WHEN** `cmd/` production code reintroduces raw print calls or forbidden direct standard-stream references
- **THEN** an executable repository test SHALL fail

### Requirement: Cmd entrypoint exit guards stay executable
Repository-level top-level entrypoint exit regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Cmd entrypoint exit regressions are rejected
- **WHEN** `cmd/` production code reintroduces direct `os.Exit(...)` references outside explicit seam declarations
- **THEN** an executable repository test SHALL fail

### Requirement: P2P skills spec truthfulness guard stays executable
Repository-level regressions that reintroduce stale embedded-P2P-skill claims into the `p2p-skills` main spec SHALL be enforced by an executable test instead of relying only on manual review.

#### Scenario: Stale embedded P2P skill claims are rejected
- **WHEN** the repository still ships only the placeholder embedded skill scaffold
- **THEN** an executable repository test SHALL fail if the `p2p-skills` main spec claims specific `skills/p2p-*/SKILL.md` files already exist

### Requirement: CLI test harness spec truthfulness guard stays executable
Repository-level regressions that reintroduce deleted shared-harness path claims into the `cli-test-harness` main spec SHALL be enforced by an executable test.

#### Scenario: Stale deleted harness path claims are rejected
- **WHEN** the current reusable CLI test helpers live in `internal/testutil/loaders.go` and `internal/testutil/helpers.go`
- **THEN** an executable repository test SHALL fail if the `cli-test-harness` main spec claims `internal/testutil/cli_harness.go` is the current shared harness implementation

### Requirement: Economy tool-builder path guards stay executable
Repository-level regressions that reintroduce deleted app-local economy tool-builder path claims into sentinel or on-chain escrow specs SHALL be enforced by an executable test.

#### Scenario: Deleted economy tool-builder path claims are rejected
- **WHEN** sentinel tools are registered by `internal/economy/escrow/sentinel/tools.go` and economy tools come from `internal/economy/tools.go`
- **THEN** an executable repository test SHALL fail if specs claim `internal/app/tools_sentinel.go` or `tools_economy.go` is the current source of truth

### Requirement: Agent memory builder-path guard stays executable
Repository-level regressions that reintroduce deleted app-local agent memory tool-builder path claims into the `agent-memory` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted agent memory builder-path claims are rejected
- **WHEN** agent memory tools are owned by `internal/agentmemory/tools.go` and wired from the current app module
- **THEN** an executable repository test SHALL fail if the `agent-memory` main spec claims `internal/app/tools_agentmemory.go` is the current source of truth

### Requirement: Package-consolidation deleted-path guard stays executable
Repository-level regressions that reintroduce deleted consolidated package-directory paths into the `package-consolidation` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted consolidated package paths are rejected
- **WHEN** the current packages already live in `internal/types`, `internal/security/passphrase`, and `internal/p2p/zkp`
- **THEN** an executable repository test SHALL fail if the `package-consolidation` main spec claims `internal/ctxutil/`, `internal/passphrase/`, or `internal/zkp/` are still the current package locations

### Requirement: Known broken single-file path guards stay executable
Repository-level regressions that reintroduce specific broken single-file references into main specs SHALL be enforced by an executable test.

#### Scenario: Known broken single-file references are rejected
- **WHEN** shared-types, skill-runtime-v2, x402-v2, phantom-feature-wiring, or p2p-trading-example specs reintroduce `internal/cli/common/`, `cmd/main.go`, `internal/x402/handler.go`, `internal/companion/discovery.go`, `contracts/MockUSDC.sol`, `scripts/test-p2p-trading.sh`, or `docker-entrypoint-p2p.sh`
- **THEN** an executable repository test SHALL fail

### Requirement: Companion connectivity docs guard stays executable
Repository-level regressions that reintroduce stale automatic companion-discovery claims into public docs or main specs SHALL be enforced by an executable test.

#### Scenario: Stale companion discovery claims are rejected
- **WHEN** companion connectivity docs or specs reintroduce automatic Bonjour/mDNS discovery claims or a legacy dedicated companion-address config key
- **THEN** an executable repository test SHALL fail

### Requirement: P2P identity feature-doc guard stays executable
Repository-level regressions that reintroduce stale `lango p2p identity` wording into public P2P feature docs SHALL be enforced by an executable test.

#### Scenario: Stale identity CLI wording is rejected
- **WHEN** `docs/features/p2p-network.md` claims that `lango p2p identity` does not print the active DID directly
- **THEN** an executable repository test SHALL fail

### Requirement: P2P feature git-command-summary guard stays executable
Repository-level regressions that reintroduce stale `lango p2p git push` or `lango p2p git fetch` summary wording into public P2P feature docs SHALL be enforced by an executable test.

#### Scenario: Stale git command summaries are rejected
- **WHEN** `docs/features/p2p-network.md` describes `lango p2p git push` as directly creating and pushing a bundle or `lango p2p git fetch` as directly fetching and applying one
- **THEN** an executable repository test SHALL fail

### Requirement: README P2P git-summary guard stays executable
Repository-level regressions that reintroduce stale `lango p2p git push` or `lango p2p git fetch` summary wording into the README quick reference SHALL be enforced by an executable test.

#### Scenario: Stale README git summaries are rejected
- **WHEN** `README.md` describes `lango p2p git push` as pushing a workspace git bundle to peers or `lango p2p git fetch` as fetching one directly
- **THEN** an executable repository test SHALL fail

### Requirement: README P2P completeness guard stays executable
Repository-level regressions that drop implemented `p2p workspace`, `p2p team`, or `p2p zkp` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented P2P command families remain listed
- **WHEN** the repository still ships the implemented `workspace`, `team`, and `zkp` P2P CLI families
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

#### Scenario: P2P quick-reference required operands remain listed
- **WHEN** the repository still ships P2P firewall and session commands that require peer identifiers
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits `--peer-did <did>` or `<peer-did>` from those quick-reference entries
- **AND** the test SHALL fail if `docs/features/p2p-network.md` or `docs/features/zkp.md` shows affected P2P commands without required peer operands

### Requirement: Memory quick-reference completeness guard stays executable
Repository-level regressions that drop required memory command operands from public quick references SHALL be enforced by an executable test.

#### Scenario: Memory clear required session key remains listed
- **WHEN** the repository still ships `lango memory clear <session-key>`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` lists the command without `<session-key>`
- **AND** the test SHALL fail if `docs/features/observational-memory.md` shows `lango memory clear` without `<session-key>`

### Requirement: Config quick-reference completeness guard stays executable
Repository-level regressions that drop `config get` output or secret flags from public quick references SHALL be enforced by an executable test.

#### Scenario: Config get full usage remains listed
- **WHEN** the repository still ships `lango config get <dot.path>` with `--output plain|json` and `--show-secrets`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits the full usage string

### Requirement: Config CLI behavior coverage stays executable
Repository-level regressions in config get/set/keys behavior SHALL be enforced by executable tests.

#### Scenario: Dynamic config key templates remain listed
- **WHEN** the repository still supports map-backed config set paths for providers, MCP server env/header values, and auth providers
- **THEN** executable config CLI tests SHALL fail if `collectKeys` or `lango config keys <prefix>` omits the corresponding dynamic templates
- **AND** the tests SHALL fail if dynamic templates include unsupported `time.Duration` leaves such as `mcp.servers.<name>.timeout`

### Requirement: Economy quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `economy escrow list/show/sentinel status` commands from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented economy escrow quick-reference entries remain listed
- **WHEN** the repository still ships `lango economy escrow list`, `show`, and `sentinel status`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries

### Requirement: README smart-account completeness guard stays executable
Repository-level regressions that drop implemented `lango account` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented smart-account quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango account` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README config profile-management completeness guard stays executable
Repository-level regressions that drop implemented `lango config` profile-management command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented config profile commands remain listed
- **WHEN** the repository still ships the implemented `lango config list/create/use/delete/import/export/validate` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README top-level utility completeness guard stays executable
Repository-level regressions that drop implemented `lango version` and `lango health` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented top-level utility quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango version` and `lango health` commands
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: Agent-diagnostics quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango agent` diagnostics command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented agent-diagnostics quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango agent trace list`, `lango agent trace show <trace-id>`, `lango agent graph <session>`, and `lango agent trace metrics` commands
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries

### Requirement: README agent-inspection completeness guard stays executable
Repository-level regressions that drop implemented `lango agent` inspection command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented agent-inspection quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango agent status`, `lango agent list`, `lango agent tools`, and `lango agent hooks` commands
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README graph completeness guard stays executable
Repository-level regressions that drop implemented `lango graph` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented graph quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: Alerts quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango alerts` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented alerts quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango alerts list` and `lango alerts summary` commands
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries

### Requirement: Extension quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango extension` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented extension quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` commands
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries

### Requirement: Extension CLI reference quality guard stays executable
Repository-level regressions that let the dedicated extension CLI reference drift away from the implemented command surface SHALL be enforced by an executable test.

#### Scenario: Implemented extension command contract remains documented
- **WHEN** the repository still ships the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` commands with `table|json|plain` output and `--yes` scripted confirmations
- **THEN** an executable repository test SHALL fail if `docs/cli/extension.md` no longer documents that command and flag surface

### Requirement: CLI index core/status section guard stays executable
Repository-level regressions that remove dedicated core or status sections from the public CLI index SHALL be enforced by an executable test.

#### Scenario: Implemented core and status sections remain listed
- **WHEN** the repository still ships the implemented core command family and the `lango status` dead-letter command family
- **THEN** an executable repository test SHALL fail if `docs/cli/index.md` no longer includes dedicated `Core Commands` and `Status Dashboard` sections for them

### Requirement: Graph CLI reference quality guard stays executable
Repository-level regressions that let the dedicated graph CLI reference drift away from the implemented command surface SHALL be enforced by an executable test.

#### Scenario: Implemented graph command contract remains documented
- **WHEN** the repository still ships the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import <file>` commands
- **THEN** an executable repository test SHALL fail if `docs/cli/graph.md` no longer documents that command surface, the `table|json` output contract, the `export --format json|csv` contract, or the `clear --force` behavior

### Requirement: CLI index graph-section guard stays executable
Repository-level regressions that put graph quick-reference rows back inside the Agent & Memory section instead of keeping a dedicated graph section SHALL be enforced by an executable test.

#### Scenario: CLI index keeps graph coverage in its own section
- **WHEN** the repository still ships a dedicated `docs/cli/graph.md` reference and the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import <file>` commands
- **THEN** an executable repository test SHALL fail if `docs/cli/index.md` drops the dedicated `Graph Store` section, loses the handoff to `docs/cli/graph.md`, or reintroduces graph command rows inside the `Agent & Memory` section

### Requirement: Agent-memory docs scope guard stays executable
Repository-level regressions that reintroduce duplicated graph command docs into the agent-and-memory CLI reference SHALL be enforced by an executable test.

#### Scenario: Agent-memory docs keep graph coverage delegated
- **WHEN** the repository still ships a dedicated `docs/cli/graph.md` reference
- **THEN** an executable repository test SHALL fail if `docs/cli/agent-memory.md` reintroduces standalone `lango graph ...` command sections instead of delegating to the graph reference

### Requirement: Core docs scope guard stays executable
Repository-level regressions that reintroduce duplicated agent diagnostics into the core CLI reference SHALL be enforced by an executable test.

#### Scenario: Core docs keep agent diagnostics delegated
- **WHEN** the repository still ships a dedicated `docs/cli/agent.md` reference
- **THEN** an executable repository test SHALL fail if `docs/cli/core.md` reintroduces standalone `lango agent trace ...` or `lango agent graph ...` sections instead of delegating to the agent reference

### Requirement: Core config docs scope guard stays executable
Repository-level regressions that reintroduce duplicated config command docs into the core CLI reference SHALL be enforced by an executable test.

#### Scenario: Core docs keep config coverage delegated
- **WHEN** the repository still ships a dedicated `docs/cli/config.md` reference
- **THEN** an executable repository test SHALL fail if `docs/cli/core.md` reintroduces standalone `lango config ...` command sections instead of delegating to the config reference

### Requirement: Architecture security CLI truthfulness guard stays executable
Repository-level regressions that let the architecture project-structure docs describe a stale security CLI surface SHALL be enforced by an executable test.

#### Scenario: Project-structure security row remains truthful
- **WHEN** the repository still ships canonical `lango security change-passphrase` and deprecated `migrate-passphrase`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` stops describing that canonical/deprecated distinction
- **AND** it SHALL fail if the row drops current `keyring store/clear/status`, `recovery setup/restore`, or `kms status/test/keys/wrap/detach` coverage

### Requirement: Architecture passphrase package-path guard stays executable
Repository-level regressions that let the architecture project-structure docs reintroduce the deleted top-level `passphrase/` package path SHALL be enforced by an executable test.

#### Scenario: Project-structure passphrase row remains truthful
- **WHEN** the repository still ships passphrase helpers under `internal/security/passphrase`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` reintroduces `passphrase/` instead of `security/passphrase/`

### Requirement: Architecture graph/metrics CLI truthfulness guard stays executable
Repository-level regressions that let the architecture project-structure docs describe stale graph or metrics CLI surfaces SHALL be enforced by an executable test.

#### Scenario: Project-structure graph and metrics rows remain truthful
- **WHEN** the repository still ships `lango graph add/export/import` and `lango metrics policy`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` stops describing those current command surfaces

### Requirement: Architecture config CLI truthfulness guard stays executable
Repository-level regressions that let the architecture project-structure docs omit the shipped `cli/configcmd/` package or its current command surface SHALL be enforced by an executable test.

#### Scenario: Project-structure config row remains truthful
- **WHEN** the repository still ships the `lango config` management surface with `get`, `set`, and `keys`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` omits `cli/configcmd/` or stops describing those current command surfaces

#### Scenario: README config inventory order remains truthful
- **WHEN** the README internal tree still documents the shipped config management surface
- **THEN** an executable repository test SHALL fail if it falls back to stale command ordering such as placing `validate` ahead of `get`, `set`, and `keys`

### Requirement: Shared CLI support inventory guard stays executable
Repository-level regressions that let the public inventory docs omit shared CLI support packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Shared CLI support packages remain visible
- **WHEN** the repository still ships `internal/cli/cliboot`, `internal/cli/clihttp`, and `internal/cli/workbenchstart`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or the README internal tree stops describing those packages and their current responsibilities

### Requirement: README mission-projection package guard stays executable
Repository-level regressions that let the README internal package tree omit shipped mission-projection packages SHALL be enforced by an executable test.

#### Scenario: Mission projection packages remain visible
- **WHEN** the repository still ships `internal/proposal`, `internal/loopview`, and `internal/collabview`
- **THEN** an executable repository test SHALL fail if the README internal tree stops describing those packages and their current responsibilities

### Requirement: Runtime-support package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped runtime-support packages or misdescribe their current responsibilities SHALL be enforced by an executable test.

#### Scenario: Runtime-support package rows remain truthful
- **WHEN** the repository still ships `internal/exportability`, `internal/knowledgeruntime`, `internal/receipts`, `internal/storagebroker`, `internal/streamx`, `internal/tooloutput`, and `internal/toolparam`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits runtime branch selection, receipt progression, storage brokering, stream combinators, tool output retention, or typed tool-parameter extraction

### Requirement: Payment-settlement support package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped payment-settlement support packages or misdescribe their current responsibilities SHALL be enforced by an executable test.

#### Scenario: Payment-settlement package rows remain truthful
- **WHEN** the repository still ships `internal/finance`, `internal/paymentapproval`, `internal/paymentgate`, `internal/settlementprogression`, `internal/settlementexecution`, `internal/partialsettlementexecution`, `internal/escrowexecution`, `internal/disputehold`, `internal/escrowadjudication`, `internal/escrowrelease`, `internal/escrowrefund`, `internal/postadjudicationreplay`, and `internal/postadjudicationstatus`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits approval evaluation, receipt gating, settlement progression, direct or partial settlement execution, escrow fund/hold/adjudication, release/refund execution, or post-adjudication retry/status projection

### Requirement: Execution-retrieval infrastructure package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped execution-retrieval infrastructure packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Execution-retrieval infrastructure rows remain truthful
- **WHEN** the repository still ships `internal/agentrt`, `internal/gatekeeper`, `internal/retrieval`, `internal/search`, `internal/turnrunner`, `internal/turntrace`, `internal/lineio`, and `internal/storeutil`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits runtime coordination, sanitization, retrieval orchestration, FTS5 search substrate, turn execution/tracing, partial-line reading, or store helper responsibilities

### Requirement: Operational-support package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped operational-support packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Operational-support rows remain truthful
- **WHEN** the repository still ships `internal/alerting`, `internal/approvalflow`, `internal/archtest`, and `internal/dbopen`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits alerting thresholds/delivery, artifact release decision mapping, architecture enforcement testing, or managed read-write/read-only database opening responsibilities

### Requirement: Ontology-storage package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped ontology-storage packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Ontology-storage rows remain truthful
- **WHEN** the repository still ships `internal/ontology`, `internal/sqlitedriver`, and `internal/storage`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits ontology governance/tooling, SQLite driver helper behavior, or storage-facade/broker-adapter composition responsibilities

### Requirement: Top-level internal-package parity guard stays executable
Repository-level regressions that let README or architecture inventory docs fall out of parity with the shipped top-level `internal/` package tree SHALL be enforced by an executable test.

#### Scenario: Every top-level internal package remains represented in both inventories
- **WHEN** the repository still ships top-level packages under `internal/`
- **THEN** an executable repository test SHALL fail if any top-level package disappears from `README.md`
- **AND** it SHALL fail if any top-level package disappears from `docs/architecture/project-structure.md`
- **AND** it SHALL therefore catch omissions such as leaving `automation/`, `deadline/`, or `llm/` out of the architecture inventory while they still exist in the codebase

### Requirement: CLI-subpackage parity guard stays executable
Repository-level regressions that let README or architecture inventory docs fall out of parity with the shipped `internal/cli/` subpackage tree SHALL be enforced by an executable test.

#### Scenario: Every CLI subpackage remains represented in both inventories
- **WHEN** the repository still ships subpackages under `internal/cli/`
- **THEN** an executable repository test SHALL fail if any `internal/cli/` subpackage disappears from `README.md`
- **AND** it SHALL fail if any `internal/cli/` subpackage disappears from `docs/architecture/project-structure.md`
- **AND** it SHALL therefore catch omissions affecting both command families and helper packages such as `cliboot`, `clihttp`, `clitypes`, `tuicore`, `workbench`, or `workbenchstart`

### Requirement: CLI index dedicated-reference catalog guard stays executable
Repository-level regressions that let the top-level CLI index drop links to dedicated CLI reference pages SHALL be enforced by an executable test.

#### Scenario: Every dedicated CLI reference remains linked
- **WHEN** the repository still ships dedicated CLI reference pages under `docs/cli/`
- **THEN** an executable repository test SHALL fail if `docs/cli/index.md` stops linking any of those pages
- **AND** it SHALL therefore catch omissions affecting dedicated references such as `core.md`, `status.md`, `agent.md`, `automation.md`, `extension.md`, `graph.md`, `payment.md`, `provenance.md`, `sandbox.md`, or `smartaccount.md`

### Requirement: Architecture index reference-catalog guard stays executable
Repository-level regressions that let the architecture landing page drop links to dedicated architecture references SHALL be enforced by an executable test.

#### Scenario: Every architecture reference remains linked
- **WHEN** the repository still ships dedicated architecture pages under `docs/architecture/`
- **THEN** an executable repository test SHALL fail if `docs/architecture/index.md` stops linking any of those pages
- **AND** it SHALL therefore catch omissions affecting references such as `overview.md`, `project-structure.md`, `data-flow.md`, `knowledge-exchange-runtime.md`, `settlement-progression.md`, `actual-settlement-execution.md`, `retry-dead-letter-handling.md`, or `p2p-knowledge-exchange-track.md`

### Requirement: Docs-home section-catalog guard stays executable
Repository-level regressions that let the docs landing page drop links to top-level documentation sections SHALL be enforced by an executable test.

#### Scenario: Every top-level section index remains linked
- **WHEN** the repository still ships top-level `docs/*/index.md` section pages
- **THEN** an executable repository test SHALL fail if `docs/index.md` stops linking any of those section indexes
- **AND** it SHALL therefore catch omissions affecting top-level sections such as `getting-started`, `architecture`, `cli`, `features`, `security`, `gateway`, `payments`, `automation`, `deployment`, or `development`

### Requirement: Features index reference-catalog guard stays executable
Repository-level regressions that let the features landing page drop links to dedicated feature references SHALL be enforced by an executable test.

#### Scenario: Every feature reference remains linked
- **WHEN** the repository still ships dedicated feature pages under `docs/features/`
- **THEN** an executable repository test SHALL fail if `docs/features/index.md` stops linking any of those pages
- **AND** it SHALL therefore catch omissions affecting references such as `agent-format.md`, `learning.md`, `knowledge.md`, `knowledge-graph.md`, `ontology.md`, `p2p-network.md`, `provenance.md`, `run-ledger.md`, or `zkp.md`

### Requirement: Generic section-index parity guard stays executable
Repository-level regressions that let any public docs section index drift away from the dedicated pages in its own directory SHALL be enforced by an executable test.

#### Scenario: Every section index remains complete
- **WHEN** the repository still ships public section indexes under `docs/*/index.md`
- **THEN** an executable repository test SHALL fail if any such section index stops linking one of its sibling `*.md` pages other than `index.md`
- **AND** it SHALL therefore catch omissions across sections such as `architecture`, `cli`, `features`, `security`, `automation`, `payments`, `gateway`, `deployment`, `development`, or `getting-started`

### Requirement: SQLite driver and db-open regression coverage stays executable
Repository-level regressions that break low-level SQLite driver helpers or managed/read-only database opening flows SHALL be enforced by executable tests close to those packages.

#### Scenario: SQLite helper and DB-open paths remain covered
- **WHEN** the repository still ships `internal/sqlitedriver` and `internal/dbopen`
- **THEN** executable package tests SHALL cover DSN construction, file-header validation including legacy unreadable-header rejection, and connection configuration in `internal/sqlitedriver`
- **AND** executable package tests SHALL cover managed DB creation, read-only reopen of an existing managed DB, missing-file read-only failure, and legacy-header fail-fast in `internal/dbopen`
- **AND** executable package tests SHALL cover concurrent `OpenManaged` invocations so Atlas-backed schema migration crashes regress visibly

### Requirement: Session-store migration safety stays executable
Repository-level regressions that reintroduce concurrent Ent schema-migration crashes in the session store constructor SHALL be enforced by executable package tests.

#### Scenario: Concurrent NewEntStore remains safe
- **WHEN** the repository still ships `internal/session.NewEntStore`
- **THEN** executable package tests SHALL fail if concurrent `NewEntStore` invocations panic or return migration/open errors caused by unsynchronized Ent schema setup

### Requirement: Deprecated session-passphrase behavior stays executable
Repository-level regressions that make the legacy session-store passphrase option behave like an active SQLCipher unlock path again SHALL be prevented by executable package tests.

#### Scenario: Plaintext session store ignores deprecated passphrase option
- **WHEN** the repository still ships `internal/session.NewEntStore` with `WithPassphrase(...)`
- **THEN** executable package tests SHALL fail if a plaintext session store starts failing solely because the deprecated passphrase option was provided

### Requirement: Deprecated SQLCipher open-arg behavior stays executable
Repository-level regressions that make legacy encryption arguments behave as active SQLCipher controls again SHALL be prevented by executable package tests around the DB-open paths.

#### Scenario: Plaintext DB-open paths ignore deprecated encryption args
- **WHEN** the repository still ships managed and read-only DB-open paths without SQLCipher runtime support
- **THEN** executable package tests SHALL fail if plaintext managed or read-only opens start failing solely because deprecated encryption arguments are provided

### Requirement: Shared test schema helper stays cycle-safe and reusable
Repository-level regressions that reintroduce duplicated or import-cycle-prone Ent schema bootstrap logic in tests SHALL be prevented by a shared test-only helper boundary.

#### Scenario: Serialized schema setup remains reusable from tests
- **WHEN** tests in multiple packages need Ent schema bootstrap
- **THEN** they SHALL be able to use a minimal helper under `internal/testutil/schemautil` without importing broader testutil surfaces that create package cycles
- **AND** the helper SHALL continue to serialize Atlas-backed schema creation for those tests

### Requirement: Direct test Schema.Create usage stays blocked
Repository-level regressions that let test code bypass the shared schema helper and call Ent schema creation directly SHALL be prevented by an executable quality guard.

#### Scenario: Tests keep using the shared schema helper
- **WHEN** test code under `internal/` still needs Ent schema bootstrap
- **THEN** an executable repository test SHALL fail if a test file reintroduces direct `Schema.Create` usage outside the approved `internal/testutil/schemautil` helper and the guard test that intentionally scans for that token

### Requirement: README graph inventory truthfulness guard stays executable
Repository-level regressions that let the README internal tree describe a stale graph CLI subset SHALL be enforced by an executable test.

#### Scenario: README graph inventory remains truthful
- **WHEN** the repository still ships `lango graph add/export/import`
- **THEN** an executable repository test SHALL fail if `README.md` stops describing those current command surfaces

### Requirement: Payment/metrics inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe stale payment or metrics CLI surfaces SHALL be enforced by an executable test.

#### Scenario: Payment and metrics inventory remains truthful
- **WHEN** the repository still ships `lango payment x402` and `lango metrics policy`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
- **AND** it SHALL fail if the README internal tree falls back to the stale bracket shorthand for the metrics family

### Requirement: P2P inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe stale P2P CLI subsets SHALL be enforced by an executable test.

#### Scenario: P2P inventory remains truthful
- **WHEN** the repository still ships P2P workspace, git, provenance, team, and ZKP command families
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
- **AND** it SHALL fail if the README internal tree falls back to broad P2P family-only shorthand instead of the current subcommand slices
- **AND** it SHALL fail if the README internal tree falls back to hyphen-compressed shorthand for those P2P subcommand slices

### Requirement: README P2P package-subtree guard stays executable
Repository-level regressions that let the README internal package tree omit shipped `internal/p2p` subpackages SHALL be enforced by an executable test.

#### Scenario: P2P package subtree remains visible
- **WHEN** the repository still ships the current `internal/p2p` subpackages
- **THEN** an executable repository test SHALL fail if the README internal tree drops one of those package rows
- **AND** it SHALL fail if the parent `p2p/` summary falls back to the older narrower wording

### Requirement: Alerts inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs omit the current alerts CLI surface SHALL be enforced by an executable test.

#### Scenario: Alerts inventory remains truthful
- **WHEN** the repository still ships `lango alerts list` and `lango alerts summary`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces

### Requirement: Memory inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe a stale memory CLI subset SHALL be enforced by an executable test.

#### Scenario: Memory inventory remains truthful
- **WHEN** the repository still ships `lango memory agents` and `lango memory agent <name>`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
- **AND** it SHALL fail if the README falls back to inventory wording that omits the `<name>` placeholder

### Requirement: Contract inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe a stale contract CLI subset SHALL be enforced by an executable test.

#### Scenario: Contract inventory remains truthful
- **WHEN** the repository still ships `lango contract read`, `lango contract call`, and `lango contract abi load`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces

### Requirement: Economy inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe stale economy family shorthand instead of the current status-oriented command surface SHALL be enforced by an executable test.

#### Scenario: Economy inventory remains truthful
- **WHEN** the repository still ships `lango economy budget status`, `risk status`, `pricing status`, `negotiate status`, and `escrow status/list/show/sentinel status`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces

### Requirement: README security inventory truthfulness guard stays executable
Repository-level regressions that let the README internal CLI inventory collapse the current security command surface into stale shorthand SHALL be enforced by an executable test.

#### Scenario: README security inventory remains truthful
- **WHEN** the repository still ships canonical `lango security change-passphrase`, deprecated `migrate-passphrase`, `recovery setup/restore`, and `kms wrap/detach`
- **THEN** an executable repository test SHALL fail if `README.md` stops describing those current security command surfaces in its internal tree inventory
- **AND** it SHALL fail if the README falls back to hyphen-compressed inventory wording such as `store-clear-status`, `setup-restore`, or `status-test-keys-wrap-detach`

### Requirement: Remaining CLI inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped chat, extension, provenance, run, sandbox, or status CLI surfaces SHALL be enforced by an executable test.

#### Scenario: Remaining CLI inventory remains truthful
- **WHEN** the repository still ships chat, extension, provenance, run, sandbox, status, and workflow `validate <file>` command surfaces
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
- **AND** it SHALL fail if the README falls back to vague status wording instead of the concrete dead-letter command inventory
- **AND** it SHALL fail if the run inventory drops the `journal <run-id>` placeholder
- **AND** it SHALL fail if the provenance inventory falls back to broad family-only shorthand instead of the current subcommand slices
- **AND** it SHALL fail if the README provenance row falls back to hyphen-compressed shorthand for those subcommand slices

### Requirement: A2A/agent inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped A2A or agent diagnostics surfaces, or keep stale duplicate chat inventory rows, SHALL be enforced by an executable test.

#### Scenario: A2A and agent inventory remains truthful
- **WHEN** the repository still ships A2A card/check and agent trace/graph diagnostics commands
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
- **AND** it SHALL fail if the README internal tree reintroduces the stale duplicate `chat` inventory row

#### Scenario: README agent inventory rejects stale hyphen shorthand
- **WHEN** the README internal tree still documents the shipped agent diagnostics surface
- **THEN** an executable repository test SHALL fail if it falls back to `trace list-show-metrics/graph` instead of the current `trace list/show/metrics/graph` wording

#### Scenario: Duplicate A2A README inventory rows are rejected
- **WHEN** the README internal tree still documents the shipped `lango a2a card/check` surface
- **THEN** an executable repository test SHALL fail if that `a2a/` inventory row appears more than once

### Requirement: Smart-account inventory truthfulness guard stays executable
Repository-level regressions that let smart-account inventory docs describe stale command subsets SHALL be enforced by an executable test.

#### Scenario: Smart-account inventory remains truthful
- **WHEN** the repository still ships `lango account session list/create/revoke`, `module list/install`, `policy show/set`, and `paymaster status/approve`
- **THEN** an executable repository test SHALL fail if `docs/cli/smartaccount.md`, `docs/architecture/project-structure.md`, or `README.md` stops describing those current command surfaces

### Requirement: README approval completeness guard stays executable
Repository-level regressions that drop implemented `lango approval` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented approval quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango approval` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: Security quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango security` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented security quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango security` command family, including canonical `change-passphrase`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries

### Requirement: Security quick-reference truthfulness guard stays executable
Repository-level regressions that blur the distinction between canonical `change-passphrase` and deprecated `migrate-passphrase` in public quick references SHALL be enforced by an executable test.

#### Scenario: Security quick references keep canonical/deprecated wording
- **WHEN** the repository still ships canonical `lango security change-passphrase` and deprecated `lango security migrate-passphrase`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` stops distinguishing the non-reencrypting canonical path from the deprecated legacy migration path

### Requirement: README A2A completeness guard stays executable
Repository-level regressions that drop implemented `lango a2a` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented A2A quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango a2a` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README learning completeness guard stays executable
Repository-level regressions that drop implemented `lango learning` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented learning quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango learning` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README librarian completeness guard stays executable
Repository-level regressions that drop implemented `lango librarian` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented librarian quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango librarian` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README memory completeness guard stays executable
Repository-level regressions that drop implemented `lango memory` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented memory quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango memory` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README contract completeness guard stays executable
Repository-level regressions that drop implemented `lango contract` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented contract quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango contract` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README payment completeness guard stays executable
Repository-level regressions that drop implemented `lango payment` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented payment quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango payment` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: Metrics quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango metrics` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented metrics quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango metrics` command family, including `lango metrics policy`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries

### Requirement: README run completeness guard stays executable
Repository-level regressions that drop implemented `lango run` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented run quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango run` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README MCP completeness guard stays executable
Repository-level regressions that drop implemented `lango mcp` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented MCP quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango mcp` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README provenance completeness guard stays executable
Repository-level regressions that drop implemented `lango provenance` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented provenance quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango provenance` command family
- **THEN** an executable repository test SHALL fail if `README.md` or the main `docs-only` spec omits those quick-reference entries

### Requirement: README background completeness guard stays executable
Repository-level regressions that drop implemented `lango bg` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented background quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango bg` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README sandbox completeness guard stays executable
Repository-level regressions that drop implemented `lango sandbox` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented sandbox quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango sandbox` command family
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README automation completeness guard stays executable
Repository-level regressions that drop implemented `lango cron`, `lango workflow`, or `lango bg` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented automation quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango cron`, `lango workflow`, and `lango bg` command families
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README core P2P completeness guard stays executable
Repository-level regressions that drop implemented core `p2p` command families from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented core P2P families remain listed
- **WHEN** the repository still ships the implemented `status`, `peers`, `connect`, `disconnect`, `firewall`, `discover`, `identity`, `reputation`, `pricing`, `provenance`, `session`, and `sandbox` P2P CLI families
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: P2P feature mixed-command-mode guard stays executable
Repository-level regressions that reintroduce stale all-ephemeral wording into public P2P feature docs SHALL be enforced by an executable test.

#### Scenario: Mixed command-mode intro is preserved
- **WHEN** `docs/features/p2p-network.md` would describe the whole `lango p2p` command list as ephemeral-node execution independent of the running server
- **THEN** an executable repository test SHALL fail

### Requirement: P2P on-chain examples script-pattern guard stays executable
Repository-level regressions that reintroduce stale universal polling-loop claims into the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Fixed-sleep exception is preserved
- **WHEN** `examples/p2p-trading/scripts/test-p2p-trading.sh` still uses `sleep 15` before peer checks
- **THEN** an executable repository test SHALL fail if `openspec/specs/p2p-onchain-examples/spec.md` claims that discovery scripts universally use polling loops instead of fixed sleep

### Requirement: P2P on-chain examples count guard stays executable
Repository-level regressions that reintroduce a stale shipped-example count into the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Current seven-example inventory is preserved
- **WHEN** seven example directories ship under `examples/`
- **THEN** an executable repository test SHALL fail if `openspec/specs/p2p-onchain-examples/spec.md` claims there are only six Docker Compose examples

### Requirement: P2P on-chain examples exact-count guard stays executable
Repository-level regressions that reintroduce stale exact `Tests (N)` claims into the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Stale exact test-count claims are rejected
- **WHEN** the shipped example scripts continue evolving independently
- **THEN** an executable repository test SHALL fail if `openspec/specs/p2p-onchain-examples/spec.md` claims hard-coded `Tests (N)` totals

### Requirement: P2P on-chain examples inventory guard stays executable
Repository-level regressions that omit shipped example summaries from the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Shipped example headings are preserved
- **WHEN** the repository ships the current seven top-level example directories
- **THEN** an executable repository test SHALL fail if `openspec/specs/p2p-onchain-examples/spec.md` omits one of the shipped example headings such as `p2p-trading`

### Requirement: Escrow lifecycle tests cover expiry settlement invariants
Escrow lifecycle tests SHALL cover both explicit and implicit expiry paths so refund and persistence failures cannot drift.

#### Scenario: Implicit expiry refund regression is covered
- **WHEN** `checkExpiry` is reached through a lifecycle method for a funded or active escrow
- **THEN** tests SHALL assert the buyer refund is invoked
- **AND** tests SHALL assert the expired state is persisted only after the refund path succeeds

#### Scenario: Implicit expiry update error is covered
- **WHEN** the store update fails during implicit expiry
- **THEN** tests SHALL assert the error is returned to the caller
- **AND** tests SHALL assert `ErrEscrowExpired` remains matchable on implicit expiry failures
- **AND** tests SHALL assert the previous escrow state is preserved

#### Scenario: Expiry boundary is covered
- **WHEN** the current time equals ExpiresAt
- **THEN** tests SHALL assert implicit and explicit expiry paths treat the escrow as expired

#### Scenario: Early explicit expiry regression is covered
- **WHEN** `Expire` is called before ExpiresAt has been reached
- **THEN** tests SHALL assert the escrow state is preserved and locked funds are not refunded

#### Scenario: Dangling detector expiry gate is covered
- **WHEN** a pending escrow is older than `maxPending` but has not reached ExpiresAt
- **THEN** tests SHALL assert the detector preserves the pending state and publishes no dangling event
- **AND** tests SHALL assert pending escrows older than `maxPending` are expired once ExpiresAt has been reached

### Requirement: Interactive command guard tests
CLI tests SHALL cover interactive-only command guards so TUI commands do not drift into non-TTY bootstrap or Bubble Tea startup.

#### Scenario: Onboard guard is covered
- **WHEN** the onboard command's interactive-terminal guard fails
- **THEN** tests SHALL assert the command returns the guard error
- **AND** tests SHALL assert the onboard run path is not invoked

#### Scenario: Settings guard is covered
- **WHEN** the settings command's interactive-terminal guard fails
- **THEN** tests SHALL assert the command returns the guard error
- **AND** tests SHALL assert the settings run path is not invoked

### Requirement: Structured CLI exit-code tests

Executable tests SHALL verify that command packages can return structured CLI exit-code errors and that the `lango` entrypoint preserves those codes.

#### Scenario: Main preserves structured CLI exit code
- **WHEN** the root Cobra command returns a structured CLI error carrying exit code 3
- **THEN** `runMain()` SHALL return 3
- **AND** stderr SHALL include the underlying error message exactly once

#### Scenario: Extension commands return structured exit errors
- **WHEN** `lango extension install` or `lango extension remove` exits through user-declined or user-error paths
- **THEN** direct command tests SHALL observe a returned structured CLI error with the documented code
- **AND** the tests SHALL NOT intercept `os.Exit` through panic seams

### Requirement: Internal CLI os.Exit hygiene guard

Executable repository tests SHALL reject direct `os.Exit` usage from non-test Go files under `internal/cli/`.

#### Scenario: Internal CLI os.Exit regressions are rejected
- **WHEN** an `internal/cli/**/*.go` production file reintroduces a direct `os.Exit` reference
- **THEN** a repository-level test SHALL fail and identify the offending file and line

### Requirement: ExtendableDeadline parent cancellation coverage

Executable tests SHALL cover parent context cancellation reason classification for the shared deadline package and the app compatibility alias. Tests SHALL also cover preservation of parent deadline metadata and `context.DeadlineExceeded` semantics.

#### Scenario: Shared deadline parent cancellation is covered
- **WHEN** an `internal/deadline` test cancels the parent context before timer expiry
- **THEN** the test SHALL assert the derived context is cancelled and `Reason()` is `"cancelled"`

#### Scenario: App alias parent cancellation is covered
- **WHEN** an `internal/app` test cancels the parent context before timer expiry through `NewExtendableDeadline`
- **THEN** the test SHALL assert the derived context is cancelled and `Reason()` is `"cancelled"`

#### Scenario: Shared deadline parent deadline semantics are covered
- **WHEN** an `internal/deadline` test creates an ExtendableDeadline from a parent context with a deadline
- **THEN** the test SHALL assert the derived context exposes the parent deadline
- **AND** parent deadline expiry leaves the derived context with `context.DeadlineExceeded`
- **AND** `Reason()` is `"cancelled"`

### Requirement: Payment approval panic guard stays executable

Executable tests SHALL prevent production `panic` calls from being reintroduced in `internal/paymentapproval` non-test Go files.

#### Scenario: Payment approval package panic regressions are rejected
- **WHEN** `internal/paymentapproval` non-test Go source files contain a `panic(` call
- **THEN** an executable package test SHALL fail and identify the offending file and line

### Requirement: Smart account module panic guard stays executable

Executable tests SHALL prevent production `panic` calls from being reintroduced in `internal/smartaccount/module` non-test Go files.

#### Scenario: Smart account module panic regressions are rejected
- **WHEN** `internal/smartaccount/module` non-test Go source files contain a `panic(` call
- **THEN** an executable package test SHALL fail and identify the offending file and line

### Requirement: Session migration panic regression stays executable

Executable tests SHALL cover panic recovery in session secret migration.

#### Scenario: Panicking migration callback is covered
- **WHEN** the session package tests run
- **THEN** they SHALL exercise `MigrateSecrets` with a panicking re-encryption callback
- **AND** they SHALL assert the method returns an error instead of panicking

### Requirement: Ontology exchange panic guard stays executable

Executable tests SHALL prevent production `panic` calls from being reintroduced in `internal/ontology/exchange.go`.

#### Scenario: Ontology exchange panic regressions are rejected
- **WHEN** `internal/ontology/exchange.go` contains a `panic(` call
- **THEN** an executable package test SHALL fail and identify the offending file and line

### Requirement: Gateway-backed CLI address resolution coverage stays executable

Repository-level regressions in gateway-backed CLI default address resolution SHALL be enforced by executable tests.

#### Scenario: Metrics CLI configured gateway default remains covered
- **WHEN** `lango metrics` and metrics subcommands are constructed with a config loader
- **THEN** executable tests SHALL fail if they ignore configured `server.host` and `server.port` when `--addr` is omitted
- **AND** executable tests SHALL fail if explicit `--addr` stops overriding the configured gateway

#### Scenario: Alerts CLI configured gateway default remains covered
- **WHEN** `lango alerts list` or `lango alerts summary` is constructed with a config loader
- **THEN** executable tests SHALL fail if they ignore configured `server.host` and `server.port` when `--addr` is omitted
- **AND** executable tests SHALL fail if explicit `--addr` stops overriding the configured gateway

#### Scenario: Status CLI configured gateway probe remains covered
- **WHEN** `lango status` bootstraps a config with custom `server.host` and `server.port`
- **THEN** executable tests SHALL fail if the live `/health` probe still uses a hardcoded localhost/18789 default

### Requirement: Gateway CLI docs default wording guard stays executable

Repository-level docs guards SHALL prevent gateway-backed CLI docs from presenting localhost/18789 as the only default when the command now honors configured server host and port.

#### Scenario: Gateway CLI docs configured-default wording remains covered
- **WHEN** public CLI docs for metrics, alerts, or status are checked
- **THEN** executable tests SHALL fail if the docs omit configured-gateway default wording for `--addr`
