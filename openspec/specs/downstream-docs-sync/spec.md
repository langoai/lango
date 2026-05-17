## Purpose

Requirements for keeping downstream artifacts (documentation, prompts, Docker config, Makefile) synchronized with core feature changes.
## Requirements
### Requirement: Prompt files reflect all tool categories
The system prompts SHALL list all current tool categories including the Team category, with accurate tool counts and descriptions for all 7 team tools.

#### Scenario: AGENTS.md tool count
- **WHEN** a user reads `prompts/AGENTS.md`
- **THEN** the document SHALL state "fifteen" tool categories and include a Team category section

#### Scenario: TOOL_USAGE.md team tools
- **WHEN** a user reads `prompts/TOOL_USAGE.md`
- **THEN** the document SHALL contain a Team Tool section documenting `team_form`, `team_delegate`, `team_status`, `team_list`, `team_disband`, `team_form_with_budget`, `team_complete_milestone` with parameters and return values

### Requirement: README reflects all implemented features
The README SHALL list all implemented features including Team Health Monitoring, Incremental Git Bundles, Task Branch Management, Config Presets, Event-Driven Bridges, EventMonitor Reorg Protection, and Escrow Hub V2.

#### Scenario: New features in README
- **WHEN** a user reads `README.md`
- **THEN** all 7 new feature areas SHALL be listed in the features section

#### Scenario: CLI commands in README
- **WHEN** a user reads the CLI commands section of `README.md`
- **THEN** `lango status`, `lango onboard --preset`, cron `--timeout`, cron `--deliver`, and cron management by `id-or-name` SHALL be documented

#### Scenario: Provenance quick reference includes required operands
- **WHEN** a user reads the provenance quick reference in `README.md`
- **THEN** it SHALL include `lango provenance checkpoint list --run <id>`, `lango provenance checkpoint create <label> --run <id>`, and `lango provenance checkpoint show <id>`
- **AND** it SHALL include `lango provenance session tree <session-key>` and `lango provenance session list`
- **AND** it SHALL include `lango provenance attribution show <session-key>` and `lango provenance attribution report <session-key>`
- **AND** it SHALL include `lango provenance bundle export <session-key>` and `lango provenance bundle import <file>`

#### Scenario: P2P reputation quick reference includes required peer DID
- **WHEN** a user reads the P2P quick reference in `README.md`
- **THEN** it SHALL include `lango p2p reputation --peer-did <did>`
- **AND** workspace quick-reference summaries SHALL describe direct local workspace actions for create, list, status, join, and leave

#### Scenario: README quick references include required memory and P2P operands
- **WHEN** a user reads the CLI commands section of `README.md`
- **THEN** it SHALL include `lango memory clear <session-key>`
- **AND** it SHALL include `lango p2p firewall add --peer-did <did>`
- **AND** it SHALL include `lango p2p firewall remove <peer-did>`
- **AND** it SHALL include `lango p2p session revoke --peer-did <did>`

#### Scenario: README config get quick reference includes output and secret flags
- **WHEN** a user reads the config quick reference in `README.md`
- **THEN** it SHALL include `lango config get <dot.path> [--output plain|json] [--show-secrets]`

### Requirement: CLI index quick references include required operands
The CLI index SHALL list quick-reference commands with required positional
arguments and required flags for provenance and P2P reputation commands.

#### Scenario: CLI index provenance quick reference includes required operands
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango provenance checkpoint list --run <id>`, `lango provenance checkpoint create <label> --run <id>`, and `lango provenance checkpoint show <id>`
- **AND** it SHALL include `lango provenance session tree <session-key>` and `lango provenance session list`
- **AND** it SHALL include `lango provenance attribution show <session-key>` and `lango provenance attribution report <session-key>`
- **AND** it SHALL include `lango provenance bundle export <session-key>` and `lango provenance bundle import <file>`

#### Scenario: CLI index P2P reputation includes required peer DID
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango p2p reputation --peer-did <did>`
- **AND** workspace quick-reference summaries SHALL describe direct local workspace actions for create, list, status, join, and leave

#### Scenario: CLI index quick references include required memory and P2P operands
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango memory clear <session-key>`
- **AND** it SHALL include `lango p2p firewall add --peer-did <did>`
- **AND** it SHALL include `lango p2p firewall remove <peer-did>`
- **AND** it SHALL include `lango p2p session revoke --peer-did <did>`

#### Scenario: CLI index config get quick reference includes output and secret flags
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango config get <dot.path> [--output plain|json] [--show-secrets]`

#### Scenario: Config CLI docs explain dynamic key placeholders
- **WHEN** a user reads `docs/cli/config.md`
- **THEN** the config keys section SHALL show dynamic map-backed templates such as `providers.<name>.apiKey`, `mcp.servers.<name>.env.<key>`, and `mcp.servers.<name>.headers.<key>`
- **AND** it SHALL explain what `<name>` and `<key>` represent

### Requirement: Feature docs include required command operands
Public feature docs SHALL include required operands when showing runnable command examples.

#### Scenario: Observational memory docs include clear session key
- **WHEN** a user reads `docs/features/observational-memory.md`
- **THEN** it SHALL include `lango memory clear <session-key>`

