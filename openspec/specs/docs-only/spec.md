## Purpose

Documentation accuracy requirements ensuring README.md stays in sync with codebase configuration and feature state.
## Requirements
### Requirement: README documents librarian configuration

README.md Configuration Reference table SHALL include all `librarian.*` fields matching `LibrarianConfig` in `internal/config/types.go`.

#### Scenario: Librarian config fields present
- **WHEN** a user reads the Configuration Reference in README.md
- **THEN** the table contains entries for `librarian.enabled`, `librarian.observationThreshold`, `librarian.inquiryCooldownTurns`, `librarian.maxPendingInquiries`, `librarian.autoSaveConfidence`, `librarian.provider`, `librarian.model`

### Requirement: README documents automation defaultDeliverTo

README.md Configuration Reference table SHALL include `defaultDeliverTo` fields for cron, background, and workflow sections.

#### Scenario: defaultDeliverTo fields present
- **WHEN** a user reads the Cron Scheduling, Background Execution, and Workflow Engine config sections
- **THEN** each section contains a `*.defaultDeliverTo` entry with type `[]string` and default `[]`

### Requirement: README multi-agent table reflects librarian tools

The multi-agent orchestration table SHALL list proactive knowledge extraction in the librarian role and include `librarian_pending_inquiries` and `librarian_dismiss_inquiry` in the tools column.

#### Scenario: Librarian row updated
- **WHEN** a user reads the Multi-Agent Orchestration table
- **THEN** the librarian row includes "proactive knowledge extraction" in Role and both `librarian_pending_inquiries` and `librarian_dismiss_inquiry` in Tools

### Requirement: README documents streaming in gateway feature

README.md Features list SHALL describe the Gateway as supporting real-time streaming.

#### Scenario: Gateway feature line updated
- **WHEN** a user reads the Features list in README.md
- **THEN** the Gateway bullet reads "WebSocket/HTTP server with real-time streaming"

### Requirement: README documents observational memory context limit configs

README.md Configuration Reference table SHALL include `observationalMemory.maxReflectionsInContext` and `observationalMemory.maxObservationsInContext` fields matching `ObservationalMemoryConfig` in `internal/config/types.go`.

#### Scenario: Context limit config fields present
- **WHEN** a user reads the Observational Memory config section in README.md
- **THEN** the table contains `observationalMemory.maxReflectionsInContext` (int, default `5`) and `observationalMemory.maxObservationsInContext` (int, default `20`)

### Requirement: README documents embedding cache

README.md Embedding & RAG section SHALL include an Embedding Cache subsection describing in-memory caching with 5-minute TTL and 100-entry limit.

#### Scenario: Embedding cache subsection present
- **WHEN** a user reads the Embedding & RAG section in README.md
- **THEN** there is an "Embedding Cache" heading describing automatic in-memory caching with 5-minute TTL and 100-entry limit

### Requirement: README documents observational memory context limits

README.md Observational Memory section SHALL describe context limits for reflections and observations.

#### Scenario: Context limits bullet present
- **WHEN** a user reads the Observational Memory component list in README.md
- **THEN** there is a "Context Limits" bullet describing default limits of 5 reflections and 20 observations

### Requirement: README documents WebSocket events

README.md SHALL include a WebSocket Events subsection documenting `agent.thinking`, `agent.chunk`, and `agent.done` events with their payloads.

#### Scenario: WebSocket events table present
- **WHEN** a user reads the WebSocket section in README.md
- **THEN** there is a "WebSocket Events" heading with a table listing `agent.thinking`, `agent.chunk`, and `agent.done` events

#### Scenario: Backward compatibility noted
- **WHEN** a user reads the WebSocket Events section
- **THEN** there is a note that clients not handling `agent.chunk` will still receive the full response in the RPC result

### Requirement: Documentation accuracy

Documentation, prompts, and CLI help text SHALL accurately reflect all implemented features including P2P REST API endpoints, CLI flags, and example projects.

#### Scenario: P2P REST API documented
- **WHEN** a user reads the HTTP API documentation
- **THEN** the P2P REST endpoints (`/api/p2p/status`, `/api/p2p/peers`, `/api/p2p/identity`) SHALL be documented with request/response examples

#### Scenario: Secrets --value-hex documented
- **WHEN** a user reads the secrets set CLI documentation
- **THEN** the `--value-hex` flag SHALL be documented with non-interactive usage examples

#### Scenario: P2P trading example discoverable
- **WHEN** a user reads the README
- **THEN** the `examples/p2p-trading/` directory SHALL be referenced in an Examples section

### Requirement: Public config CLI examples stay aligned with the real command contract
Public docs and README examples for config import/export/get SHALL use the real flag and argument shapes exposed by the CLI.

#### Scenario: Stale config get --format example is rejected
- **WHEN** a public doc or README reintroduces `lango config get ... --format json`
- **THEN** an executable repository test SHALL fail

#### Scenario: Profile-less config export/import examples are rejected
- **WHEN** a public doc or README reintroduces `lango config export` without the required profile argument or `lango config import` without the explicit `--profile` example
- **THEN** an executable repository test SHALL fail

### Requirement: Public config CLI docs stay complete for implemented read/write commands
Public CLI documentation SHALL continue to expose the implemented `lango config get`, `lango config set`, and `lango config keys` commands.

#### Scenario: Config get/set/keys docs regressions are rejected
- **WHEN** README or public CLI docs drop one of `lango config get <dot.path>`, `lango config set <dot.path> <value>`, or `lango config keys [prefix]`
- **THEN** an executable repository test SHALL fail

### Requirement: CLI quick reference stays complete and well-formed for implemented operator commands
The public CLI quick reference SHALL keep implemented KMS wrap/detach and P2P workspace/provenance commands visible, and prose notes SHALL not split command tables.

#### Scenario: CLI index operator-command regressions are rejected
- **WHEN** `docs/cli/index.md` drops implemented KMS wrap/detach or P2P workspace/provenance command rows
- **THEN** an executable repository test SHALL fail

#### Scenario: CLI index prose does not split the Agent & Memory table
- **WHEN** `docs/cli/index.md` reintroduces explanatory prose between Agent & Memory table rows
- **THEN** an executable repository test SHALL fail

### Requirement: Contract CLI docs stay aligned with the explicit output-format contract
Public contract CLI docs and the main contract interaction spec SHALL keep the current `--output table|json` contract instead of drifting back to a boolean output toggle.

#### Scenario: Stale contract boolean-output docs are rejected
- **WHEN** public contract CLI docs or the main contract interaction spec reintroduce a boolean `--output` flag table entry or a bare `--output` example without an explicit format
- **THEN** an executable repository test SHALL fail

### Requirement: Migrated CLI docs stay aligned with explicit output-format contracts
Public docs and main specs for CLI families already migrated from `--json` toggles to `--output table|json` SHALL not drift back to the old flag shape.

#### Scenario: Stale migrated CLI `--json` docs are rejected
- **WHEN** public docs or main specs for migrated command families reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail

#### Scenario: Stale P2P operator `--json` docs are rejected
- **WHEN** public docs or main specs for the migrated P2P operator family (`status`, `peers`, `identity`, `discover`, `firewall list`, `pricing`, `reputation`, `session list`, `team list`, `team status`, `zkp status`, `zkp circuits`, `workspace create`, `workspace list`, `workspace status`, `git log`) reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail

### Requirement: Approval Pipeline documentation in P2P feature docs
The `docs/features/p2p-network.md` file SHALL include an "Approval Pipeline" section describing the three-stage inbound gate (Firewall ACL → Owner Approval → Tool Execution) with a Mermaid flowchart diagram and auto-approval shortcut rules table.

#### Scenario: Approval Pipeline section present
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** there SHALL be an "Approval Pipeline" section between Knowledge Firewall and Discovery with a Mermaid diagram and descriptions of all three stages

### Requirement: Auto-Approval for Small Amounts in Paid Value Exchange docs
The Paid Value Exchange section in `docs/features/p2p-network.md` SHALL include an "Auto-Approval for Small Amounts" subsection describing the three conditions checked by `IsAutoApprovable`: threshold, maxPerTx, and maxDaily.

#### Scenario: Auto-approval subsection present
- **WHEN** a user reads the Paid Value Exchange section
- **THEN** there SHALL be a subsection documenting the three auto-approval conditions and fallback to interactive approval

### Requirement: P2P skills spec stays truthful about embedded defaults
The `p2p-skills` main spec SHALL not claim embedded default P2P skill files exist when the repository only ships the placeholder embedded skill scaffold.

#### Scenario: Stale embedded P2P skill claims are rejected
- **WHEN** the repository ships only the placeholder `skills/.placeholder/SKILL.md`
- **THEN** the `p2p-skills` main spec SHALL not claim that `skills/p2p-*/SKILL.md` files already exist

### Requirement: CLI test harness spec stays truthful about current helper layout
The `cli-test-harness` main spec SHALL not refer to deleted shared harness files when the current reusable helpers live in `internal/testutil/loaders.go` and `internal/testutil/helpers.go`.

#### Scenario: Stale deleted harness path is rejected
- **WHEN** a maintainer updates the `cli-test-harness` main spec
- **THEN** it SHALL not claim that `internal/testutil/cli_harness.go` is the current shared harness implementation

### Requirement: Economy tool-builder specs stay truthful about current paths
Specs SHALL not claim deleted app-local sentinel or economy tool-builder files are still the current registration source once those builders moved into their owning packages.

#### Scenario: Stale deleted economy tool-builder paths are rejected
- **WHEN** a maintainer updates the sentinel or on-chain escrow specs
- **THEN** they SHALL not claim that `internal/app/tools_sentinel.go` or `tools_economy.go` is the current tool-builder source of truth

### Requirement: Agent memory spec stays truthful about current builder ownership
The `agent-memory` main spec SHALL not claim a deleted app-local tool builder path once agent memory tools are owned by the `agentmemory` package and wired from the current app module.

