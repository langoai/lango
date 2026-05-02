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

### Wave 1: Mission Control Projection

Use existing runtime data to create a Mission Control surface. Do not introduce a new mission domain engine yet.

Inputs:

- Background tasks
- Approval requests and approval history
- Run ledger state
- Teammate runtime state
- Chat transcript events
- Learning or agent suggestions
- Metrics and context summaries

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
- Live Decisions: urgent approvals, choices, risk gates, and user direction prompts.
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
| [active] Weekly planning      next: draft   | [approval] send... |
| [active] Vendor research      waiting...    | [choice] pick path |
| [proposed] Calendar cleanup   prepared      | [risk] file write  |
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

Wave 1 does not persist first-class missions. It creates mission view models from existing sources. A proposed mission can be rendered and accepted, but acceptance may initially translate to an existing background task, chat prompt, run ledger action, or approval flow.

Mission titles must be user-facing work units, not internal IDs. A mission may show raw identifiers only in detail views.

Examples:

- `Prepare weekly planning brief`
- `Research vendor options`
- `Review tomorrow's calendar conflicts`
- `Draft release follow-up`
- `Stabilize Mission Control TUI implementation`

## Decision Semantics

Live Decisions are interrupt-level items that require user direction.

Decision categories:

- `approval`: approve or deny a requested action.
- `choice`: choose among agent-proposed paths.
- `risk`: confirm a higher-risk action.
- `blocker`: unblock a mission with missing information.

Every decision must include:

- The requested action.
- Why the decision is needed.
- What will happen if approved.
- The risk level or scope of effect.

Decisions should reuse existing approval pipeline behavior where possible. Wave 1 must not bypass established approval tiers or double-confirm behavior.

## Activity Timeline

The activity timeline is not a raw log. It should convert runtime events into concise work-language entries.

Examples:

- `planner prepared a weekly planning outline`
- `operator is waiting for approval to send a message`
- `research found 3 vendor candidates`
- `reviewer marked mission blocked: missing budget range`
- `system compacted context and reclaimed 4.2k tokens`

Tool names, run IDs, and event types may appear in detail views, not as the primary timeline language.

## Architecture

Wave 1 should add a Mission Control UI projection layer, not a new application engine.

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

The UI layer may introduce `MissionControlPage`. It should depend on projection interfaces rather than importing concrete application internals directly. Business decisions remain in application services and existing approval/task/run-ledger code.

## Existing Surface Integration

Wave 1 should reuse and reframe existing surfaces:

- Background tasks become active mission candidates.
- Approval requests become live decisions.
- Run ledger and teammate state provide progress, blocked state, and owner agent.
- Chat transcript events feed the timeline and composer.
- Learning suggestions or agent suggestions become proposed mission candidates.
- Metrics and context panel data feed the header or optional context view.

Existing cockpit pages should stay reachable as details. They should not define the default mental model.

## Teammate Role Mapping

The work should follow the repository teammate roles:

- PM: define mission-first acceptance criteria and protect scope.
- Architect: define UI projection boundaries and prevent mission lifecycle creep in Wave 1.
- UI/UX Developer: implement the TUI surface as a thin view over projections.
- Application Developer: only extend existing services if a required projection cannot be produced cleanly.
- QA: validate empty states, narrow terminals, live decisions, proposed missions, and blocked missions.
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
- Projection tests for approvals to live decisions.
- Projection tests for proposed mission fixtures.
- Rendering tests for empty state, active missions, proposed missions, decisions, and timeline.
- Narrow terminal rendering tests.
- Keyboard routing tests for focus cycling and decision actions.
- Regression tests that existing Chat, Tasks, Approvals, Status, Settings, Tools, and Sessions remain reachable.

## Documentation Strategy

Documentation updates must be based on actual command wiring. Public docs should describe only implemented behavior. Internal planning remains under `internal-docs/superpowers`.

Expected public docs after implementation:

- Update cockpit feature docs to describe Mission Control as the default `lango` surface.
- Keep `lango chat` documented as focused chat mode.
- Describe current limitations honestly: Wave 1 uses projection over existing tasks, approvals, run ledger, and teammate state.

## Non-Goals For Wave 1

- Do not build a persistent mission domain engine.
- Do not remove the existing cockpit pages.
- Do not split `lango cockpit` into a separate diagnostics-only surface yet.
- Do not add broad autonomous execution beyond existing approval and task behavior.
- Do not replace the approval pipeline.
- Do not document future waves as shipped behavior.

## Acceptance Criteria For Wave 1

- Running `lango` opens a Mission Control first screen.
- Current missions are rendered from existing runtime sources.
- Live decisions render pending approval requests and user-choice prompts with action, reason, effect, and risk text.
- Proposed missions can be displayed from deterministic test fixtures in Wave 1, and from existing suggestion events when such events are already wired.
- The conversation composer remains available from the default screen.
- Existing cockpit pages remain reachable.
- Risky actions still route through existing approval behavior.
- Empty states guide the user toward starting or reviewing missions.
- Tests cover projection, rendering, keyboard focus, and existing page reachability.
- Public docs match the implemented command behavior.