#### Scenario: P2P feature docs include firewall and session peer operands
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** it SHALL include `lango p2p firewall add --peer-did <did>`
- **AND** it SHALL include `lango p2p session revoke --peer-did <did>`

#### Scenario: ZKP feature docs include session revoke peer operand
- **WHEN** a user reads `docs/features/zkp.md`
- **THEN** it SHALL include `lango p2p session revoke --peer-did <did>`

### Requirement: Public cron automation docs match CLI flags
Public cron automation documentation SHALL use command examples accepted by the current CLI.

#### Scenario: Cron docs show accepted add flags
- **WHEN** a user reads `docs/automation/cron.md`
- **THEN** the add examples SHALL use accepted delivery flags
- **AND** per-job timeout examples SHALL match `lango cron add --timeout`

#### Scenario: Cron docs show accepted control selectors
- **WHEN** a user reads `docs/automation/cron.md`
- **THEN** pause, resume, delete, and job-specific history examples SHALL use accepted cron job selectors

### Requirement: Config presets documentation exists
A dedicated documentation page SHALL exist for config presets describing all 4 presets with feature matrices.

#### Scenario: Presets doc page
- **WHEN** a user navigates to `docs/features/config-presets.md`
- **THEN** the page SHALL document `minimal`, `researcher`, `collaborator`, `full` presets with feature flags

### Requirement: Status command documentation exists
A dedicated CLI reference page SHALL exist for the `lango status` command.

#### Scenario: Status doc page
- **WHEN** a user navigates to `docs/cli/status.md`
- **THEN** the page SHALL document `--output` flag, `--addr` flag, output sections, and JSON schema

### Requirement: TUI core docs describe local runtime status boundaries
Public TUI core documentation SHALL describe which configured features may still initialize in local interactive mode and how `/status` distinguishes configured intent from active runtime state.

#### Scenario: TUI core docs describe MCP runtime status truthfully
- **WHEN** a user reads `docs/cli/core.md`
- **THEN** the docs SHALL state that local TUI `/status` reports MCP as active when the local interactive bootstrap initialized MCP

### Requirement: Focused chat docs describe setup readiness gating
Public CLI documentation SHALL explain that focused chat uses the same setup readiness contract as the workbench.

#### Scenario: Core docs explain chat setup-required guidance
- **WHEN** a user reads `docs/cli/core.md`
- **THEN** the focused chat documentation SHALL state that incomplete profiles show setup guidance before normal turns
- **AND** the documentation SHALL mention `lango onboard`, `lango settings`, and `lango doctor`

### Requirement: Bare root docs describe non-interactive fallback
Public CLI documentation that describes bare `lango` SHALL state that the workbench launch requires an interactive terminal and that non-interactive bare-root execution prints help instead of starting the TUI.

#### Scenario: Public bare root docs distinguish interactive and non-interactive behavior
- **WHEN** a user reads README, `docs/cli/index.md`, or `docs/cli/core.md`
- **THEN** the document SHALL state that bare `lango` launches the mission workbench in an interactive terminal
- **AND** the document SHALL state that non-interactive bare `lango` prints help and exits successfully without starting the TUI
- **AND** the document SHALL distinguish this fallback from `lango cockpit` and `lango chat` non-interactive errors

### Requirement: Background CLI docs describe server-boundary caveat
Public documentation that lists `lango bg` commands SHALL explain that the current root CLI is not a remote client for the server process's in-memory background manager.

#### Scenario: Public bg command references include runtime caveat
- **WHEN** a user reads README, `docs/cli/index.md`, or `docs/automation/background.md`
- **THEN** any `lango bg list/status/cancel/result` command reference SHALL be accompanied by a caveat that task state is in-memory and root CLI management is not yet connected to the running server process

### Requirement: Feature index pages updated
The feature index pages SHALL include cards for P2P Workspaces, P2P Teams, and Config Presets.

#### Scenario: Feature index cards
- **WHEN** a user reads `docs/features/index.md` or `docs/index.md`
- **THEN** cards for Workspaces, Teams, and Config Presets SHALL be present

### Requirement: Makefile test targets for new packages
The Makefile SHALL provide dedicated test targets for team, economy, and bridge packages.

#### Scenario: Makefile test-team target
- **WHEN** a user runs `make test-team`
- **THEN** tests in `./internal/p2p/team/...` SHALL execute

#### Scenario: Makefile test-economy target
- **WHEN** a user runs `make test-economy`
- **THEN** tests in `./internal/economy/...` SHALL execute

#### Scenario: Makefile test-bridges target
- **WHEN** a user runs `make test-bridges`
- **THEN** bridge-related tests SHALL execute

### Requirement: Docker config supports workspaces
The Docker Compose configuration SHALL include a workspace volume and team/economy environment variables.

#### Scenario: Docker workspace volume
- **WHEN** a user reads `docker-compose.yml`
- **THEN** a `lango-workspaces` volume SHALL be defined

### Requirement: README advanced setup paths stay truth-aligned
README guidance for advanced configuration SHALL point only to setup paths that exist in the current product surfaces.

#### Scenario: README avoids false onboard submenu navigation
- **WHEN** a user reads README guidance for prompts, embedding, graph, multi-agent, A2A, security mode, or OIDC auth
- **THEN** the document SHALL point to real setup paths such as `lango settings` or `lango config import/export`
- **AND** SHALL NOT describe nonexistent advanced onboard submenu navigation