#### Scenario: Stale deleted agent memory builder path is rejected
- **WHEN** a maintainer updates the `agent-memory` main spec
- **THEN** it SHALL not claim that `internal/app/tools_agentmemory.go` is the current registration path

### Requirement: Package-consolidation spec stays truthful about current package locations
The `package-consolidation` main spec SHALL not keep deleted package-directory paths once the consolidation is complete and the current packages already live elsewhere.

#### Scenario: Deleted consolidated package paths are rejected
- **WHEN** a maintainer updates the `package-consolidation` main spec
- **THEN** it SHALL not claim `internal/ctxutil/`, `internal/passphrase/`, or `internal/zkp/` as current package locations

### Requirement: Main specs avoid known broken single-file path references
Main specs SHALL not keep stale single-file references after code moved or package paths were renamed.

#### Scenario: Known broken single-file references are rejected
- **WHEN** a maintainer updates the affected main specs
- **THEN** they SHALL not reintroduce stale references such as `internal/cli/common/`, `cmd/main.go`, `internal/x402/handler.go`, `internal/companion/discovery.go`, `contracts/MockUSDC.sol`, `scripts/test-p2p-trading.sh`, or `docker-entrypoint-p2p.sh`

### Requirement: Companion connectivity docs stay truthful about the shipped model
Public docs and main specs SHALL not claim that automatic companion discovery ships when the current runtime uses gateway-backed companion connections.

#### Scenario: Stale companion discovery claims are rejected
- **WHEN** a maintainer updates companion connectivity docs or specs
- **THEN** they SHALL not claim automatic Bonjour/mDNS companion discovery or a legacy dedicated companion-address config key as current shipped behavior

### Requirement: Reputation and Pricing endpoints in REST API tables
All REST API documentation (p2p-network.md, http-api.md, README.md, examples/p2p-trading/README.md) SHALL list `GET /api/p2p/reputation` and `GET /api/p2p/pricing` with curl examples and JSON response samples.

#### Scenario: Endpoints in p2p-network.md
- **WHEN** a user reads the REST API table in `docs/features/p2p-network.md`
- **THEN** reputation and pricing endpoints SHALL be listed with curl examples

#### Scenario: Endpoints in http-api.md
- **WHEN** a user reads `docs/gateway/http-api.md`
- **THEN** there SHALL be full endpoint sections for reputation and pricing with query parameters, JSON response examples, and curl commands

### Requirement: Reputation and Pricing CLI commands documented
The CLI command listings in `docs/features/p2p-network.md` and `README.md` SHALL include `lango p2p reputation` and `lango p2p pricing` commands.

#### Scenario: CLI commands in feature docs
- **WHEN** a user reads the CLI Commands section of `docs/features/p2p-network.md`
- **THEN** reputation and pricing commands SHALL be listed

### Requirement: P2P identity feature docs stay truthful about current CLI output
Public P2P feature docs SHALL not describe `lango p2p identity` as omitting the active DID when the CLI already prints it directly when available.

#### Scenario: Stale identity CLI wording is rejected
- **WHEN** a maintainer updates `docs/features/p2p-network.md`
- **THEN** it SHALL not claim that `lango p2p identity` only shows peer/node identity and listen addresses without printing the DID

### Requirement: P2P feature command summaries stay truthful about git runtime guidance
Public P2P feature docs SHALL not summarize `lango p2p git push` or `lango p2p git fetch` as direct live git execution when the current CLI is still a server-backed guidance surface.

#### Scenario: Stale git command summaries are rejected
- **WHEN** a maintainer updates `docs/features/p2p-network.md`
- **THEN** it SHALL not describe `lango p2p git push` as directly creating and pushing a bundle or `lango p2p git fetch` as directly fetching and applying one

### Requirement: README P2P git summaries stay truthful about runtime guidance
The README quick reference SHALL not summarize `lango p2p git push` or `lango p2p git fetch` as direct live bundle execution when the current CLI is still a server-backed guidance surface.

#### Scenario: Stale README git summaries are rejected
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL not describe `lango p2p git push` as pushing a workspace git bundle to peers or `lango p2p git fetch` as fetching one directly

### Requirement: README includes implemented P2P operator commands
The README quick reference SHALL include the implemented `p2p workspace`, `p2p team`, and `p2p zkp` command surfaces that are already present in the public CLI index and feature docs.

#### Scenario: Implemented P2P command families stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango p2p workspace`, `lango p2p team`, and `lango p2p zkp` command entries

### Requirement: Public economy quick references include implemented escrow operator commands
The public quick-reference docs SHALL include the implemented `lango economy escrow list`, `show`, and `sentinel status` commands.

#### Scenario: Implemented escrow commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango economy escrow list`, `lango economy escrow show`, and `lango economy escrow sentinel status` entries

### Requirement: README includes implemented smart-account commands
The README quick reference SHALL include the implemented `lango account` command family that is already present in the public CLI index and dedicated smart-account docs.

#### Scenario: Implemented smart-account commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango account info/deploy/session/module/policy/paymaster` command entries

### Requirement: README includes implemented config profile-management commands
The README quick reference SHALL include the implemented `lango config` profile-management commands that are already present in the public CLI index and dedicated config docs.

#### Scenario: Implemented config profile commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango config list/create/use/delete/import/export/validate` command entries

### Requirement: README includes implemented top-level utility commands
The README quick reference SHALL include the implemented `lango version` and `lango health` commands that are already present in the public CLI index and top-level utility docs.

#### Scenario: Implemented top-level utility commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango version` and `lango health` command entries

### Requirement: Public quick references include implemented agent-diagnostics commands
The public quick-reference docs SHALL include the implemented `lango agent` diagnostics commands that are already present in the CLI index and core command docs.

#### Scenario: Implemented agent-diagnostics commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango agent trace list`, `lango agent trace show <trace-id>`, `lango agent graph <session>`, and `lango agent trace metrics` command entries

### Requirement: README includes implemented agent-inspection commands
The README quick reference SHALL include the implemented `lango agent` inspection commands that are already present in the public CLI index.

#### Scenario: Implemented agent-inspection commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango agent status`, `lango agent list`, `lango agent tools`, and `lango agent hooks` command entries

### Requirement: README includes implemented graph commands
The README quick reference SHALL include the implemented `lango graph` command family that is already present in the public CLI index.

#### Scenario: Implemented graph commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import` command entries

### Requirement: Public quick references include implemented alerts commands
The public quick-reference docs SHALL include the implemented `lango alerts` command family that is already present in dedicated alerts docs and the wired root CLI.

#### Scenario: Implemented alerts commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango alerts list` and `lango alerts summary` command entries

### Requirement: Public quick references include implemented extension commands
The public quick-reference docs SHALL include the implemented `lango extension` command family that is already present in README extension-pack docs and the wired root CLI.

#### Scenario: Implemented extension commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` command entries

### Requirement: Extension CLI reference stays aligned with the current command surface
The dedicated extension CLI reference SHALL describe the implemented `inspect`, `install`, `list`, and `remove` commands and their current output/confirmation contracts.

#### Scenario: Implemented extension command contract stays documented
- **WHEN** a maintainer updates `docs/cli/extension.md`
- **THEN** it SHALL document the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` commands, the `table|json|plain` output contract, and the `--yes` scripted-run path

### Requirement: CLI index includes dedicated core and status command sections
The public CLI index SHALL include dedicated sections for the implemented core and status command families so the index structure matches the existing `docs/cli/core.md` and `docs/cli/status.md` references.

#### Scenario: Implemented core and status sections stay discoverable
- **WHEN** a maintainer updates `docs/cli/index.md`
- **THEN** it SHALL include dedicated `Core Commands` and `Status Dashboard` sections covering the implemented `lango`/`cockpit`/`chat`/`serve`/`version`/`health`/`onboard`/`settings`/`doctor` entries and the `lango status` dead-letter command family

### Requirement: Graph CLI reference stays aligned with the current command surface
The dedicated graph CLI reference SHALL describe the implemented `status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands and their current output/format contracts.

