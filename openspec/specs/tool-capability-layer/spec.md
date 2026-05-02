## Purpose

The Tool Capability Layer enriches built-in tools with structured metadata (`ToolCapability`) and exposure policies that drive discovery, search ranking, prompt visibility, and runtime access control. It unifies how the catalog, dispatcher, orchestrator, and middleware consume tool metadata.
## Requirements
### Requirement: ToolCapability struct with backward-compatible zero values
The `agent.ToolCapability` struct SHALL have fields whose zero values preserve existing behavior. A `Tool` created without setting `Capability` SHALL behave identically to pre-capability code.

#### Scenario: Zero-value ToolCapability preserves visibility
- **GIVEN** a `Tool` with no `Capability` field set
- **WHEN** `tool.Capability.Exposure.IsVisible()` is evaluated
- **THEN** it SHALL return `true` (ExposureDefault is visible)

#### Scenario: Zero-value ToolCapability preserves safety assumptions
- **GIVEN** a `Tool` with no `Capability` field set
- **WHEN** `tool.Capability.ReadOnly` and `tool.Capability.ConcurrencySafe` are evaluated
- **THEN** both SHALL be `false` (fail-safe: assume mutation possible, not concurrency safe)

#### Scenario: Zero-value slices are nil
- **GIVEN** a `Tool` with no `Capability` field set
- **WHEN** `tool.Capability.Aliases`, `tool.Capability.SearchHints`, and `tool.Capability.RequiredCapabilities` are evaluated
- **THEN** all SHALL be `nil`

### Requirement: ExposurePolicy behavior
The `ExposurePolicy` enum SHALL define four levels controlling tool visibility in agent prompts and search results.

#### Scenario: ExposureDefault is visible
- **GIVEN** a tool with `Exposure: ExposureDefault`
- **WHEN** `IsVisible()` is called
- **THEN** it SHALL return `true`

#### Scenario: ExposureAlwaysVisible is visible
- **GIVEN** a tool with `Exposure: ExposureAlwaysVisible`
- **WHEN** `IsVisible()` is called
- **THEN** it SHALL return `true`

#### Scenario: ExposureDeferred is search-only
- **GIVEN** a tool with `Exposure: ExposureDeferred`
- **WHEN** `IsVisible()` is called
- **THEN** it SHALL return `false`
- **AND** the tool SHALL appear in `SearchableEntries()` results
- **AND** the tool SHALL appear in `builtin_search` results

#### Scenario: ExposureHidden is never shown
- **GIVEN** a tool with `Exposure: ExposureHidden`
- **WHEN** `IsVisible()` is called
- **THEN** it SHALL return `false`
- **AND** the tool SHALL NOT appear in `SearchableEntries()` results
- **AND** the tool SHALL NOT appear in `builtin_search` results
- **AND** the tool SHALL NOT appear in `builtin_list` results

### Requirement: builtin_search ranking
The `SearchIndex` SHALL rank results using weighted scoring: exact name match (10) > name prefix (8) > exact alias (7) > alias prefix (5) > search hint (4) > category (3) > description (2) > activity (1). Multi-token queries SHALL sum scores across all tokens. Ties SHALL be broken by name ascending.

#### Scenario: Exact name outranks alias
- **GIVEN** a catalog containing tool `fs_read` with alias `cat`
- **WHEN** `builtin_search` is invoked with query `fs_read`
- **THEN** `fs_read` SHALL be the first result with score 10 and match_field `name`

#### Scenario: Alias outranks description
- **GIVEN** a catalog containing tool `fs_read` with alias `cat`
- **WHEN** `builtin_search` is invoked with query `cat`
- **THEN** `fs_read` SHALL be the first result with match_field `alias`
- **AND** the score SHALL be greater than any description-only match

#### Scenario: Search hint outranks description
- **GIVEN** a catalog containing tool `exec_shell` with search hint `terminal`
- **WHEN** `builtin_search` is invoked with query `terminal`
- **THEN** `exec_shell` SHALL appear with match_field `search_hint` and score 4

#### Scenario: Multi-token scores are additive
- **GIVEN** a catalog containing tool `fs_read` with alias `read` and search hint `file`
- **WHEN** `builtin_search` is invoked with query `file read`
- **THEN** `fs_read` SHALL score the sum of the search hint weight (4) and the exact alias weight (7)

#### Scenario: Deferred tools are searchable
- **GIVEN** a catalog containing a deferred tool `fs_list` with alias `ls`
- **WHEN** `builtin_search` is invoked with query `ls`
- **THEN** `fs_list` SHALL appear in the results

#### Scenario: Hidden tools are not searchable
- **GIVEN** a catalog containing a hidden tool `browser_internal`
- **WHEN** `builtin_search` is invoked with query `browser_internal`
- **THEN** `browser_internal` SHALL NOT appear in the results

### Requirement: builtin_list filtering
The `builtin_list` tool SHALL return only visible tools (ExposureDefault or ExposureAlwaysVisible) and include a deferred count hint when deferred tools exist.

#### Scenario: Visible tools listed
- **GIVEN** a catalog with 3 visible tools and 2 deferred tools
- **WHEN** `builtin_list` is invoked with no parameters
- **THEN** the `tools` array SHALL contain exactly 3 entries
- **AND** the `deferred_count` field SHALL equal 2
- **AND** a `hint` field SHALL contain "builtin_search"

#### Scenario: No hint when no deferred tools
- **GIVEN** a catalog with only visible tools
- **WHEN** `builtin_list` is invoked
- **THEN** the result SHALL NOT contain a `hint` field