### Requirement: Browser action docs mention action-specific required inputs
Public documentation SHALL describe the action-specific required inputs for `browser_action`.

#### Scenario: README and CLI docs mention browser action input contract
- **WHEN** a user reads the multi-agent browser tool surfaces in `README.md` or `docs/cli/agent-memory.md`
- **THEN** those docs SHALL mention that `browser_action` click/get_text/get_element_info/wait require `selector`
- **AND** SHALL mention that `type` requires both `selector` and `text`
- **AND** SHALL mention that `eval` requires JavaScript in `text`

### Requirement: Browser search and navigate docs mention top-level required inputs
Public documentation SHALL describe the top-level required inputs for `browser_search` and `browser_navigate`.

#### Scenario: README and CLI docs mention browser search/navigation contract
- **WHEN** a user reads the multi-agent browser tool surfaces in `README.md` or `docs/cli/agent-memory.md`
- **THEN** those docs SHALL mention that `browser_search` requires `query`
- **AND** SHALL mention that `browser_navigate` requires `url`

### Requirement: Librarian web retrieval docs stay truth-aligned
Public documentation SHALL describe that lightweight `web_search` and `web_fetch` route through the librarian role rather than the interactive browser role.

#### Scenario: README and multi-agent docs mention librarian web retrieval
- **WHEN** a user reads the librarian role description in `README.md` or `docs/features/multi-agent.md`
- **THEN** those docs SHALL mention lightweight web retrieval via `web_search` and `web_fetch`
- **AND** SHALL distinguish that interactive browsing and screenshots still belong to `navigator`

### Requirement: Automator background docs mention required inputs
Public documentation SHALL describe the required wrapper inputs for background-task tools.

#### Scenario: README and multi-agent docs mention background input contract
- **WHEN** a user reads the automator role description in `README.md` or `docs/features/multi-agent.md`
- **THEN** those docs SHALL mention that `bg_submit` requires `prompt`
- **AND** SHALL mention that `bg_status`, `bg_result`, and `bg_cancel` require `task_id`

### Requirement: Automator workflow docs mention required inputs
Public documentation SHALL describe the required wrapper inputs for workflow tools.

#### Scenario: README and multi-agent docs mention workflow input contract
- **WHEN** a user reads the automator role description in `README.md` or `docs/features/multi-agent.md`
- **THEN** those docs SHALL mention that `workflow_status` and `workflow_cancel` require `run_id`
- **AND** SHALL mention that `workflow_save` requires both `name` and `yaml_content`

### Requirement: Automator cron docs mention required inputs
Public documentation SHALL describe the required wrapper inputs for cron tools.

#### Scenario: README and multi-agent docs mention cron input contract
- **WHEN** a user reads the automator role description in `README.md` or `docs/features/multi-agent.md`
- **THEN** those docs SHALL mention that `cron_add` requires `name`, `schedule_type`, `schedule`, and `prompt`
- **AND** SHALL mention that `cron_pause`, `cron_resume`, and `cron_remove` require `id`

### Requirement: Knowledge and memory docs mention required inputs
Public documentation SHALL describe the required wrapper inputs for graph and agent-memory tools.

#### Scenario: README and multi-agent docs mention graph and memory input contracts
- **WHEN** a user reads the librarian or chronicler role description in `README.md` or `docs/features/multi-agent.md`
- **THEN** those docs SHALL mention that `graph_traverse` requires `start_node`
- **AND** SHALL mention that `graph_query` requires `subject` or `object`
- **AND** SHALL mention that `memory_agent_save` requires `key` and `content`
- **AND** SHALL mention that `memory_agent_recall` requires `query`
- **AND** SHALL mention that `memory_agent_forget` requires `key`

### Requirement: Librarian inquiry docs mention required inputs
Public documentation SHALL describe the required wrapper inputs for librarian inquiry tools.

#### Scenario: README and multi-agent docs mention inquiry input contract
- **WHEN** a user reads the librarian role description in `README.md` or `docs/features/multi-agent.md`
- **THEN** those docs SHALL mention that `librarian_dismiss_inquiry` requires `inquiry_id`
- **AND** SHALL clarify that `librarian_pending_inquiries` keeps its parameters optional

### Requirement: Output tool docs mention retrieval inputs
Public documentation SHALL describe the required retrieval inputs for `tool_output_get`.

#### Scenario: README and tool-usage docs mention output retrieval contract
- **WHEN** a user reads `README.md` or the output tools section of `TOOL_USAGE.md`
- **THEN** those docs SHALL mention that `tool_output_get` requires `ref`
- **AND** SHALL mention that grep mode requires `pattern`

### Requirement: Built-in teammate examples use current names
Public workflow and multi-agent examples SHALL use the current built-in teammate names instead of legacy examples such as `executor` or `researcher`.

#### Scenario: Workflow docs use operator, librarian, planner
- **WHEN** a user reads workflow examples in `README.md` or `docs/cli/automation.md`
- **THEN** built-in teammate examples SHALL use current names such as `operator`, `librarian`, and `planner`