#### Scenario: Implemented graph command contract stays documented
- **WHEN** a maintainer updates `docs/cli/graph.md`
- **THEN** it SHALL document the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import <file>` commands, the `table|json` output contract, the `export --format json|csv` contract, and the `clear --force` confirmation bypass

### Requirement: CLI index gives graph commands a dedicated section
The public CLI index SHALL keep implemented `lango graph` commands in a dedicated graph section once the repository ships a dedicated graph CLI reference.

#### Scenario: Graph commands stay separated from Agent & Memory
- **WHEN** a maintainer updates `docs/cli/index.md`
- **THEN** it SHALL keep a dedicated `Graph Store` section for the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands
- **AND** it SHALL hand off detailed graph coverage to `docs/cli/graph.md` instead of leaving those command rows embedded inside the `Agent & Memory` section

### Requirement: Agent & memory CLI reference delegates graph commands to the dedicated graph page
The agent-and-memory CLI reference SHALL not keep a duplicated embedded graph command manual once a dedicated graph CLI reference exists.

#### Scenario: Graph command duplication stays removed from agent-memory docs
- **WHEN** a maintainer updates `docs/cli/agent-memory.md`
- **THEN** it SHALL hand off graph command coverage to `docs/cli/graph.md` instead of embedding standalone `lango graph ...` sections

### Requirement: Core CLI reference delegates agent diagnostics to the dedicated agent page
The core CLI reference SHALL not keep a duplicated embedded agent-diagnostics manual once a dedicated agent CLI reference exists.

#### Scenario: Agent diagnostics duplication stays removed from core docs
- **WHEN** a maintainer updates `docs/cli/core.md`
- **THEN** it SHALL hand off `lango agent trace ...` and `lango agent graph ...` coverage to `docs/cli/agent.md` instead of embedding standalone diagnostics sections

### Requirement: Core CLI reference delegates config commands to the dedicated config page
The core CLI reference SHALL not keep a duplicated embedded config command manual once a dedicated config CLI reference exists.

#### Scenario: Config command duplication stays removed from core docs
- **WHEN** a maintainer updates `docs/cli/core.md`
- **THEN** it SHALL hand off `lango config ...` coverage to `docs/cli/config.md` instead of embedding standalone config subcommand sections

### Requirement: Architecture project-structure docs stay aligned with the current security CLI surface
The public architecture project-structure reference SHALL describe `cli/security/` using the current canonical `change-passphrase` command and mark `migrate-passphrase` as deprecated.

#### Scenario: Project-structure security row stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** the `cli/security/` row SHALL include `change-passphrase`
- **AND** it SHALL describe `migrate-passphrase` as deprecated legacy surface rather than as the primary passphrase-rotation command
- **AND** it SHALL continue to mention `keyring store/clear/status`, `recovery setup/restore`, and `kms status/test/keys/wrap/detach`

### Requirement: Architecture project-structure docs stay aligned with the current passphrase package path
The public architecture project-structure reference SHALL not keep the deleted top-level `passphrase/` package path once passphrase helpers live under the security subtree.

#### Scenario: Project-structure passphrase row stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** it SHALL describe `security/passphrase/`
- **AND** it SHALL NOT reintroduce the deleted `passphrase/` package path

### Requirement: Architecture project-structure docs stay aligned with the current graph and metrics CLI surface
The public architecture project-structure reference SHALL list the currently implemented `cli/graph/` and `cli/metrics/` command families rather than outdated subsets.

#### Scenario: Project-structure graph and metrics rows stay truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** the `cli/graph/` row SHALL include `add`, `export`, and `import`
- **AND** the `cli/metrics/` row SHALL include `policy`

### Requirement: Architecture project-structure docs include the current config CLI surface
The public architecture project-structure reference SHALL include the shipped `cli/configcmd/` package and its current configuration-management command surface.

#### Scenario: Project-structure config row stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** it SHALL include a `cli/configcmd/` row
- **AND** that row SHALL describe `lango config list`, `create`, `use`, `delete`, `import`, `export`, `get`, `set`, `keys`, and `validate`

#### Scenario: README config inventory keeps the current command order
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe `lango config list/create/use/delete/import/export/get/set/keys/validate`

### Requirement: Architecture and README inventory docs include shared CLI support packages
The public inventory docs SHALL include the shipped shared CLI support packages that back gateway-oriented commands.

#### Scenario: Shared CLI support packages stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `cli/cliboot/`, `cli/clihttp/`, and `cli/workbenchstart/`
- **AND** the README internal tree SHALL include `cliboot/`, `clihttp/`, and `workbenchstart/`
- **AND** those rows SHALL describe bootstrap loader callbacks, shared HTTP/JSON helper responsibilities, and context-aware workbench starter/recovery prompt builders truthfully

### Requirement: README internal package tree includes current mission-projection packages
The README internal package tree SHALL include the shipped durable mission projection packages instead of omitting them.

#### Scenario: Mission projection packages stay visible
- **WHEN** a maintainer updates the README internal package tree
- **THEN** it SHALL include `proposal/`, `loopview/`, and `collabview/`
- **AND** those rows SHALL describe transient proposal flow, deterministic operator-loop projection, and deterministic mission-collaboration projection truthfully

### Requirement: Architecture and README inventory docs include current runtime-support packages
The public inventory docs SHALL include shipped runtime-support packages that back exportability, receipt progression, storage brokering, stream composition, and dynamic tool plumbing instead of omitting them from the package inventory.

#### Scenario: Runtime-support package rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `exportability/`, `knowledgeruntime/`, `receipts/`, `storagebroker/`, `streamx/`, `tooloutput/`, and `toolparam/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe source-class exportability evaluation, knowledge-exchange runtime branch selection, canonical receipt/event progression, persistent stdio JSON storage brokering, iterator-based stream combinators, TTL-backed tool output retention, and typed tool parameter extraction truthfully

### Requirement: Architecture and README inventory docs include current payment-settlement support packages
The public inventory docs SHALL include the shipped payment and settlement support packages that implement approval, direct payment gating, settlement progression, escrow execution, dispute adjudication, and post-adjudication retry/status flows instead of omitting them from the package inventory.

#### Scenario: Payment-settlement package rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `finance/`, `paymentapproval/`, `paymentgate/`, `settlementprogression/`, `settlementexecution/`, `partialsettlementexecution/`, `escrowexecution/`, `disputehold/`, `escrowadjudication/`, `escrowrelease/`, `escrowrefund/`, `postadjudicationreplay/`, and `postadjudicationstatus/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe USDC monetary helpers, upfront-payment approval evaluation, direct-payment receipt gating, settlement progression mapping, direct and partial settlement execution, escrow create/fund flow, dispute hold and adjudication, escrow release/refund execution, and post-adjudication retry/status projection truthfully

### Requirement: Architecture and README inventory docs include current execution-retrieval infrastructure packages
The public inventory docs SHALL include the shipped execution and retrieval infrastructure packages that implement runtime coordination, response sanitization, retrieval orchestration, search substrate, turn execution/tracing, and store/line helpers instead of omitting them from the package inventory.

#### Scenario: Execution-retrieval infrastructure rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `agentrt/`, `gatekeeper/`, `retrieval/`, `search/`, `turnrunner/`, `turntrace/`, `lineio/`, and `storeutil/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe runtime coordination, response sanitization, fact/temporal retrieval orchestration, domain-agnostic FTS5 search, turn execution/tracing, partial-line reading, and store copy/JSON helpers truthfully

### Requirement: Architecture and README inventory docs include current operational-support packages
The public inventory docs SHALL include the shipped operational-support packages that implement alerting, canonical artifact approval flow, architecture enforcement tests, and managed database opening instead of omitting them from the package inventory.

#### Scenario: Operational-support rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `alerting/`, `approvalflow/`, `archtest/`, and `dbopen/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe threshold-based alerting, artifact release decision mapping, architecture boundary/bootstrap enforcement tests, and managed read-write/read-only database opening truthfully
- **AND** the `dbopen/` row SHALL mention serialized schema migration rather than implying fully parallel-safe Ent migration

### Requirement: Architecture and README inventory docs include current ontology-storage packages
The public inventory docs SHALL include the shipped ontology and storage foundation packages that implement ontology governance, shared SQLite opening, and storage-facade composition instead of omitting them from the package inventory.

#### Scenario: Ontology-storage rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `ontology/`, `sqlitedriver/`, and `storage/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe ontology governance/tooling, shared SQLite driver helpers, and broker-aware storage-facade composition truthfully

### Requirement: README and architecture inventory stay in top-level internal-package parity
The public README internal tree and architecture project-structure inventory SHALL both mention every shipped top-level `internal/` package so that new package additions do not silently appear in only one document.

#### Scenario: Every top-level internal package appears in both inventories
- **WHEN** the repository still ships top-level packages under `internal/`
- **THEN** `README.md` SHALL include every top-level package row
- **AND** `docs/architecture/project-structure.md` SHALL include every top-level package row
- **AND** the architecture inventory SHALL continue to include `automation/`, `deadline/`, and `llm/` alongside the other shipped top-level packages

### Requirement: README and architecture inventory stay in CLI-subpackage parity
The public README internal tree and architecture project-structure inventory SHALL both mention every shipped `internal/cli/` subpackage so that new CLI subpackages, including shared helper packages, do not silently appear in only one document.

#### Scenario: Every CLI subpackage appears in both inventories
- **WHEN** the repository still ships subpackages under `internal/cli/`
- **THEN** `README.md` SHALL include every `internal/cli/` subpackage row
- **AND** `docs/architecture/project-structure.md` SHALL include every `internal/cli/` subpackage row
- **AND** that parity SHALL cover both command families and helper packages such as `cliboot/`, `clihttp/`, `clitypes/`, `tuicore/`, `workbench/`, and `workbenchstart/`

### Requirement: CLI index links every dedicated CLI reference page
The public CLI index SHALL provide an explicit catalog of every dedicated page under `docs/cli/` so operators can discover the deeper command-family references from the top-level index.

#### Scenario: Dedicated CLI references stay linked from the index
- **WHEN** a maintainer updates `docs/cli/index.md`
- **THEN** it SHALL include links to every dedicated page under `docs/cli/` other than `index.md`
- **AND** that catalog SHALL cover command-family pages such as `core.md`, `status.md`, `agent.md`, `agent-memory.md`, `automation.md`, `extension.md`, `graph.md`, `payment.md`, `provenance.md`, `sandbox.md`, and `smartaccount.md`

### Requirement: Architecture index links every architecture reference page
The public architecture index SHALL provide an explicit catalog of every dedicated page under `docs/architecture/` so deep-dive architecture references remain discoverable from the top-level architecture landing page.

#### Scenario: Architecture references stay linked from the index
- **WHEN** a maintainer updates `docs/architecture/index.md`
- **THEN** it SHALL include links to every dedicated page under `docs/architecture/` other than `index.md`
- **AND** that catalog SHALL cover pages such as `overview.md`, `project-structure.md`, `data-flow.md`, `knowledge-exchange-runtime.md`, `settlement-progression.md`, `actual-settlement-execution.md`, `retry-dead-letter-handling.md`, and `p2p-knowledge-exchange-track.md`

### Requirement: Docs home links every top-level section index
The public docs home page SHALL provide a section catalog that links every top-level docs section carrying its own `index.md`, so major documentation areas remain discoverable from the landing page.

#### Scenario: Top-level section indexes stay linked from docs home
- **WHEN** a maintainer updates `docs/index.md`
- **THEN** it SHALL include links to every top-level `docs/*/index.md` section index
- **AND** that catalog SHALL cover sections such as `getting-started/`, `architecture/`, `cli/`, `features/`, `security/`, `gateway/`, `payments/`, `automation/`, `deployment/`, and `development/`

