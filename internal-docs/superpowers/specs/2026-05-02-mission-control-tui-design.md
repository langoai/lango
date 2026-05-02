# Mission Control TUI Design

## Strategic Product Shift

Lango should move from a chat-first or dashboard-first terminal interface to a mission-native agent operating surface. The user should feel like they are directing ongoing work across life, work, research, automation, and coding, not merely typing prompts into a coding assistant.

The default `lango` experience should make agent activity, current missions, pending decisions, and meaningful progress visible immediately. Chat remains available, but it is not the primary information architecture.

## North Star

Lango is a mission-native agent OS:

- The user directs missions instead of managing raw tasks, runs, tools, and sessions.
- Agents may discover proposed missions and prepare useful work before asking the user to commit.
- The user keeps control over irreversible, external, expensive, or risky actions through explicit live decisions.
- Activity is presented in human work units: goals, next actions, blockers, decisions, and outcomes.

## Program Waves

### Wave 1: Mission Control Projection And Producer Hardening

Use existing runtime data to create a Mission Control surface. Do not introduce a new mission domain engine yet. This wave is not "projection only" in the strict sense: some producers are already wired, while others require new UI-facing projectors or dependency wiring before they can be used honestly.

Inputs:

| Source | Wave 1 status | Notes |
| --- | --- | --- |
| Background tasks | wired | `background.Manager.List()` already provides snapshots and is adapted by the existing Tasks page. |
| Approval history and grants | wired after decision | `approval.HistoryStore` and `GrantStore` expose completed decisions and active grants, not pending requests. |
| Pending approval requests | wired to chat path, needs shared pending surface | Pending approvals currently arrive as `chat.ApprovalRequestMsg` with a response channel. Mission Control must share this path instead of resolving from history. |
| Learning suggestions | wired as event stream, informational today | `LearningSuggestionEvent` reaches the TUI, but current handling only renders a status line. Mission Control may render it as a proposed mission; acceptance is not persistence in Wave 1. |
| Metrics and context summaries | wired | Existing observability snapshots and context panel data can feed the header or detail panel. |
| Run ledger state | needs UI projector/wiring | The store has snapshots and summaries, but cockpit does not currently receive a run ledger dependency. |
| Teammate runtime state | needs UI projector/wiring | `AgentRunStore.List()` exists, but the automation module keeps the store internal. RunLedger mirrors only selected blocked-approval state. |
| Chat transcript events | partial | ChatModel owns transcript entries in memory. EventBus has turn-level and continuity events, not a message-level transcript stream. |

Outputs:

- Current Missions
- Proposed Missions
- Live Decisions
- Activity Timeline
- Header context

### Wave 2: Mission Lifecycle

Promote missions to real domain objects with persistence and lifecycle states such as `proposed`, `active`, `prepared`, `blocked`, `waiting_decision`, `done`, and `cancelled`. Link background tasks, run ledger entries, approvals, and teammate state to mission IDs.

### Wave 3: Proactive Agent Behavior

Allow agents to create proposed missions from observed context. Agents may automatically perform read, analyze, draft, and prepare work. Actions with external effects, filesystem changes, spending, message sending, calendar confirmation, or dangerous command execution must become live decisions.

### Wave 4: Work And Life Operating Loops

Add first-class loops for agenda, open loops, follow-ups, documents, research, communication, coding, and automation. These loops should all express state through missions and decisions.

### Wave 5: Multi-Agent Coworking

Show multi-agent collaboration inside missions. The interface should reveal handoffs, conflicts, review, budget, trust, and blocked states without exposing raw internal plumbing first.

### Wave 6: Surface Split

Clarify the command surfaces:

- `lango`: Mission Control
- `lango chat`: focused conversation mode
- `lango cockpit`: operational diagnostics dashboard

Wave 1 does not require this split, but the design should not make it harder.

## Wave 1 Product Shape