#### Scenario: Preset docs use current multi-agent names
- **WHEN** a user reads the multi-agent line in `docs/features/config-presets.md`
- **THEN** it SHALL reference current built-in teammate names such as `operator`, `librarian`, and `planner`

#### Scenario: Metrics docs use current built-in names
- **WHEN** a user reads the `lango metrics agents` example in `docs/cli/metrics.md`
- **THEN** built-in teammate examples SHALL use current names such as `operator`, `librarian`, and `planner`

### Requirement: Public multi-agent docs avoid built-in REJECT claims
Public multi-agent overviews SHALL not claim that built-in teammates emit textual `[REJECT]` markers in production.

#### Scenario: README reflects visible escalation instead of built-in REJECT text
- **WHEN** a user reads the multi-agent overview in `README.md`
- **THEN** it SHALL describe built-in teammate misrouting as a visible escalation summary
- **AND** it SHALL NOT claim that built-in teammates reject production work with textual `[REJECT]` markers

### Requirement: Exec docs mention wrapper required inputs
Public documentation SHALL describe the required wrapper inputs for the exec tool cluster.

#### Scenario: README and CLI docs mention exec input contract
- **WHEN** a user reads the multi-agent operator tool surfaces in `README.md` or `docs/cli/agent-memory.md`
- **THEN** those docs SHALL mention that `exec` and `exec_bg` require `command`
- **AND** SHALL mention that `exec_status` and `exec_stop` require the background-process `id`

### Requirement: Vault security docs mention required inputs
Public documentation SHALL describe the required tool inputs for vault crypto and secrets tools.

#### Scenario: README and CLI docs mention vault security input contracts
- **WHEN** a user reads the multi-agent vault tool surfaces in `README.md` or `docs/cli/agent-memory.md`
- **THEN** those docs SHALL mention that `crypto_encrypt`, `crypto_sign`, and `crypto_hash` require `data`
- **AND** SHALL mention that `crypto_decrypt` requires `ciphertext`
- **AND** SHALL mention that `secrets_store` requires `name` and `value`
- **AND** SHALL mention that `secrets_get` and `secrets_delete` require `name`

### Requirement: Advanced feature docs avoid false onboard submenu flows
Downstream feature documentation SHALL describe advanced feature setup using the actual onboarding and settings surfaces.

#### Scenario: Feature docs keep advanced setup paths synchronized
- **WHEN** a user reads advanced feature docs for embedding/RAG, A2A, or the knowledge graph
- **THEN** those docs SHALL describe `lango settings` and/or config import/export as the interactive setup path
- **AND** SHALL clarify that the five-step onboard wizard is only the initial bootstrap flow

### Requirement: Workbench docs mention incomplete-profile setup guidance
Public workbench documentation SHALL describe that the empty workbench state now points incomplete profiles at setup and verification commands.

#### Scenario: README and CLI/TUI docs mention setup recovery path
- **WHEN** a user reads the README workbench section or the CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that incomplete profiles are guided toward `lango onboard`, `lango settings`, and `lango doctor`

### Requirement: Workbench docs mention ready-profile starter prompts
Public workbench documentation SHALL describe that a ready profile sees starter prompts in the empty workbench state.

#### Scenario: Workbench docs mention starter prompts
- **WHEN** a user reads the README or CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that ready profiles are shown concrete starter prompts in the empty state

### Requirement: Workbench docs mention state-aware composer guidance
Public workbench documentation SHALL mention that the composer placeholder follows the same incomplete-vs-ready guidance split as the empty-state body.

#### Scenario: README and CLI/TUI docs mention composer guidance split
- **WHEN** a user reads the README or CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that incomplete profiles get setup-first composer guidance and ready profiles get starter-prompt composer guidance

### Requirement: Advanced configuration docs avoid false onboarding and file-path guidance
Public documentation for advanced configuration SHALL point to the actual interactive and programmatic configuration paths in the current product.

#### Scenario: Advanced docs use settings or import-export paths
- **WHEN** a user reads advanced configuration guidance in README or feature docs
- **THEN** that guidance SHALL point to `lango settings` and/or `lango config import/export`
- **AND** SHALL NOT claim that advanced feature setup happens through nonexistent `lango onboard` submenu paths
- **AND** SHALL NOT describe a canonical plaintext `~/.lango/config.yaml` configuration file

### Requirement: Workbench docs mention readiness-aware header summary
Public workbench documentation SHALL mention that the header summary now reports setup-required state for incomplete profiles.

#### Scenario: README and CLI/TUI docs mention setup-required header
- **WHEN** a user reads the README or CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that incomplete profiles show `Model: Setup required` in the header summary

### Requirement: Container runtime docs stay honest about the gVisor stub
Public configuration docs SHALL describe the current gVisor runtime status accurately instead of implying that a real gVisor backend is already available.

#### Scenario: Runtime tables mention gVisor stub status
- **WHEN** a user reads the runtime configuration tables in README or configuration docs
- **THEN** those tables SHALL still list `gvisor` as an accepted runtime value
- **AND** SHALL clarify that the current implementation is a stub whose explicit selection returns a runtime-unavailable error

### Requirement: Security provider docs stay aligned with supported providers and bootstrap rules
Public configuration docs SHALL describe the currently supported security signer providers accurately and SHALL not mention removed provider names.