### Requirement: Features index links every feature reference page
The public features landing page SHALL provide a catalog of every dedicated page under `docs/features/` so feature deep dives remain discoverable from the features section itself.

#### Scenario: Feature references stay linked from the features index
- **WHEN** a maintainer updates `docs/features/index.md`
- **THEN** it SHALL include links to every dedicated page under `docs/features/` other than `index.md`
- **AND** that catalog SHALL cover pages such as `agent-format.md`, `learning.md`, `knowledge.md`, `knowledge-graph.md`, `ontology.md`, `p2p-network.md`, `provenance.md`, `run-ledger.md`, and `zkp.md`

### Requirement: Every docs section index links its own dedicated pages
Each public docs section that ships a local `index.md` SHALL also link every dedicated Markdown page in that same section directory, so section-level navigation remains complete as the docs tree grows.

#### Scenario: Section indexes stay complete
- **WHEN** a maintainer updates a section landing page such as `docs/architecture/index.md`, `docs/cli/index.md`, `docs/features/index.md`, `docs/security/index.md`, `docs/automation/index.md`, `docs/payments/index.md`, `docs/gateway/index.md`, `docs/deployment/index.md`, `docs/development/index.md`, or `docs/getting-started/index.md`
- **THEN** that section index SHALL include links to every sibling `*.md` page in the same directory other than `index.md`
- **AND** this completeness rule SHALL apply generically across all public section indexes under `docs/`

### Requirement: README internal tree stays aligned with the current graph CLI surface
The README internal tree inventory SHALL include the currently implemented graph command family rather than an outdated subset.

#### Scenario: README graph inventory stays truthful
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** the `graph/` row SHALL include `add`, `export`, and `import`

### Requirement: Architecture and README inventory docs stay aligned with the current payment and metrics CLI surface
The public architecture inventory docs SHALL include the currently implemented `payment x402` and `metrics policy` surfaces rather than outdated subsets.

#### Scenario: Payment and metrics inventory rows stay truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the payment inventory SHALL include `x402`
- **AND** the metrics inventory SHALL include `policy`
- **AND** the README internal tree SHALL describe `lango metrics/sessions/tools/agents/policy/history`

### Requirement: Architecture and README inventory docs stay aligned with the current P2P CLI surface
The public architecture inventory docs SHALL include the currently implemented P2P workspace, git, provenance, team, and ZKP surfaces rather than outdated subsets.

#### Scenario: P2P inventory rows stay truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the P2P inventory SHALL include workspace, git, provenance, team, and ZKP command families
- **AND** the README internal tree SHALL describe the current firewall, session, sandbox, workspace, git, provenance, team, and ZKP subcommand slices instead of only broad family names
- **AND** the README internal tree SHALL use slash-separated subcommand slices instead of hyphen-compressed shorthand for those P2P families

### Requirement: README internal tree includes the current P2P package subtree
The README internal package tree SHALL include the currently shipped `internal/p2p` subpackages rather than a partial subset.

#### Scenario: P2P package subtree stays truthful
- **WHEN** a maintainer updates the README internal package tree
- **THEN** it SHALL include `agentpool`, `discovery`, `firewall`, `gitbundle`, `handshake`, `identity`, `ontologybridge`, `paygate`, `protocol`, `provenanceproto`, `reputation`, `settlement`, `team`, `trustpolicy`, `workspace`, and `zkp`
- **AND** the parent `p2p/` summary SHALL mention collaborative workspaces, git/provenance exchange, trust policy, payments, and ZK proofs

### Requirement: Architecture and README inventory docs stay aligned with the current alerts CLI surface
The public architecture inventory docs SHALL include the currently implemented alerts command surface rather than omitting it.

#### Scenario: Alerts inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the alerts inventory SHALL include `list` and `summary`

### Requirement: Architecture and README inventory docs stay aligned with the current memory CLI surface
The public architecture inventory docs SHALL include the currently implemented observational and per-agent memory command surface rather than outdated subsets.

#### Scenario: Memory inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the memory inventory SHALL include `agents` and `agent <name>`
- **AND** the README internal tree SHALL describe `lango memory list/status/clear/agents/agent <name>`

### Requirement: Architecture and README inventory docs stay aligned with the current contract CLI surface
The public architecture inventory docs SHALL include the currently implemented contract command surface rather than truncated stale shorthand.

#### Scenario: Contract inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `abi load`
- **AND** the README internal tree SHALL describe `lango contract read/call/abi load`

### Requirement: Architecture and README inventory docs stay aligned with the current economy CLI surface
The public architecture inventory docs SHALL describe the current economy command surface using the implemented `... status` paths instead of stale family-only shorthand.

#### Scenario: Economy inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `budget status`, `risk status`, `pricing status`, `negotiate status`, and `escrow status/list/show/sentinel status`
- **AND** the README internal tree SHALL describe `lango economy budget status/risk status/pricing status/negotiate status/escrow status/list/show/sentinel status`

### Requirement: README internal tree stays aligned with the current security CLI surface
The README internal CLI inventory SHALL describe the current security command families instead of collapsing canonical, deprecated, recovery, and KMS surfaces into stale shorthand.

#### Scenario: README security inventory stays truthful
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe canonical `change-passphrase` and deprecated `migrate-passphrase`
- **AND** it SHALL continue to mention `secrets`, `keyring store/clear/status`, `recovery setup/restore`, `kms status/test/keys/wrap/detach`, and legacy `db-*` tombstones
- **AND** it SHALL use slash-separated inventory wording instead of stale hyphen-compressed shorthand for those subfamilies

### Requirement: Architecture and README inventory docs stay aligned with the remaining shipped CLI surfaces
The public architecture inventory docs SHALL include the currently implemented chat, extension, provenance, run, sandbox, and status CLI families rather than omitting them.

#### Scenario: Remaining CLI inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** those inventories SHALL include chat, extension, provenance, run, sandbox, status, and workflow `validate <file>` surfaces
- **AND** the README internal tree SHALL describe the status family as `lango status/dead-letter-summary/dead-letters/dead-letter/dead-letter retry`
- **AND** the run inventory SHALL keep the `journal <run-id>` placeholder
- **AND** the provenance inventory SHALL keep the current checkpoint/session/attribution/bundle subcommand slices
- **AND** the README internal tree SHALL use slash-separated provenance subcommand slices instead of hyphen-compressed shorthand

### Requirement: Architecture and README inventory docs stay aligned with the current A2A and agent CLI surface
The public architecture inventory docs SHALL include the currently implemented A2A and agent diagnostics surfaces rather than omitting or abbreviating them to stale subsets.

#### Scenario: A2A and agent inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the inventories SHALL include A2A card/check
- **AND** they SHALL include the agent trace list/show/metrics and graph diagnostics surface
- **AND** the README internal tree SHALL not keep a stale duplicate `chat` row

#### Scenario: README agent inventory uses the current trace slash form
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe the agent diagnostics slice as `trace list/show/metrics/graph` instead of stale hyphen-compressed shorthand

#### Scenario: README internal tree keeps a single A2A row
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL contain exactly one `a2a/` row describing `lango a2a card/check`

### Requirement: Smart-account inventory docs stay aligned with the current command surface
The public smart-account inventory docs SHALL include the currently implemented session, module, policy, and paymaster subcommands rather than abbreviated subsets.

#### Scenario: Smart-account inventory stays truthful
- **WHEN** a maintainer updates `docs/cli/smartaccount.md`, `docs/architecture/project-structure.md`, or the README internal tree inventory
- **THEN** those docs SHALL include `session list/create/revoke`, `module list/install`, `policy show/set`, and `paymaster status/approve`

### Requirement: README includes implemented approval commands
The README quick reference SHALL include the implemented `lango approval` command family that is already present in the public CLI index and dedicated approval docs.

#### Scenario: Implemented approval commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango approval status` command entry

### Requirement: Public quick references include implemented security commands
The public quick-reference docs SHALL include the implemented `lango security` command family that is already present in dedicated security docs.

#### Scenario: Implemented security commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango security` status, `change-passphrase`, deprecated `migrate-passphrase`, secrets, keyring, recovery, legacy db, and kms command entries

### Requirement: Security quick references distinguish canonical and deprecated passphrase rotation paths
Public quick-reference docs SHALL describe `lango security change-passphrase` as the canonical passphrase-rotation path and `lango security migrate-passphrase` as the deprecated legacy full re-encryption path.

#### Scenario: Passphrase rotation quick-reference wording stays truthful
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL describe `lango security change-passphrase` as a non-reencrypting passphrase change
- **AND** they SHALL mark `lango security migrate-passphrase` as deprecated legacy migration

### Requirement: README includes implemented A2A commands
The README quick reference SHALL include the implemented `lango a2a` command family that is already present in the public CLI index and dedicated A2A docs.

#### Scenario: Implemented A2A commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango a2a card` and `lango a2a check <url>` command entries

### Requirement: README includes implemented learning commands
The README quick reference SHALL include the implemented `lango learning` command family that is already present in the public CLI index and dedicated learning docs.

#### Scenario: Implemented learning commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango learning status` and `lango learning history` command entries

### Requirement: README includes implemented librarian commands
The README quick reference SHALL include the implemented `lango librarian` command family that is already present in the public CLI index and dedicated librarian docs.

#### Scenario: Implemented librarian commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango librarian status` and `lango librarian inquiries` command entries

### Requirement: README includes implemented memory commands
The README quick reference SHALL include the implemented `lango memory` command family that is already present in the public CLI index and dedicated memory docs.

#### Scenario: Implemented memory commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango memory list/status/clear/agents/agent <name>` command entries