The first implementation changes the default TUI from page-first cockpit to mission-first control surface.

Primary regions:

- Current Missions: active and proposed work units.
- Live Decisions: pending approval requests and their risk/effect summaries.
- Activity Timeline and Composer: meaningful activity feed plus user input.
- Header Context: active agents, policy mode, model or runtime summary, and context health.

The first screen should answer:

- What is Lango moving forward right now?
- What is waiting for the user?
- What did agents recently do?
- What can the user steer next?

## Layout

The default layout should be:

```text
+------------------------------------------------------------------+
| Lango Mission Control       active agents / policy / context      |
+---------------------------------------------+--------------------+
| Current Missions                            | Live Decisions     |
|                                             |                    |
| [active] Weekly planning      next: draft   | [approval] exec    |
| [active] Vendor research      waiting...    | [approval] write   |
| [proposed] Calendar cleanup   prepared      | no more pending    |
|                                             |                    |
+---------------------------------------------+--------------------+
| Activity Timeline / Conversation Composer                         |
| agent activity, user messages, tool summaries, mission updates     |
| >                                                                |
+------------------------------------------------------------------+
```

The existing sidebar becomes secondary navigation. It may be opened with `Ctrl+B`, but it should not dominate the default first impression. Existing pages remain available for detail views: Chat, Tasks, Approvals, Status, Settings, Tools, and Sessions.

## Navigation

Initial keyboard model:

- `Tab`: cycle focus through Missions, Decisions, and Composer.
- `j` / `k` or arrow keys: move within the focused list.
- `Enter`: open selected mission or decision detail.
- `a`: approve the focused decision when applicable.
- `d` or `Esc`: deny or dismiss the focused decision when applicable.
- `/`: open a command palette or quick action prompt.
- `Ctrl+B`: toggle secondary navigation.
- `Ctrl+P`: toggle context or detail panel.

The user should not need to pick a page before understanding the system state.

## Mission Semantics

There are two mission kinds in Wave 1:

- `active`: a mission already accepted by the user or derived from an explicit user request.
- `proposed`: a mission suggested by agents from context or preparation work.

Wave 1 does not persist first-class missions. It creates mission view models from existing sources. A proposed mission can be rendered and selected, but acceptance does not create a durable mission object in Wave 1. If a selected proposed mission needs action, the UI submits an explicit prompt through the composer path or routes to an existing approval/task flow that already exists.

Mission titles must be user-facing work units, not internal IDs. A mission may show raw identifiers only in detail views.

Examples:

- `Prepare weekly planning brief`
- `Research vendor options`
- `Review tomorrow's calendar conflicts`
- `Draft release follow-up`
- `Stabilize Mission Control TUI implementation`

## Mission Derivation Rules

Wave 1 must use explicit derivation rules. It must not assume every source can provide every field.

### Background Task To Mission

| Mission field | Source | Rule |
| --- | --- | --- |
| `ID` | `TaskSnapshot.ID` | Prefix with `bg:` to avoid collisions. |
| `Kind` | task origin | `active`. Background tasks are already accepted work. |
| `Title` | `TaskSnapshot.Prompt` | Use a deterministic first-line summary: trim whitespace, take the first non-empty line, strip common automation prefixes when present, truncate for display. Do not call an LLM in Wave 1. |
| `Status` | `TaskSnapshot.StatusText` | Map directly to `pending`, `running`, `done`, `failed`, or `cancelled`. |
| `NextAction` | task status | `waiting to start`, `running`, `review result`, `retry or inspect error`, or `cancelled`; leave empty if no useful text exists. |
| `OwnerAgent` | `OriginChannel` / task source | Use `automator` for background-originated tasks unless a linked AgentRun/RunLedger source provides a better value. |
| `Risk` | none | Empty in Wave 1. Risk belongs to approvals unless a linked source provides it. |
| `UpdatedAt` | `StartedAt` / `CompletedAt` / `NextRetryAt` | Use the newest non-zero timestamp. |