#### Scenario: KMS operation docs mention build-tag and wiring requirements
- **WHEN** a user reads README, CLI, or security docs for KMS-backed signer setup
- **THEN** those docs SHALL mention that KMS providers require the matching build tag in the running binary
- **AND** SHALL mention that the runtime still depends on bootstrap-backed storage wiring for the key registry and secrets store

### Requirement: Security status docs match actual field semantics
Public docs for `lango security status` SHALL describe the current field semantics exposed by the command rather than older narrower examples.

#### Scenario: Security status JSON field docs stay truth-aligned
- **WHEN** a user reads the `lango security status` field descriptions
- **THEN** `signer_provider` SHALL be documented as the active provider or `unavailable` when DB-backed config could not be read non-interactively
- **AND** `db_encryption` SHALL be documented using the current payload-protection / legacy-DB status strings rather than implying active SQLCipher page encryption
- **AND** `kms_fallback` SHALL be documented as a KMS-only enabled/disabled status

### Requirement: Public workbench docs mention starter hotkeys

Public docs that describe the standalone workbench startup flow SHALL mention the ready-profile starter-prompt hotkeys once the product exposes them.

#### Scenario: Workbench docs mention starter hotkeys
- **WHEN** README or CLI/TUI docs describe the ready-profile workbench empty state
- **THEN** they SHALL mention that starter prompts are bound to `1`, `2`, and `3`
- **AND** they SHALL describe that those keys load prompts into the composer instead of leaving the prompts as passive copy

### Requirement: Public workbench docs explain starter prompt context-awareness

Public docs that describe the standalone workbench quick-start flow SHALL explain that ready-profile starter prompts adapt to the detected workspace context.

#### Scenario: Docs describe context-aware prompt behavior
- **WHEN** README or CLI/TUI docs describe ready-profile starter prompts
- **THEN** they SHALL mention that the prompts adapt to the detected workdir or repository
- **AND** they SHALL mention the Go-specific structure guidance when a `go.mod` is present

### Requirement: Public workbench docs mention Git-aware quick-start behavior

Public docs that describe standalone workbench starter prompts SHALL mention that the workbench can use live Git branch and dirty-state signals when available.

#### Scenario: Docs mention Git-aware starter prompts
- **WHEN** README or CLI/TUI docs describe context-aware starter prompts
- **THEN** they SHALL mention current-branch awareness
- **AND** they SHALL mention uncommitted-change awareness when Git metadata is available

### Requirement: Public workbench docs mention changed-target-aware quick-start behavior

Public docs that describe Git-aware workbench starter prompts SHALL mention changed-file or changed-directory awareness when that signal is available.

#### Scenario: Docs mention changed targets in starter prompts
- **WHEN** README or CLI/TUI docs describe dirty-repository starter prompt behavior
- **THEN** they SHALL mention changed-file or changed-directory awareness in addition to branch and dirty-state awareness

### Requirement: Public workbench docs mention Enter quick-start behavior

Public docs that describe the standalone workbench starter prompts SHALL mention that `Enter` seeds the default starter prompt on the empty ready-profile workbench.

#### Scenario: Docs mention Enter quick-start
- **WHEN** README or CLI/TUI docs describe ready-profile starter prompt behavior
- **THEN** they SHALL mention that pressing `Enter` seeds the default starter prompt

### Requirement: Public workbench docs mention Enter quick-start discoverability

Public docs that describe ready-profile starter prompts SHALL mention `Enter` as the default quick-start seed in addition to the numeric hotkeys.

#### Scenario: Docs mention Enter quick-start
- **WHEN** README or CLI/TUI docs describe ready-profile starter prompt behavior
- **THEN** they SHALL mention `Enter` as the default starter-prompt seed

### Requirement: Public docs describe Enter as a context-aware default

Public docs that describe the ready-profile `Enter` quick-start path SHALL describe it as seeding the default context-aware prompt rather than a fixed summary prompt.

#### Scenario: Docs describe context-aware Enter default
- **WHEN** README or CLI/TUI docs describe the ready-profile `Enter` quick-start path
- **THEN** they SHALL state that `Enter` seeds the default context-aware starter prompt

### Requirement: Public docs describe the two-step Enter quick-start flow

Public docs that describe ready-profile `Enter` quick-start behavior SHALL mention both the seed step and the submit step.

#### Scenario: Docs mention Enter seed then submit flow
- **WHEN** README or CLI/TUI docs describe the ready-profile `Enter` quick-start path
- **THEN** they SHALL mention that the first `Enter` seeds the default starter prompt
- **AND** they SHALL mention that the next `Enter` submits it

### Requirement: Public docs describe armed starter replacement

Public docs that describe seeded starter prompts SHALL mention that `1/2/3` remain available to replace the armed starter choice.

#### Scenario: Docs mention replacement after seeding
- **WHEN** README or CLI/TUI docs describe the seeded starter prompt state
- **THEN** they SHALL mention that `1/2/3` can replace the armed starter prompt

### Requirement: Public workbench docs mention post-turn next-step defaults

Public workbench documentation SHALL explain that the empty workbench changes its default `Enter` starter after a completed turn.

#### Scenario: Docs mention post-turn next-step starter
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that, after a turn completes, the empty workbench defaults `Enter` to the next-step starter instead of returning to the original summary starter