### Requirement: README includes implemented contract commands
The README quick reference SHALL include the implemented `lango contract` command family that is already present in the public CLI index and dedicated contract docs.

#### Scenario: Implemented contract commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango contract read/call/abi load` command entries

### Requirement: README includes implemented payment commands
The README quick reference SHALL include the implemented `lango payment` command family that is already present in the public CLI index and dedicated payment docs.

#### Scenario: Implemented payment commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango payment balance/history/limits/info/send/x402` command entries

### Requirement: Public quick references include implemented metrics commands
The public quick-reference docs SHALL include the implemented `lango metrics` command family that is already present in dedicated metrics docs.

#### Scenario: Implemented metrics commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango metrics`, `sessions`, `tools`, `agents`, `policy`, and `history` command entries

### Requirement: README includes implemented run commands
The README quick reference SHALL include the implemented `lango run` command family that is already present in the public CLI index and dedicated RunLedger docs.

#### Scenario: Implemented run commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango run list`, `lango run status`, and `lango run journal <run-id>` command entries

### Requirement: README includes implemented MCP commands
The README quick reference SHALL include the implemented `lango mcp` command family that is already present in the public CLI index and dedicated MCP docs.

#### Scenario: Implemented MCP commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango mcp list/add/remove/get/test/enable/disable` command entries

### Requirement: README includes implemented provenance commands
The README quick reference SHALL include the implemented `lango provenance` command family that is already present in the public CLI index and dedicated provenance docs.

#### Scenario: Implemented provenance commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango provenance status`
- **AND** it SHALL include `lango provenance checkpoint list --run <id>`, `lango provenance checkpoint create <label> --run <id>`, and `lango provenance checkpoint show <id>`
- **AND** it SHALL include `lango provenance session tree <session-key>` and `lango provenance session list`
- **AND** it SHALL include `lango provenance attribution show <session-key>` and `lango provenance attribution report <session-key>`
- **AND** it SHALL include `lango provenance bundle export <session-key>` and `lango provenance bundle import <file>`

### Requirement: README includes implemented background-task commands
The README quick reference SHALL include the implemented `lango bg` command family that is already present in the public CLI index and dedicated automation docs.

#### Scenario: Implemented background commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango bg list/status/cancel/result` command entries

### Requirement: README includes implemented sandbox commands
The README quick reference SHALL include the implemented `lango sandbox` command family that is already present in the public CLI index and dedicated sandbox docs.

#### Scenario: Implemented sandbox commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango sandbox status` and `lango sandbox test` command entries

### Requirement: README includes implemented automation commands
The README quick reference SHALL include the implemented `lango cron`, `lango workflow`, and `lango bg` command families that are already present in the public CLI index and dedicated automation docs.

#### Scenario: Implemented automation commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango cron`, `lango workflow`, and `lango bg` command entries

### Requirement: README includes implemented core P2P operator commands
The README quick reference SHALL include the implemented core `p2p` read/control families that are already present in the public CLI index and feature docs.

#### Scenario: Core P2P command families stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango p2p status`, `peers`, `connect`, `disconnect`, `firewall`, `discover`, `identity`, `lango p2p reputation --peer-did <did>`, `pricing`, `provenance`, `session`, and `sandbox` command entries

### Requirement: P2P feature CLI intro stays truthful about mixed command modes
Public P2P feature docs SHALL not describe the entire `lango p2p` surface as ephemeral-node execution when some command families are still server-backed guidance surfaces.

#### Scenario: Mixed command-mode intro is preserved
- **WHEN** a maintainer updates `docs/features/p2p-network.md`
- **THEN** it SHALL not claim that all listed `lango p2p` commands create ephemeral nodes independent of the running server

### Requirement: P2P on-chain examples spec stays truthful about discovery-script patterns
The `p2p-onchain-examples` main spec SHALL not overstate a universal polling-loop pattern when some shipped examples still use fixed warm-up sleeps.

#### Scenario: Fixed-sleep exception remains documented
- **WHEN** a maintainer updates `openspec/specs/p2p-onchain-examples/spec.md`
- **THEN** it SHALL not claim that all example discovery scripts use polling loops instead of fixed sleeps

### Requirement: P2P on-chain examples spec stays truthful about shipped example count
The `p2p-onchain-examples` main spec SHALL reflect the current number of shipped Docker Compose examples.

#### Scenario: Current example count is reflected
- **WHEN** a maintainer updates `openspec/specs/p2p-onchain-examples/spec.md`
- **THEN** it SHALL not describe the current inventory as six examples when seven example directories ship in `examples/`

### Requirement: P2P on-chain examples spec avoids stale exact test counts
The `p2p-onchain-examples` main spec SHALL not present evolving example-script coverage as stale exact `Tests (N)` counts.

#### Scenario: Exact test-count claims are rejected
- **WHEN** a maintainer updates `openspec/specs/p2p-onchain-examples/spec.md`
- **THEN** it SHALL describe representative checks instead of hard-coded stale `Tests (N)` totals

### Requirement: P2P on-chain examples spec lists every shipped example
The `p2p-onchain-examples` main spec SHALL include a summary entry for each shipped top-level example directory.

#### Scenario: Missing shipped example summaries are rejected
- **WHEN** a maintainer updates `openspec/specs/p2p-onchain-examples/spec.md`
- **THEN** it SHALL not omit the shipped `p2p-trading` example from the numbered example summaries

### Requirement: README P2P config fields complete
The README.md P2P configuration reference table SHALL include `p2p.autoApproveKnownPeers`, `p2p.minTrustScore`, `p2p.pricing.enabled`, and `p2p.pricing.perQuery` fields.

#### Scenario: Missing config fields added
- **WHEN** a user reads the P2P Network section of the Configuration Reference in README.md
- **THEN** all four fields SHALL be present with correct types, defaults, and descriptions

### Requirement: Tool usage prompts reflect approval behavior
The `prompts/TOOL_USAGE.md` file SHALL describe auto-approval behavior for `p2p_pay`, the remote owner's approval pipeline for `p2p_query`, and inbound tool invocation gates.

#### Scenario: p2p_pay auto-approval documented
- **WHEN** a user reads the `p2p_pay` description
- **THEN** it SHALL mention that payments below `autoApproveBelow` are auto-approved

#### Scenario: Inbound invocation gates documented
- **WHEN** a user reads the P2P Networking Tool section
- **THEN** there SHALL be a description of the three-stage inbound gate

### Requirement: USDC docs cross-reference P2P auto-approval
The `docs/payments/usdc.md` file SHALL include a P2P integration note explaining that `autoApproveBelow` applies to both outbound payments and inbound paid tool approval.

#### Scenario: P2P integration note present
- **WHEN** a user reads `docs/payments/usdc.md`
- **THEN** there SHALL be a note after the config table linking to the P2P approval pipeline

### Requirement: P2P trading example documents configuration highlights
The `examples/p2p-trading/README.md` SHALL include a "Configuration Highlights" section with a table of key approval and payment settings used in the example.

#### Scenario: Configuration highlights section present
- **WHEN** a user reads the example README
- **THEN** there SHALL be a Configuration Highlights section with autoApproveBelow, autoApproveKnownPeers, pricing settings, and a production warning

### Requirement: test-p2p Makefile target
The root `Makefile` SHALL include a `test-p2p` target that runs `go test -v -race ./internal/p2p/... ./internal/wallet/...` and SHALL be listed in the `.PHONY` declaration.

#### Scenario: test-p2p target runs successfully
- **WHEN** a user runs `make test-p2p`
- **THEN** P2P and wallet tests SHALL execute with race detector enabled

### Requirement: Quickstart references config presets
The getting started quickstart documentation SHALL reference the `--preset` flag and link to the config presets documentation.

#### Scenario: Preset flag in quickstart
- **WHEN** a user reads `docs/getting-started/quickstart.md`
- **THEN** the `--preset` flag SHALL be mentioned with a brief preset table and link to `config-presets.md`

### Requirement: CLI index includes status command
The CLI index quick reference table SHALL include the `lango status` command.

#### Scenario: Status in CLI index
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** `lango status` SHALL appear in the Quick Reference table under Getting Started

### Requirement: Quickstart installation anchor resolves
The getting started quickstart documentation SHALL link to the existing installation anchor instead of a missing fragment.

#### Scenario: Installation anchor is valid
- **WHEN** a user reads `docs/getting-started/quickstart.md`
- **THEN** the installation link SHALL target the existing installation section and its compiler setup anchor

### Requirement: Cockpit public-entry consolidation
After the hidden cockpit guides move out of `docs/`, the public cockpit documentation SHALL keep `docs/features/cockpit.md` as the single public entry for operator-facing material from the cockpit approval, channels, tasks, and troubleshooting guides.

#### Scenario: Approval guidance is on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find approval operations guidance previously split into the approval sub-guide

#### Scenario: Channel, task, and troubleshooting guidance are on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find channel operations, background task operations, and troubleshooting guidance previously split into the other cockpit sub-guides

### Requirement: Identity trust reputation audit documents the landed Reputation V2 contract
The `docs/architecture/identity-trust-reputation-audit.md` page SHALL describe the landed Reputation V2 contract, including separated `earnedTrustScore`, `durableNegativeUnits`, and `temporarySafetySignals`, plus the canonical trust-entry states `bootstrap`, `established`, `review`, and `temporarily_unsafe`.

#### Scenario: Audit page reflects the V2 contract
- **WHEN** a user reads `docs/architecture/identity-trust-reputation-audit.md`
- **THEN** they SHALL find the composite and earned trust distinction documented
- **AND** they SHALL find durable negative units separated from temporary safety signals
- **AND** they SHALL find the four canonical trust-entry states documented

### Requirement: P2P feature docs describe runtime trust-entry consumption
The `docs/features/p2p-network.md` and `docs/features/economy.md` pages SHALL describe how runtime consumers use the landed trust-entry contract.