### Run Ledger To Mission

RunLedger-derived missions are allowed only after cockpit receives a RunLedger reader. The first projector should derive:

| Mission field | Source | Rule |
| --- | --- | --- |
| `ID` | `RunSnapshot.RunID` | Prefix with `run:`. |
| `Kind` | `RunSnapshot.SourceKind` | `active`. |
| `Title` | `RunSnapshot.Goal` then `OriginalRequest` | Prefer `Goal`; fall back to trimmed `OriginalRequest`. |
| `Status` | `RunSnapshot.Status` and blocker fields | Use the snapshot status, with `blocked` when `CurrentBlocker` or teammate approval blocked state is present. |
| `NextAction` | `NextExecutableStep()` | Use the next step goal when available; otherwise leave empty. |
| `OwnerAgent` | current step | Use `Step.OwnerAgent` for the current or next executable step when available. |

### Learning Suggestion To Proposed Mission

`LearningSuggestionEvent` can produce a proposed mission only as a UI proposal:

- `ID`: `learn:` + `SuggestionID`
- `Kind`: `proposed`
- `Title`: "Review learning suggestion" plus a short deterministic summary of `ProposedRule`
- `Status`: `prepared`
- `NextAction`: `accept, dismiss, or inspect`

Accepting this proposed mission in Wave 1 must not imply durable learning persistence unless the existing learning approval path is also implemented. Current code treats learning suggestions as informational in the TUI.

## Decision Semantics

Live Decisions are interrupt-level items that require user direction.

Wave 1 starts with one decision category:

- `approval`: approve, allow for session, or deny a pending approval request.

The following categories are deferred until mission lifecycle producers exist:

- `choice`: requires a producer that can express alternatives and consequences.
- `blocker`: requires mission lifecycle or run-ledger blocker projection.
- `risk`: remains an attribute of approval view models in Wave 1, not a separate decision category.

Every approval decision must include:

- The requested action.
- Why the decision is needed.
- What will happen if approved.
- The risk level or scope of effect.

Decisions should reuse existing approval pipeline behavior where possible. Wave 1 must not bypass established approval tiers or double-confirm behavior.

## Approval Single Path

Mission Control must not create a parallel approval system.

Pending approval source of truth:

```text
tool middleware
  -> approval.Provider.RequestApproval(...)
  -> TUI fallback provider sends chat.ApprovalRequestMsg{Request, ViewModel, Response}
  -> cockpit routes the message to Mission Control and/or Chat
  -> user action writes approval.ApprovalResponse to Response channel
  -> middleware appends HistoryStore and executes or denies the tool
```

`approval.HistoryStore` is a completed-decision log, not a pending-decision queue. The Mission Control page can read it for history, but it must resolve active approvals through the pending `ApprovalRequestMsg.Response` channel or a shared pending approval model that owns that channel.

When the user resolves an approval from Mission Control, the same pending approval must disappear from Chat and Approvals-derived surfaces on the next render. The implementation should keep one pending approval state inside cockpit and render it in multiple places, rather than copying independent pending states.

## Composer Behavior

The Wave 1 composer remains a turn submission surface. Submitting text from Mission Control should run the same `TurnRunner` path as chat mode and should echo the user input into the activity/conversation area in place.

Submitting a prompt in Wave 1 must not implicitly create a durable mission. If the turn starts an existing background or run-ledger flow, that work appears in the mission list through normal projection. General agent-authored proposed missions are deferred until a producer exists; learning suggestions are the only proposed-mission event source required in Wave 1.

## Activity Timeline

The activity timeline is not a raw log, but Wave 1 must keep transformation cheap and deterministic. Do not call an LLM per event. Use terse rule-based text that preserves source, event, and useful payload.

Examples:

- `agent navigator: delegation started - research vendor candidates`
- `approval: exec waiting for user decision`
- `run run-123: step started - generate plan`
- `learning: suggestion prepared - save retry preference`
- `system: context compacted, reclaimed 4.2k tokens`