### Requirement: Public workbench docs describe context-sensitive post-turn defaults

Public workbench documentation SHALL explain that the post-turn default `Enter` starter depends on whether a workspace context is available.

#### Scenario: Docs mention generic-vs-repo post-turn default
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that post-turn defaults stay structure-oriented outside repo context and stay next-change-oriented inside detected workspaces

### Requirement: Public workbench docs use next-step wording after a completed turn

Public workbench documentation SHALL describe the completed-turn empty workbench as a next-step loop, not as the original quick-start state.

#### Scenario: Docs use next-step wording
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL describe the completed-turn default starter as the next step rather than as the initial quick-start prompt

### Requirement: Public workbench docs describe calmer empty-state warning behavior

Public workbench documentation SHALL explain that the standalone empty workbench avoids showing cockpit-style degraded warnings before any active mission/control context exists.

#### Scenario: Docs mention empty-state warning suppression
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that the empty workbench emphasizes setup and next actions instead of surfacing cockpit degraded warnings immediately

### Requirement: Public workbench docs describe the completed-turn body as a next-step state

Public workbench documentation SHALL explain that the completed-turn empty body shifts from a blank no-missions message to a next-step prompt.

#### Scenario: Docs mention completed-turn body shift
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that a completed turn changes the empty body into a next-step prompt

### Requirement: Public workbench docs mention the completed-turn next-prompt hint

Public workbench documentation SHALL explain that the completed-turn empty workbench invites the next prompt explicitly.

#### Scenario: Docs mention next-prompt hint
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that, after a completed turn, the workbench tells the operator to type the next prompt here

### Requirement: Public workbench docs mention next-step placeholder wording

Public workbench documentation SHALL explain that the completed-turn placeholder uses next-step wording.

#### Scenario: Docs mention next-step placeholder
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that the completed-turn composer placeholder switches to next-step wording

### Requirement: Public workbench docs mention next-prompt footer wording

Public workbench documentation SHALL explain that the completed-turn footer uses next-prompt wording.

#### Scenario: Docs mention next-prompt footer
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that the completed-turn footer switches to next-prompt wording

### Requirement: Public workbench docs mention the completed-turn result preview

Public workbench documentation SHALL explain that the completed-turn empty workbench surfaces the latest result directly in the body.

#### Scenario: Docs mention completed-turn result preview
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that the completed-turn body shows a compact preview of the latest assistant result

### Requirement: Public workbench docs use neutral completed-turn wording

Public workbench documentation SHALL describe the completed-turn state with neutral wording that does not imply success.

#### Scenario: Docs say finished rather than complete
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL describe the previous turn as finished rather than completed

### Requirement: Public workbench docs mention failure-aware completed-turn wording

Public workbench documentation SHALL explain that failed turns are called out explicitly in the completed-turn empty state.

#### Scenario: Docs mention failed-turn wording
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that failed turns surface an attention-oriented completed-turn lead

### Requirement: Public workbench docs describe the compact result preview cleanly

Public workbench documentation SHALL describe the completed-turn result preview as a compact result summary rather than as a repeated assistant label.

#### Scenario: Docs mention compact result preview
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL describe the completed-turn body as showing a compact latest result preview

### Requirement: Public workbench docs mention recovery wording after failed turns

Public workbench documentation SHALL explain that failed completed-turn states switch to recovery-specific wording.

#### Scenario: Docs mention recovery wording
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that failed turns pivot the next prompt loop into recovery wording

### Requirement: Public workbench docs mention recovery-footer wording

Public workbench documentation SHALL explain that failed completed-turn footers use recovery wording.

#### Scenario: Docs mention recovery footer
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that the failed completed-turn footer switches to recovery-prompt wording

### Requirement: Public workbench docs mention the recovery-step lead

Public workbench documentation SHALL explain that failed completed-turn states use recovery-step wording in the body lead.

#### Scenario: Docs mention recovery-step lead
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that failed turns tell the operator to pick the recovery step

### Requirement: Public workbench docs mention the recovery-oriented Enter default

Public workbench documentation SHALL explain that failed completed-turn states seed a recovery-oriented starter on `Enter`.

#### Scenario: Docs mention recovery Enter default
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that failed turns seed a recovery-oriented starter on `Enter`

### Requirement: Payment execution docs describe the validation-vs-deny split

Public payment and security docs SHALL describe empty transaction receipt ids as validation failures and reserve `missing_receipt` for missing referenced receipt state after request validation succeeds.

#### Scenario: Payment execution docs distinguish validation failures from deny reasons
- **WHEN** a user reads the direct payment execution gate docs
- **THEN** those docs SHALL describe an empty `transaction_receipt_id` as an actionable validation error
- **AND** SHALL describe `missing_receipt` as a deny reason for missing referenced receipt state rather than missing input

### Requirement: Settlement execution docs describe the validation-vs-deny split

Public settlement execution docs SHALL describe empty transaction receipt ids as validation failures and reserve deny reasons for referenced receipt or settlement state after request validation succeeds.