#### Scenario: P2P network page describes admission and approval states
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** they SHALL find `bootstrap`, `established`, `review`, and `temporarily_unsafe` documented as the canonical trust-entry states
- **AND** they SHALL find `autoApproveKnownPeers` described as limited to returning peers in `established` state
- **AND** they SHALL find post-pay routing described as using earned trust for returning peers

#### Scenario: Economy page describes score consumption
- **WHEN** a user reads `docs/features/economy.md`
- **THEN** they SHALL find bootstrap peers described as using the bootstrap effective score
- **AND** they SHALL find returning peers described as using earned trust for risk and pricing inputs

### Requirement: P2P knowledge exchange track reflects landed reputation runtime integration
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the identity/trust/reputation detailed audit and the first `reputation v2 + runtime integration` slice as landed work, and it SHALL narrow the follow-on work accordingly.

#### Scenario: Track follow-on list is updated
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the required follow-on plan SHALL state that the identity/trust/reputation detailed audit is now landed
- **AND** they SHALL find the first `reputation v2 + runtime integration` slice described as landed work
- **AND** the remaining work SHALL be narrowed to owner-root-aware policy adoption, broader dispute-to-reputation feeds, and richer operator-facing trust/review surfaces

### Requirement: P2P knowledge exchange track reflects the landed pricing negotiation settlement audit
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the pricing/negotiation/settlement detailed audit as landed work and list the follow-on work as `runtime integration`, `settlement execution`, and `escrow lifecycle completion`.

#### Scenario: Track follow-on list is updated
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the required follow-on plan SHALL state that the pricing/negotiation/settlement detailed audit is now landed
- **AND** the follow-on work SHALL include `runtime integration`, `settlement execution`, and `escrow lifecycle completion`

### Requirement: Knowledge exchange runtime architecture page describes the first control-plane slice
The `docs/architecture/knowledge-exchange-runtime.md` page SHALL describe the first transaction-oriented runtime control-plane design slice for `knowledge exchange v1`, centered on transaction receipt and submission receipt, and SHALL list the current limits of that slice.

#### Scenario: Runtime page shows the bounded slice
- **WHEN** a user reads `docs/architecture/knowledge-exchange-runtime.md`
- **THEN** they SHALL find sections covering the runtime design slice, canonical state, current limits, and follow-on work

### Requirement: P2P knowledge exchange track links the runtime design slice
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL reference `knowledge-exchange-runtime.md` as the first transaction-oriented runtime design slice and SHALL state that the remaining work is runtime implementation and broader progression handling.

#### Scenario: Track page points to the runtime slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find the runtime design slice referenced by name and linked to `knowledge-exchange-runtime.md`
- **AND** the follow-on work SHALL be described as implementation, not redesign of the landed slice

### Requirement: Settlement progression architecture page describes the current progression slice
The `docs/architecture/settlement-progression.md` page SHALL describe the current transaction-level settlement progression slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Settlement progression page shows the bounded slice
- **WHEN** a user reads `docs/architecture/settlement-progression.md`
- **THEN** they SHALL find sections describing the current progression slice, what ships, canonical state, and current limits
- **AND** they SHALL find `dispute-ready` described as a public canonical path for renewed disagreement
- **AND** they SHALL find re-escalation from `partially-settled` described as preserving the canonical `partial_settlement_hint`
- **AND** they SHALL find `apply_settlement_progression` described as returning `dispute_lifecycle_status`

### Requirement: P2P knowledge exchange track reflects landed settlement progression
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the current settlement progression slice as landed work and list the remaining work as repeated partial execution, broader multi-round settlement orchestration, and operator/policy surfaces.

#### Scenario: Track page points to the landed settlement progression slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find settlement progression described as a landed first slice
- **AND** they SHALL find explicit `dispute-ready` re-entry described as landed work
- **AND** the remaining work SHALL be described as repeated partial execution, broader multi-round settlement orchestration, and operator/policy surfaces

### Requirement: Actual settlement execution page describes the first direct execution slice
The `docs/architecture/actual-settlement-execution.md` page SHALL describe the first direct settlement execution slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Actual settlement execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/actual-settlement-execution.md`
- **THEN** they SHALL find sections describing the current execution slice, what ships, canonical gate, and current limits

### Requirement: P2P knowledge exchange track reflects landed actual settlement execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the actual settlement execution first slice as landed work and list the remaining work as escrow lifecycle completion and dispute engine completion.

#### Scenario: Track page points to the landed actual settlement execution slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find actual settlement execution described as a landed first slice
- **AND** the remaining work SHALL be described as escrow lifecycle completion and dispute engine completion

### Requirement: Partial settlement execution page describes the first direct partial slice
The `docs/architecture/partial-settlement-execution.md` page SHALL describe the first direct partial settlement execution slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Partial settlement execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/partial-settlement-execution.md`
- **THEN** they SHALL find sections describing the current partial slice, canonical hint model, success/failure semantics, and current limits

### Requirement: P2P knowledge exchange track reflects landed partial settlement execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the partial settlement execution first slice as landed work and list the remaining work as escrow lifecycle completion and dispute engine completion.

#### Scenario: Track page points to the landed partial settlement execution slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find partial settlement execution described as a landed first slice
- **AND** the remaining work SHALL be described as escrow lifecycle completion and dispute engine completion

### Requirement: Escrow release page describes the first funded release slice
The `docs/architecture/escrow-release.md` page SHALL describe the first escrow release slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Escrow release page shows the bounded slice
- **WHEN** a user reads `docs/architecture/escrow-release.md`
- **THEN** they SHALL find sections describing the current escrow release slice, what currently ships, and current limits
- **AND** they SHALL find matching `escrow_adjudication = release` described as part of the gate
- **AND** they SHALL find opposite-branch refund evidence described as blocking execution
- **AND** they SHALL find dispute lifecycle cleanup on successful release described

### Requirement: P2P knowledge exchange track reflects landed escrow release
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the escrow release first slice as landed work and list the remaining work as milestone-aware release, broader execution policy defaults, and richer operator policy surfaces.

#### Scenario: Track page points to the landed escrow release slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find escrow release described as a landed first slice
- **AND** they SHALL find adjudication-aware release gating described as landed work
- **AND** the remaining work SHALL be described as milestone-aware release, broader execution policy defaults, and richer operator policy surfaces

### Requirement: Escrow refund page describes the first funded refund slice
The `docs/architecture/escrow-refund.md` page SHALL describe the first escrow refund slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Escrow refund page shows the bounded slice
- **WHEN** a user reads `docs/architecture/escrow-refund.md`
- **THEN** they SHALL find sections describing the current escrow refund slice, what currently ships, and current limits
- **AND** they SHALL find matching `escrow_adjudication = refund` described as part of the gate
- **AND** they SHALL find opposite-branch release evidence described as blocking execution
- **AND** they SHALL find dispute lifecycle cleanup on successful refund described
- **AND** they SHALL find concurrent refund attempts for the same transaction described as serialized inside the service boundary

### Requirement: P2P knowledge exchange track reflects landed escrow refund
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the escrow refund first slice as landed work and list the remaining work as refund terminal-state design, release-after-refund safety rules, and richer operator policy surfaces.

#### Scenario: Track page points to the landed escrow refund slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find escrow refund described as a landed first slice
- **AND** they SHALL find adjudication-aware refund gating described as landed work
- **AND** the remaining work SHALL be described as refund terminal-state design, release-after-refund safety rules, and richer operator policy surfaces

### Requirement: Dispute hold page describes the first funded dispute hold slice
The `docs/architecture/dispute-hold.md` page SHALL describe the first dispute hold slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Dispute hold page shows the bounded slice
- **WHEN** a user reads `docs/architecture/dispute-hold.md`
- **THEN** they SHALL find sections describing the current dispute hold slice, what currently ships, and current limits
- **AND** they SHALL find canonical `dispute_lifecycle_status = hold-active` described as a hold success outcome
- **AND** they SHALL find `hold_escrow_for_dispute` described as returning `dispute_lifecycle_status`
- **AND** they SHALL find concurrent hold attempts for the same transaction described as serialized inside the service boundary

### Requirement: P2P knowledge exchange track reflects landed dispute hold
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the dispute hold first slice as landed work and list the remaining work as richer arbitration policy, a separate held-escrow lifecycle only if needed, and operator UI.

#### Scenario: Track page points to the landed dispute hold slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dispute hold described as a landed first slice
- **AND** they SHALL find canonical `hold-active` lifecycle state described as landed work
- **AND** the remaining work SHALL be described as richer arbitration policy, a separate held-escrow lifecycle only if needed, and operator UI

### Requirement: Release-vs-refund adjudication page describes the first post-hold branching slice
The `docs/architecture/release-vs-refund-adjudication.md` page SHALL describe the first post-hold release-vs-refund adjudication slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Adjudication page shows the bounded slice
- **WHEN** a user reads `docs/architecture/release-vs-refund-adjudication.md`
- **THEN** they SHALL find sections describing the current adjudication slice, what currently ships, and current limits
- **AND** they SHALL find atomic settlement progression updates described for `release` and `refund`
- **AND** they SHALL find the active dispute lifecycle marker described as preserved through canonical adjudication
- **AND** they SHALL find `adjudicate_escrow_dispute` described as returning `dispute_lifecycle_status`
- **AND** they SHALL find concurrent adjudication attempts for the same transaction described as serialized inside the service boundary

### Requirement: P2P knowledge exchange track reflects landed release-vs-refund adjudication
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the release-vs-refund adjudication first slice as landed work and list the remaining work as config-backed non-manual defaults, richer arbitration policy, and operator UI.

#### Scenario: Track page points to the landed adjudication slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find release-vs-refund adjudication described as a landed first slice
- **AND** they SHALL find manual-recovery-by-default canonical adjudication described as landed work
- **AND** the remaining work SHALL be described as config-backed non-manual defaults, richer arbitration policy, and operator UI