Tool names, run IDs, and event types may appear in terse Wave 1 timeline entries when they are the only stable source identity. Richer humanized activity text is deferred until a later wave.

## Responsive Behavior

Mission Control must define narrow terminal behavior before implementation:

- Width >= 120: render Missions and Live Decisions side by side, with timeline/composer below.
- Width 80-119: render Missions first, then Live Decisions as a compact strip, then timeline/composer.
- Width < 80: render one focused lane at a time. `Tab` switches lane; the footer shows the active lane and pending decision count.
- Height < 24: keep header, focused lane, and footer. Composer opens as an inline prompt when the user presses `/` or starts typing, rather than permanently consuming rows.
- Live approval decisions outrank mission list content when screen space is scarce.

## Architecture

Wave 1 should add a Mission Control UI projection layer plus missing UI-facing producer wiring. It should not add a new mission domain engine.

Recommended types:

```go
type MissionControlSnapshot struct {
    Header    HeaderView
    Missions  []MissionView
    Decisions []DecisionView
    Timeline  []ActivityView
}

type MissionView struct {
    ID         string
    Kind       MissionKind
    Title      string
    Status     string
    NextAction string
    OwnerAgent string
    Risk       string
    UpdatedAt  time.Time
}

type DecisionView struct {
    ID          string
    Category    DecisionCategory
    Title       string
    Reason      string
    ApproveText string
    DenyText    string
    Risk        string
    UpdatedAt   time.Time
}

type ActivityView struct {
    ID        string
    MissionID string
    Text      string
    Tone      string
    UpdatedAt time.Time
}
```

The UI layer should introduce `MissionControlPage` as a cockpit page/root surface. It should depend on projection interfaces rather than importing concrete application internals directly. Business decisions remain in application services and existing approval/task/run-ledger code.

Wiring shape:

```text
cmd/lango
  -> cockpit.New(Deps{..., MissionControlDeps})
  -> cockpit.Model(activePage = PageMissionControl)
  -> MissionControlPage
       -> MissionControlProjector
            -> BackgroundTaskReader
            -> PendingApprovalReader/Resolver
            -> ApprovalHistoryReader
            -> LearningSuggestionBuffer
            -> MetricsReader
            -> optional RunLedgerReader
            -> optional AgentRunReader
```

`MissionControlPage` should satisfy the existing `cockpit.Page` interface unless replacing the cockpit root is explicitly chosen later. Wave 1 should register it as the default active page and keep Chat as a detail page.

EventBus subscriptions should start and stop with the page or cockpit lifecycle. Subscriptions that feed global pending approvals must remain active even when the user switches pages, otherwise approval prompts can be missed.

## Existing Surface Integration

Wave 1 should reuse and reframe existing surfaces:

- Background tasks become active mission candidates.
- Pending approval requests become live decisions through the shared pending approval path.
- Approval history and grants remain history/detail data.
- Run ledger state may provide progress, blocked state, and owner agent only after a RunLedger reader is wired to cockpit.
- Chat turn events and continuity events feed the timeline. Message-level transcript reuse is local to Chat until a transcript event stream exists.
- Learning suggestions become proposed mission candidates with informational acceptance semantics in Wave 1.
- Metrics and context panel data feed the header or optional context view.

Existing cockpit pages should stay reachable as details. They should not define the default mental model.

## Relationship To Ontology

Mission Control should not create ontology entities in Wave 1. The projection stays separate from the ontology and graph stores.

The long-term path is that Wave 2 mission lifecycle can promote durable missions into ontology-backed concepts:

- Mission maps to a goal-like entity.
- Mission steps map to activity-like entities.
- Results, receipts, and completed deliverables map to outcome-like entities.
- Relationships connect missions to people, channels, documents, projects, tools, and learned preferences.