#### Scenario: Settlement execution docs distinguish validation failures from deny reasons
- **WHEN** a user reads the actual settlement execution docs
- **THEN** those docs SHALL describe an empty `transaction_receipt_id` as an actionable validation error
- **AND** SHALL describe `missing_receipt` and the other deny reasons as post-validation execution outcomes

### Requirement: P2P CLI git-bundle examples stay truth-aligned

Public P2P CLI docs SHALL show workspace/git runtime examples using the actual registered tool names and parameter names.

#### Scenario: P2P git example uses actual workspace tool names
- **WHEN** a user reads the git-bundle workflow example in the P2P CLI docs
- **THEN** the example SHALL use `workspaceId` instead of `workspace_id`
- **AND** SHALL reference only the currently registered runtime tools such as `p2p_git_init`, `p2p_git_push`, `p2p_git_log`, and `p2p_git_diff`

### Requirement: Filesystem prompt guidance mentions fs_list default path

Prompt guidance for the filesystem tool SHALL mention that `fs_list` defaults to the current working directory when `path` is omitted.

#### Scenario: TOOL_USAGE mentions fs_list current-directory default
- **WHEN** an agent reads the filesystem section of `TOOL_USAGE.md`
- **THEN** it SHALL find that `fs_list` accepts an optional `path`
- **AND** that omitting `path` lists the current working directory

### Requirement: Filesystem prompt guidance mentions write/edit required inputs

Prompt guidance for the filesystem tool SHALL describe the required-input contract for `fs_write` and `fs_edit`.

#### Scenario: TOOL_USAGE mentions write/edit required inputs
- **WHEN** an agent reads the filesystem section of `TOOL_USAGE.md`
- **THEN** it SHALL find that `fs_write` requires `path` and `content`
- **AND** that `fs_edit` requires `path`, `startLine`, `endLine`, and `content`

### Requirement: Public cockpit docs describe current page availability
Public cockpit documentation SHALL describe the current page roster and runtime availability behavior instead of older assumptions about always-live or disabled optional destinations.

#### Scenario: Cockpit feature page describes current degraded-page routing
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** the page SHALL describe the current 9-item sidebar roster headed by Mission Control
- **AND** it SHALL explain that cockpit pages remain routable and surface degraded in-page messaging when backing services are unavailable
- **AND** it SHALL describe Dead Letters as an always-registered degraded page rather than as a disabled destination

### Requirement: Public cockpit docs describe Dead Letters as a degraded page
Public cockpit documentation SHALL describe Dead Letters as an always-registered cockpit page that degrades to unavailable messaging when its bridge is absent.

#### Scenario: Cockpit feature page describes degraded Dead Letters availability
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** it SHALL describe Dead Letters as always available in the cockpit roster
- **AND** it SHALL explain that the page surfaces unavailable/degraded messaging until the dead-letter bridge is ready

### Requirement: Home page uses current built-in teammate names
The public home page SHALL describe multi-agent orchestration using current built-in teammate names rather than legacy examples.

#### Scenario: Home page feature card uses current built-in teammates
- **WHEN** a user reads the multi-agent feature card in `docs/index.md`
- **THEN** the card SHALL use current built-in teammate names such as `Operator`, `Librarian`, `Planner`, or `Vault`
- **AND** it SHALL NOT describe the system with legacy built-in names such as `Executor`, `Researcher`, or `Memory Manager`

### Requirement: README describes Dead Letters as a degraded cockpit page
The public README SHALL describe the cockpit Dead Letters page using the same degraded-page contract as the runtime and feature docs.

#### Scenario: README cockpit shortcut table uses degraded-page wording
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Dead Letters row SHALL describe the page as available with degraded unavailable messaging until the dead-letter bridge is ready
- **AND** it SHALL NOT imply that the page appears only when the bridge is available

### Requirement: README describes Tasks and Approvals as degraded cockpit pages when dependencies are absent
The public README SHALL describe the cockpit Tasks and Approvals pages using the same degraded-surface contract as the runtime and cockpit feature docs.

#### Scenario: README cockpit shortcut table uses degraded-page wording for Tasks and Approvals
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Tasks row SHALL mention unavailable/degraded messaging when the background task manager is absent
- **AND** the Approvals row SHALL mention unavailable/degraded messaging when approval stores are absent

### Requirement: Public cockpit docs describe Status page unavailable messaging
Public cockpit documentation SHALL describe the concrete unavailable-state behavior of the Status page.

#### Scenario: Cockpit feature page describes status dependency gaps
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** it SHALL explain that a missing feature-status provider surfaces explicit unavailable messaging in the Feature Status section
- **AND** it SHALL explain that a missing observability collector surfaces explicit unavailable messaging in the Token Usage, Tool Execution, and Graph Admission sections

### Requirement: Public cockpit docs describe the Tools page as a degraded surface
Public cockpit documentation SHALL describe the Tools page using the same degraded-surface contract as the runtime.

#### Scenario: README describes the empty-catalog Tools state
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Tools row SHALL mention that a configured catalog with zero categories surfaces an explicit no-categories state

### Requirement: README and CLI overview describe cockpit core pages as degraded surfaces when dependencies are absent
The public README and CLI overview SHALL describe cockpit core pages using the current degraded-page routing contract.