### Requirement: Adjudication-aware release/refund execution gating page describes the first executor-contract slice
The `docs/architecture/adjudication-aware-release-refund-execution-gating.md` page SHALL describe the first slice that connects canonical escrow adjudication to release/refund execution gating, including what currently ships and the current limits of the slice.

#### Scenario: Adjudication-aware execution gating page shows the bounded slice
- **WHEN** a user reads `docs/architecture/adjudication-aware-release-refund-execution-gating.md`
- **THEN** they SHALL find sections describing the current execution-gating slice, what currently ships, and current limits
- **AND** they SHALL find dispute lifecycle cleanup on successful release or refund described

### Requirement: P2P knowledge exchange track reflects landed adjudication-aware release/refund execution gating
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the adjudication-aware release/refund execution gating first slice as landed work and list the remaining work as milestone-aware branch execution, broader dispute automation, and operator/policy surfaces.

#### Scenario: Track page points to the landed adjudication-aware gating slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find adjudication-aware release/refund execution gating described as a landed first slice
- **AND** they SHALL find terminal lifecycle cleanup described as landed work
- **AND** the remaining work SHALL be described as milestone-aware branch execution, broader dispute automation, and operator/policy surfaces

### Requirement: Automatic post-adjudication execution page describes the first inline orchestration slice
The `docs/architecture/automatic-post-adjudication-execution.md` page SHALL describe the first inline convenience slice after escrow adjudication, including what currently ships and the current limits of the slice.

#### Scenario: Automatic post-adjudication execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/automatic-post-adjudication-execution.md`
- **THEN** they SHALL find sections describing the current auto-execution slice, what currently ships, and current limits
- **AND** they SHALL find that omitted execution flags default to `manual_recovery`
- **AND** they SHALL find that `auto_execute` and `background_execute` are mutually exclusive

### Requirement: P2P knowledge exchange track reflects landed automatic post-adjudication execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the automatic post-adjudication execution first slice as landed work, including the shared execution-mode default of `manual_recovery`, and list the remaining work as config-backed non-manual defaults, policy editing for execution-mode selection, and broader dispute engine integration.

#### Scenario: Track page points to the landed auto-execution slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find automatic post-adjudication execution described as a landed first slice
- **AND** the remaining work SHALL be described as background execution, retry orchestration, automatic execution as policy default, and broader dispute engine integration