### Requirement: builtin_invoke works for visible and deferred tools
The `builtin_invoke` tool SHALL execute any safe tool in the catalog regardless of exposure policy.

#### Scenario: Invoke visible safe tool
- **GIVEN** a visible safe tool `browser_navigate` in the catalog
- **WHEN** `builtin_invoke` is called with `tool_name: "browser_navigate"`
- **THEN** it SHALL execute the handler and return `{tool, result}`

#### Scenario: Invoke deferred safe tool
- **GIVEN** a deferred safe tool `fs_list` in the catalog
- **WHEN** `builtin_invoke` is called with `tool_name: "fs_list"`
- **THEN** it SHALL execute the handler and return `{tool, result}`

#### Scenario: Block dangerous tool regardless of exposure
- **GIVEN** a dangerous tool `exec_shell` in the catalog (any exposure)
- **WHEN** `builtin_invoke` is called with `tool_name: "exec_shell"`
- **THEN** it SHALL return an error containing "requires approval"
- **AND** it SHALL NOT execute the handler

### Requirement: Orchestrator uses capability summaries
The orchestrator routing table SHALL describe sub-agent capabilities via capability summaries and tool counts, not individual tool name lists. This reduces prompt size and avoids prompt-injection risks from tool names.

#### Scenario: Routing entry format
- **WHEN** the orchestrator instruction is built
- **THEN** each routing entry SHALL contain a `Capabilities` line and a `Tool count` indicator
- **AND** individual tool names SHALL NOT appear in the routing table

### Requirement: Chain() preserves Capability field
The `toolchain.Chain()` function SHALL copy the `Capability` field from the original tool to the wrapped tool.

#### Scenario: Middleware wrapping preserves capability
- **GIVEN** a tool with `Capability{Aliases: ["cat"], Exposure: ExposureDeferred, ReadOnly: true}`
- **WHEN** `Chain(tool, someMiddleware)` is called
- **THEN** the returned tool SHALL have the same `Capability` field values

### Requirement: DynamicAllowedTools with runtime essentials

`DynamicAllowedTools` remains the runtime ceiling for tool execution. For built-in teammate types, that ceiling is interpreted under the teammate role maximum scope. For custom or non-built-in teammate paths, the current allowlist itself SHALL be treated as the effective ceiling because no built-in role registry applies.

#### Scenario: Custom teammate uses current allowlist as its ceiling
- **GIVEN** a non-built-in teammate with `CurrentAllowed: ["exec"]`
- **WHEN** capability policy evaluates tool `exec`
- **THEN** the runtime SHALL treat `exec` as in-scope because it is already present in the current allowlist
- **AND** the runtime SHALL NOT require a built-in teammate role definition for that decision

#### Scenario: Custom teammate cannot escalate beyond current allowlist
- **GIVEN** a non-built-in teammate with `CurrentAllowed: ["fs_read"]`
- **WHEN** capability policy evaluates tool `exec`
- **THEN** the runtime SHALL treat `exec` as out-of-scope
- **AND** the current allowlist SHALL act as the effective ceiling for that teammate path

### Requirement: Capability runtime rechecks grant state before blocked projection
Before persisting a `blocked_waiting_approval` projection for a built-in teammate run, the runtime SHALL re-read the latest grant and allowlist state. The check narrows the stale projection window but correctness is measured on the final observed run state rather than every intermediate transition.

#### Scenario: Final observed run state is clear after grant wins the race
- **WHEN** a grant becomes effective before the final observed state is read
- **THEN** the final observed teammate run state SHALL NOT remain stuck in `blocked_waiting_approval`
- **AND** the runtime regression tests SHALL be the concrete verification gate for that condition

### Requirement: Teammate approval request identity policy
The teammate capability layer SHALL define approval identity using a stable logical `GrantRequestID` per `(runID, toolName)`. Repeated approval blocks for the same logical request SHALL reuse that `GrantRequestID` and SHALL surface renewed-attempt semantics through explicit attempt metadata.

For this slice, the required attempt metadata is:

- `grant_attempt` — integer counter for the current active blocked cycle
- `grant_state` — one of `pending`, `granted`, `denied`, or empty when no approval cycle is currently active

The current slice does **not** introduce a separate long-lived cycle identifier. It only preserves the latest active blocked cycle semantics.

#### Scenario: First approval block initializes stable identity
- **WHEN** a built-in teammate first blocks on a dangerous in-scope tool
- **THEN** the runtime SHALL assign a stable logical `GrantRequestID`
- **AND** the runtime SHALL initialize `grant_attempt = 1`
- **AND** the runtime SHALL expose a pending-style attempt state

#### Scenario: Repeated block for the same run and tool reuses the logical identity
- **WHEN** the same built-in teammate run blocks again on the same tool while it is still in the same active approval-blocked cycle
- **THEN** the runtime SHALL reuse the same logical `GrantRequestID`
- **AND** it SHALL increment separate attempt metadata instead of rotating the logical request ID

#### Scenario: Grant or denial ends the active attempt cycle
- **WHEN** a blocked approval request is granted or denied
- **THEN** the runtime SHALL clear the active blocked projection state
- **AND** it SHALL clear the active-cycle attempt metadata

#### Scenario: Later fresh block starts a new active cycle with the same logical request ID
- **WHEN** the same `(run, tool)` blocks again after a prior grant or denial cleared the active blocked cycle
- **THEN** the runtime SHALL reuse the same logical `GrantRequestID`
- **AND** it SHALL restart `grant_attempt` at `1`
- **AND** that later block SHALL be treated as the latest active cycle of the same logical request rather than as a separately identified historical cycle