#### Scenario: CLI overview describes Chat idle and failed quit semantics
- **WHEN** a user reads the `lango cockpit` overview in `docs/cli/core.md`
- **THEN** the Chat key-binding summary SHALL describe `Ctrl+C` as a double-press quit path for idle or failed states

### Requirement: Quickstart describes the workbench/cockpit split correctly
The public quickstart guide SHALL describe bare `lango` and `lango cockpit` using the current workbench/cockpit entrypoint split.

#### Scenario: Quickstart TUI tip uses current entrypoints
- **WHEN** a user reads the interactive TUI tip in `docs/getting-started/quickstart.md`
- **THEN** it SHALL describe bare `lango` as the standalone mission workbench
- **AND** it SHALL describe `lango cockpit` as the explicit multi-panel operator dashboard

### Requirement: Public cockpit docs describe Sessions page behavior
Public cockpit documentation SHALL describe the Sessions page using the current runtime contract.

#### Scenario: README describes Sessions ordering and empty-state split
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Sessions row SHALL describe the page as a newest-first session summary list
- **AND** it SHALL mention the page's explicit unavailable or empty-state messaging

### Requirement: Quickstart TUI tip title does not imply a chat-only surface
The public quickstart guide SHALL not title the bare-`lango` entrypoint tip as if it were a chat-only surface.

#### Scenario: Quickstart tip title uses neutral TUI wording
- **WHEN** a user reads the interactive TUI tip in `docs/getting-started/quickstart.md`
- **THEN** the tip title SHALL use neutral wording such as `Interactive TUI`
- **AND** it SHALL NOT title the workbench entrypoint as `Interactive TUI Chat`

### Requirement: README uses neutral request wording for the Mission Control composer
The public README SHALL describe the first-screen Mission Control composer using neutral request-entry wording rather than chat-only wording.

#### Scenario: README Mission Control overview uses request wording
- **WHEN** a user reads the Mission Control overview block in `README.md`
- **THEN** it SHALL describe the shared composer as a place to type a request
- **AND** it SHALL NOT describe the first-screen entry point only as "type directly into the shared composer" for chat-oriented behavior

### Requirement: Public cockpit docs describe the Mission Control composer with neutral request wording
Public cockpit documentation SHALL describe the default first-screen Mission Control composer hint using neutral request wording rather than chat-only wording.

#### Scenario: Cockpit feature page uses request wording
- **WHEN** a user reads the Mission Control composer description in `docs/features/cockpit.md`
- **THEN** it SHALL describe the hint as `Type a request here, or use lango chat for focused chat`
- **AND** it SHALL NOT describe the default first-screen hint as `Type to chat here`

### Requirement: Cockpit feature docs describe the Dead Letters operator surface
The public cockpit feature reference SHALL describe the current Dead Letters page beyond simple roster availability.

#### Scenario: Cockpit feature page includes Dead Letters section
- **WHEN** `docs/features/cockpit.md` documents cockpit pages
- **THEN** it SHALL include a dedicated Dead Letters section
- **AND** it SHALL describe filter controls, retry request flow, and degraded/unavailable or load-failure states

### Requirement: Cockpit feature docs describe the Tools operator surface
The public cockpit feature reference SHALL describe the current Tools page beyond simple roster availability.

#### Scenario: Cockpit feature page includes Tools section
- **WHEN** `docs/features/cockpit.md` documents cockpit pages
- **THEN** it SHALL include a dedicated Tools section
- **AND** it SHALL describe category cursor navigation, immediate detail-panel updates, and degraded catalog states

### Requirement: Cockpit feature docs describe the Settings and Status operator surfaces
The public cockpit feature reference SHALL describe the current Settings and Status pages beyond simple roster availability.

#### Scenario: Cockpit feature page describes Settings save feedback
- **WHEN** the public cockpit feature reference describes the Settings page
- **THEN** it SHALL explain that embedded saves surface inline success or failure banners at the top of the Settings menu

### Requirement: Cockpit feature docs describe the Mission Control key surface
The public cockpit feature reference SHALL describe the current Mission Control key surface beyond the general page overview.

#### Scenario: Cockpit feature page includes Mission Control keys subsection
- **WHEN** `docs/features/cockpit.md` documents Mission Control
- **THEN** it SHALL include a dedicated key-surface subsection
- **AND** it SHALL describe the `tab`/`enter` core actions, populated-state `↑/↓` navigation, and the reduced help surface in the true empty state

### Requirement: Cockpit feature docs describe the Chat operator surface
The public cockpit feature reference SHALL describe the current Chat page beyond simple roster availability.

#### Scenario: Cockpit feature page describes confirm-pending deny path
- **WHEN** the public cockpit feature reference describes the critical-risk double-press guardrail
- **THEN** it SHALL explain that `d` or `Esc` still deny the request immediately while confirm-pending is active

### Requirement: Public workflow docs describe scheduled workflow registration truthfully
Public automation documentation SHALL describe the current behavior of `lango workflow run --schedule` without stale "not implemented" wording.

#### Scenario: CLI automation docs show cron-backed registration
- **WHEN** `docs/cli/automation.md` documents `lango workflow run --schedule`
- **THEN** it SHALL explain that the command validates the workflow and registers an enabled cron job that asks runtime automation to invoke `workflow_run`
- **AND** it SHALL NOT instruct operators that CLI schedule registration is unavailable