### Requirement: Background post-adjudication execution page describes the first async dispatch slice
The `docs/architecture/background-post-adjudication-execution.md` page SHALL describe the first background post-adjudication execution slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Background post-adjudication execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/background-post-adjudication-execution.md`
- **THEN** they SHALL find sections describing the current background dispatch slice, what currently ships, and current limits
- **AND** they SHALL find the shared `manual_recovery` / `inline` / `background` execution-mode policy described
- **AND** they SHALL find that background execution remains an explicit opt-in when execution flags are present

### Requirement: P2P knowledge exchange track reflects landed background post-adjudication execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the background post-adjudication execution first slice as landed work and list the remaining work as config-backed non-manual defaults, operator-editable execution-mode policy, broader background-task adoption outside post-adjudication follow-up, and broader dispute engine integration.

#### Scenario: Track page points to the landed background slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find background post-adjudication execution described as a landed first slice
- **AND** the remaining work SHALL be described as retry orchestration, dead-letter handling, dedicated status observation, and policy-driven defaults

### Requirement: Retry / dead-letter handling page describes the first bounded retry slice
The `docs/architecture/retry-dead-letter-handling.md` page SHALL describe the first retry / dead-letter slice for background post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Retry / dead-letter handling page shows the bounded slice
- **WHEN** a user reads `docs/architecture/retry-dead-letter-handling.md`
- **THEN** they SHALL find sections describing the current retry/dead-letter slice, what currently ships, and current limits
- **AND** they SHALL find the normalized retry policy shape described with retry-attempt and base-delay fields
- **AND** they SHALL find the shared `post_adjudication_retry` evidence source described for retry and dead-letter events
- **AND** they SHALL find canonical re-escalation on exhausted retries described as preserving adjudication while setting `settlement_progression_status = dispute-ready`
- **AND** they SHALL find `dispute_lifecycle_status = re-escalated` described for exhausted retries
- **AND** they SHALL find canonical retry-key dedup across pending, running, and scheduled tasks described
- **AND** they SHALL find background-runner panics described as explicit task failures rather than orphaned running tasks
- **AND** they SHALL find receipt-evidence write failures described as operational errors even when the retry hook remains best-effort

### Requirement: P2P knowledge exchange track reflects landed retry / dead-letter handling
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the retry / dead-letter handling first slice as landed work and list the remaining work as operator-editable retry tuning, wider non-post-adjudication adoption of the retry policy shape, and a more generic recovery substrate for arbitrary background task families.

#### Scenario: Track page points to the landed retry slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find retry / dead-letter handling described as a landed first slice
- **AND** they SHALL find canonical re-escalation after exhausted retries described as landed work
- **AND** they SHALL find canonical retry-key dedup across pending, running, and scheduled tasks described as landed work
- **AND** the remaining work SHALL be described as operator replay, generic async retry policy, dead-letter browsing, and policy-driven backoff tuning

### Requirement: Operator replay / manual retry page describes the first replay slice
The `docs/architecture/operator-replay-manual-retry.md` page SHALL describe the first operator replay / manual retry slice for dead-lettered post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Operator replay / manual retry page shows the bounded slice
- **WHEN** a user reads `docs/architecture/operator-replay-manual-retry.md`
- **THEN** they SHALL find sections describing the current replay slice, what currently ships, and current limits
- **AND** they SHALL find replay described as part of the same recovery evidence family as automatic retry and dead-letter handling
- **AND** they SHALL find `manual-retry-requested` evidence described

### Requirement: P2P knowledge exchange track reflects landed operator replay / manual retry
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the operator replay / manual retry first slice as landed work and list the remaining work as inline replay, arbitrary background-task replay, per-transaction recovery snapshots, and broader dispute engine integration.

#### Scenario: Track page points to the landed replay slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find operator replay / manual retry described as a landed first slice
- **AND** the remaining work SHALL be described as dead-letter browsing UI, policy-driven replay controls, generic replay substrate design, and broader dispute engine integration

### Requirement: Policy-driven replay controls page describes the first replay authorization slice
The `docs/architecture/policy-driven-replay-controls.md` page SHALL describe the first policy-driven replay controls slice for post-adjudication replay, including what currently ships and the current limits of the slice.

#### Scenario: Policy-driven replay controls page shows the bounded slice
- **WHEN** a user reads `docs/architecture/policy-driven-replay-controls.md`
- **THEN** they SHALL find sections describing the current replay-authorization slice, what currently ships, and current limits
- **AND** they SHALL find replay authorization described as sitting on top of the shared recovery evidence gate

### Requirement: P2P knowledge exchange track reflects landed policy-driven replay controls
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the policy-driven replay controls first slice as landed work and list the remaining work as richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls.

#### Scenario: Track page points to the landed replay-policy slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find policy-driven replay controls described as a landed first slice
- **AND** the remaining work SHALL be described as richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing / status observation page shows the bounded slice
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find sections describing the current read-only visibility slice, what currently ships, and current limits

#### Scenario: Dead-letter browsing page describes filtering and detail hints
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find filtering and pagination described for the backlog list
- **AND** they SHALL find actor/time-based list filters described
- **AND** they SHALL find dead-letter reason and dispatch-reference filters described
- **AND** they SHALL find subtype/count filters and alternate sort modes described
- **AND** they SHALL find total retry-count and subtype-family filters described
- **AND** they SHALL find any-match family grouping described
- **AND** they SHALL find dominant family described
- **AND** they SHALL find transaction-global retry count and family grouping described
- **AND** they SHALL find transaction-global dominant family described
- **AND** they SHALL find compact per-submission breakdown described
- **AND** they SHALL find the optional detail-view raw background-task bridge described
- **AND** they SHALL find the cockpit dead-letter master-detail read surface described
- **AND** they SHALL find the thin cockpit filter bar described
- **AND** they SHALL find cockpit subtype filtering described
- **AND** they SHALL find cockpit actor/time filtering described
- **AND** they SHALL find cockpit latest-family filtering described
- **AND** they SHALL find cockpit any-match-family filtering described
- **AND** they SHALL find cockpit reason/dispatch filtering described
- **AND** they SHALL find cockpit reset/clear shortcut behavior described
- **AND** they SHALL find cockpit selection preservation described
- **AND** they SHALL find dead-letter CLI surface described
- **AND** they SHALL find dead-letter CLI subtype/latest-family filtering described
- **AND** they SHALL find dead-letter CLI retry action described
- **AND** they SHALL find dead-letter CLI actor/time filtering described
- **AND** they SHALL find dead-letter CLI reason/dispatch filtering described
- **AND** they SHALL find dead-letter CLI `offset` / `limit` pagination described
- **AND** they SHALL find dead-letter CLI summary described
- **AND** they SHALL find dead-letter CLI `by_reason_family` summary buckets described
- **AND** they SHALL find dead-letter CLI `by_actor_family` summary buckets described
- **AND** they SHALL find dead-letter CLI `by_dispatch_family` summary buckets described
- **AND** they SHALL find the CLI `By reason family` table section described
- **AND** they SHALL find the CLI `By actor family` table section described
- **AND** they SHALL find the CLI `By dispatch family` table section described
- **AND** they SHALL find the initial reason-family taxonomy described as `retry-exhausted`, `policy-blocked`, `receipt-invalid`, `background-failed`, and `unknown`
- **AND** they SHALL find the initial actor-family taxonomy described as `operator`, `system`, `service`, and `unknown`
- **AND** they SHALL find the dispatch-family classifier described as using common prefixes plus deterministic first-token fallback
- **AND** they SHALL find dead-letter CLI retry described as supporting an explicit `--actor` override
- **AND** they SHALL find machine-mode dead-letter CLI failures described as structured JSON error payloads when `--output json` is selected
- **AND** they SHALL find top latest dead-letter reasons described for the summary CLI surface
- **AND** they SHALL find raw top latest dead-letter reasons described as still available alongside grouped reason-family summaries
- **AND** they SHALL find top latest manual replay actors described for the summary CLI surface
- **AND** they SHALL find raw top latest manual replay actors described as still available alongside grouped actor-family summaries
- **AND** they SHALL find top latest dispatch references described for the summary CLI surface
- **AND** they SHALL find configurable top-N summary controls described
- **AND** they SHALL find recent dead-letter trend / time-window summary behavior described
- **AND** they SHALL find the cockpit page-top summary strip described
- **AND** they SHALL find the cockpit `reason families:` summary strip line described
- **AND** they SHALL find the cockpit `actor families:` summary strip line described
- **AND** they SHALL find the cockpit `dispatch families:` summary strip line described
- **AND** they SHALL find top latest dead-letter reasons described for the cockpit summary strip
- **AND** they SHALL find raw top latest dead-letter reasons described as still available in the cockpit summary strip
- **AND** they SHALL find top latest manual replay actors described for the cockpit summary strip
- **AND** they SHALL find raw top latest manual replay actors described as still available in the cockpit summary strip
- **AND** they SHALL find top latest dispatch references described for the cockpit summary strip
- **AND** they SHALL find the cockpit trend line described
- **AND** they SHALL find the cockpit detail-pane `Retry` action described
- **AND** they SHALL find inline confirm and success-refresh recovery UX described
- **AND** they SHALL find retry running/failure feedback described
- **AND** they SHALL find CLI retry precheck behavior described
- **AND** they SHALL find CLI retry success described as request acceptance rather than completed execution
- **AND** they SHALL find CLI and cockpit retry failure wording described as distinct from precheck rejection
- **AND** they SHALL find that CLI and cockpit retry inject a local default operator principal when the runtime context is otherwise empty
- **AND** they SHALL find CLI `--any-match-family` filtering described
- **AND** they SHALL find CLI retry follow-up polling and structured follow-up output described
- **AND** they SHALL find cockpit retry follow-up interpretation described
- **AND** they SHALL find that cockpit dead-letter filter fields are forwarded through the shell adapter into the dead-letter list tool
- **AND** they SHALL find dispatch-family grouping described as shared between CLI and cockpit
- **AND** they SHALL find detail navigation hints described for per-transaction status

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with transaction-global dominant family, compact per-submission breakdown, a thin raw background-task bridge on the detail view, a cockpit dead-letter read surface, a page-top cockpit summary strip with raw top latest dead-letter reasons, grouped `reason families:`, top latest manual replay actors, grouped `actor families:`, grouped `dispatch families:`, top latest dispatch references, and a compact trend line, a thin cockpit filter bar, cockpit subtype filtering, cockpit latest-family filtering, cockpit any-match-family filtering, cockpit actor/time filtering, cockpit reason/dispatch filtering, cockpit reset/clear shortcuts, cockpit selection preservation, a cockpit `Retry` action, confirm/refresh recovery UX, refined retry loading/failure/success messaging, and a dead-letter CLI surface including the summary command with grouped `by_reason_family`, `by_actor_family`, and `by_dispatch_family` buckets, `By reason family`, `By actor family`, and `By dispatch family` table sections, raw configurable top-N latest reasons, raw configurable top-N latest manual replay actors, raw configurable top-N latest dispatch references, recent trend / time-window summaries, subtype/latest-family/any-match-family filtering, actor/time filtering, reason/dispatch filtering, and retry follow-up polling with precheck/request-accepted/request-failed semantics, and list the remaining work as configurable taxonomy redesign, broader dead-letter history and generic background-task browsing, wider non-post-adjudication adoption of the retry/recovery substrate, and operator-editable execution/recovery policy surfaces.

#### Scenario: Track page points to the landed status slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter browsing / status observation described as a landed first slice
- **AND** they SHALL find compact per-submission breakdown described as landed work
- **AND** they SHALL find the thin detail-view raw background-task bridge described as landed work
- **AND** they SHALL find the cockpit dead-letter read surface described as landed work
- **AND** they SHALL find the thin cockpit filter bar described as landed work
- **AND** they SHALL find cockpit subtype filtering described as landed work
- **AND** they SHALL find cockpit actor/time filtering described as landed work
- **AND** they SHALL find cockpit latest-family filtering described as landed work
- **AND** they SHALL find cockpit any-match-family filtering described as landed work
- **AND** they SHALL find cockpit reason/dispatch filtering described as landed work
- **AND** they SHALL find cockpit reset/clear shortcuts described as landed work
- **AND** they SHALL find cockpit selection preservation described as landed work
- **AND** they SHALL find dead-letter CLI surface described as landed work
- **AND** they SHALL find dead-letter CLI subtype/latest-family filtering described as landed work
- **AND** they SHALL find dead-letter CLI retry action described as landed work
- **AND** they SHALL find dead-letter CLI actor/time filtering described as landed work
- **AND** they SHALL find dead-letter CLI reason/dispatch filtering described as landed work
- **AND** they SHALL find dead-letter CLI summary described as landed work
- **AND** they SHALL find dead-letter CLI `by_reason_family` described as landed work
- **AND** they SHALL find dead-letter CLI `by_actor_family` described as landed work
- **AND** they SHALL find dead-letter CLI `by_dispatch_family` described as landed work
- **AND** they SHALL find the CLI `By reason family` table section described as landed work
- **AND** they SHALL find the CLI `By actor family` table section described as landed work
- **AND** they SHALL find the CLI `By dispatch family` table section described as landed work
- **AND** they SHALL find the initial reason-family taxonomy described as `retry-exhausted`, `policy-blocked`, `receipt-invalid`, `background-failed`, and `unknown`
- **AND** they SHALL find the initial actor-family taxonomy described as `operator`, `system`, `service`, and `unknown`
- **AND** they SHALL find dispatch-family grouping described as landed work
- **AND** they SHALL find top latest dead-letter reasons described as landed work
- **AND** they SHALL find raw top latest dead-letter reasons described as still available alongside grouped reason-family summaries
- **AND** they SHALL find top latest manual replay actors described as landed work
- **AND** they SHALL find raw top latest manual replay actors described as still available alongside grouped actor-family summaries
- **AND** they SHALL find top latest dispatch references described as landed work
- **AND** they SHALL find configurable top-N summary controls described as landed work
- **AND** they SHALL find recent trend / time-window summaries described as landed work
- **AND** they SHALL find the cockpit page-top summary strip described as landed work
- **AND** they SHALL find the cockpit `reason families:` summary strip line described as landed work
- **AND** they SHALL find the cockpit `actor families:` summary strip line described as landed work
- **AND** they SHALL find the cockpit `dispatch families:` summary strip line described as landed work
- **AND** they SHALL find top latest dead-letter reasons described for the cockpit summary strip as landed work
- **AND** they SHALL find raw top latest dead-letter reasons described as still available in the cockpit summary strip
- **AND** they SHALL find top latest manual replay actors described for the cockpit summary strip as landed work
- **AND** they SHALL find raw top latest manual replay actors described as still available in the cockpit summary strip
- **AND** they SHALL find top latest dispatch references described for the cockpit summary strip as landed work
- **AND** they SHALL find the cockpit trend line described as landed work
- **AND** they SHALL find the cockpit `Retry` action described as landed work
- **AND** they SHALL find confirm/refresh recovery UX described as landed work
- **AND** they SHALL find retry loading/failure feedback described as landed work
- **AND** they SHALL find refined retry success/failure wording described as landed work
- **AND** they SHALL find CLI retry precheck/request-accepted/request-failed semantics described as landed work
- **AND** they SHALL find CLI any-match-family filtering described as landed work
- **AND** they SHALL find CLI retry follow-up polling described as landed work
- **AND** the remaining work SHALL be described as configurable taxonomy redesign, broader dead-letter history and generic background-task browsing, wider non-post-adjudication adoption of the retry/recovery substrate, and operator-editable execution/recovery policy surfaces

### Requirement: Public CLI prompt examples match current punctuation
Public CLI examples SHALL mirror the current prompt punctuation emitted by the commands they document.

#### Scenario: Shared confirmation prompt punctuation is guarded by tests
- **WHEN** a public doc or README reintroduces a stale shared confirmation example such as `[y/N] y`
- **THEN** the repository test suite SHALL fail

### Requirement: Main specs use real purpose summaries
Main specs SHALL replace archive-generated placeholder `Purpose` text with concise summaries that describe the actual scope of the spec.

#### Scenario: Archived placeholder purpose text is guarded by tests
- **WHEN** a main spec reintroduces archived placeholder purpose text
- **THEN** the repository test suite SHALL fail

### Requirement: Top-level startup stream docs stay current
Public CLI docs SHALL describe the current startup stream routing for top-level interactive entrypoints when that routing is part of the tested contract.

#### Scenario: Workbench and cockpit docs mention stderr seam routing
- **WHEN** bare `lango` workbench and `lango cockpit` use seam-aware stderr startup notices
- **THEN** the public core CLI docs SHALL mention that startup notice routing

### Requirement: README top-level command summaries stay stream-aware
README.md SHALL keep its top-level command summaries aligned with the current stdout/stderr routing contracts when those contracts are already part of the tested CLI surface.

#### Scenario: README mentions top-level stream contracts
- **WHEN** top-level utility and TUI entrypoint stream-routing behavior is part of the current tested contract
- **THEN** README SHALL summarize the stdout/stderr routing at a high level