This prevents Mission Control and the ontology graph from growing into parallel systems. Wave 1 should name fields and IDs so a later ontology bridge can attach stable mission identities without rewriting the TUI model.

## Teammate Role Mapping

The work should follow the repository teammate roles:

- PM: define mission-first acceptance criteria and protect scope.
- Architect: define UI projection boundaries and prevent mission lifecycle creep in Wave 1.
- UI/UX Developer: implement the TUI surface as a thin view over projections.
- Application Developer: only extend existing services if a required projection cannot be produced cleanly.
- QA: validate empty states, narrow terminals, live decisions, proposed missions, and optional blocked-state rendering when a RunLedger or AgentRun reader is wired.
- Technical Writer: update public docs only after verifying actual command wiring.

## Risk Policy

Automatic agent behavior should be visible and bounded:

- Observed: read, search, analyze, classify, and summarize can appear in the timeline without approval.
- Prepared: drafts, plans, candidate missions, and execution proposals can appear as proposed missions.
- Needs Decision: external effects, filesystem mutations, spending, message sending, calendar confirmation, dangerous commands, and broad automation require a live decision.

Wave 1 should display this policy in the UI language but should not invent a separate policy engine.

## Testing Strategy

Wave 1 needs focused tests:

- Projection tests for tasks to active missions.
- Projection tests for pending approval messages to live decisions.
- Projection tests for approval history as history-only data, not pending decisions.
- Projection tests for `LearningSuggestionEvent` to proposed mission views.
- Projection tests for unavailable RunLedger/AgentRun readers.
- Rendering tests for empty state, active missions, proposed missions, decisions, and timeline.
- Narrow terminal rendering tests.
- Keyboard routing tests for focus cycling and decision actions.
- Approval resolution tests proving Mission Control writes to the same pending approval response path as Chat.
- Regression tests that existing Chat, Tasks, Approvals, Status, Settings, Tools, and Sessions remain reachable.

## Documentation Strategy

Documentation updates must be based on actual command wiring. Public docs should describe only implemented behavior. Internal planning remains under `internal-docs/superpowers`.

Expected public docs after implementation:

- Update cockpit feature docs to describe Mission Control as the default `lango` surface.
- Keep `lango chat` documented as focused chat mode.
- Describe current limitations honestly: Wave 1 uses projection over background tasks, pending approvals, learning suggestions, and metrics; RunLedger and AgentRun data appear only after their readers are wired.

## Non-Goals For Wave 1

- Do not build a persistent mission domain engine.
- Do not remove the existing cockpit pages.
- Do not split `lango cockpit` into a separate diagnostics-only surface yet.
- Do not add broad autonomous execution beyond existing approval and task behavior.
- Do not replace the approval pipeline.
- Do not add LLM-based event humanization.
- Do not create durable ontology entities for missions.
- Do not claim RunLedger, AgentRun, or message-level transcript sources are available until their readers are wired.
- Do not document future waves as shipped behavior.

## Acceptance Criteria For Wave 1

- Running `lango` opens a Mission Control first screen.
- Current missions render from background tasks without new mission persistence.
- RunLedger and AgentRun mission fields render only when their readers are wired; otherwise the UI degrades cleanly.
- Live decisions render pending approval requests with action, reason, effect, and risk text.
- Mission Control approval actions resolve through the same pending approval response path as Chat.
- Proposed missions render from `LearningSuggestionEvent`; deterministic fixtures cover empty, single, and many proposed mission states in tests.
- The conversation composer remains available from the default screen and submits through the existing turn path without creating durable missions implicitly.
- Existing cockpit pages remain reachable.
- Risky actions still route through existing approval behavior.
- Empty states guide the user toward starting or reviewing missions.
- Responsive behavior is defined and tested for wide, medium, narrow, and short terminal sizes.
- Tests cover projection, rendering, keyboard focus, approval resolution, and existing page reachability.
- Public docs match the implemented command behavior.
