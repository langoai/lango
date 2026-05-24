# Mission Control Slice 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land Slice 1 Mission Control as the default `lango` cockpit surface using background tasks, shared pending approvals, learning suggestions, and optional RunLedger/AgentRun readers without introducing durable mission persistence.

**Architecture:** Keep Mission Control as a cockpit-owned UI projection, not a new domain engine. Add cockpit-lifetime shared state for pending approvals and learning suggestions, expose optional producer readers from `app`, derive mission/decision/timeline views deterministically, and integrate a new Mission Control page as the default landing surface while preserving `lango chat` as the direct chat fallback.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, `internal/cli/cockpit`, `internal/cli/chat`, `internal/app`, `internal/background`, `internal/approval`, `internal/runledger`, `internal/agentrt`, OpenSpec, Zensical docs

---

## File Map

### Worker A: OpenSpec / Docs / Public Surface Truth

- Create: `openspec/changes/2026-05-02-mission-control-slice-one/proposal.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/design.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/tasks.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/mission-control-tui/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/cockpit-shell/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/cockpit-pages/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/interactive-tui-chat/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/tui-approval-tiers/spec.md`
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/cli/core.md` only if the command surface description changes

### Worker B: Producer Exposure / Cockpit-Owned Shared State

- Modify: `internal/app/types.go`
  - Expose `AgentRunStore` on `App`.
- Modify: `internal/app/modules.go`
  - Carry the automation module's agent run store in `automationValues`.
- Modify: `internal/app/app.go`
  - Copy `AgentRunStore` from automation values into `App`.
- Modify: `internal/cli/cockpit/deps.go`
  - Add optional `RunLedgerStore` and `AgentRunStore` dependencies.
- Create: `internal/cli/cockpit/pending_approvals.go`
  - Add `PendingApprovalRegistry`.
- Modify: `internal/cli/chat/chat.go`
  - Add optional delegation to a shared pending approval owner when running inside cockpit.
- Modify: `internal/cli/chat/approval.go`
  - Keep TUI approval provider compatible with cockpit-owned pending approval handling.
- Modify: `internal/cli/chat/chat_test.go`
  - Cover shared pending approval ownership path.
- Create: `internal/cli/cockpit/learning_buffer.go`
  - Add `LearningSuggestionBuffer`.
- Create: `internal/cli/cockpit/activity_buffer.go`
  - Add `MissionActivityBuffer`.
- Create: `internal/cli/cockpit/mission_control_subscriptions.go`
  - Register cockpit-lifetime EventBus subscriptions for learning, continuity, and timeline producers that actually arrive on EventBus.
- Test: `internal/cli/cockpit/pending_approvals_test.go`
- Test: `internal/cli/cockpit/learning_buffer_test.go`
- Test: `internal/cli/cockpit/activity_buffer_test.go`
- Test: `internal/app/modules_test.go`
- Test: `internal/app/app_test.go`

### Worker C: Mission Control Projection / Page / Cockpit Shell

- Modify: `internal/cli/cockpit/router.go`
  - Add `PageMissionControl`, update sidebar order, and update routing helpers.
- Modify: `internal/cli/cockpit/cockpit.go`
  - Default `activePage` to `PageMissionControl`, wire the new page, route shared pending approvals, and append shell-level timeline events.
- Modify: `internal/cli/cockpit/cockpit_test.go`
  - Cover page switching, default page, focus, and approval routing.
- Modify: `internal/cli/cockpit/router_test.go`
  - Cover the extra page ID, menu count, and round-trip routing.
- Modify: `internal/cli/chat/input.go`
  - Expose or reuse the existing composer behavior rather than cloning it in Mission Control.
- Modify: `internal/cli/chat/commands.go`
  - Keep slash-command behavior consistent when Mission Control uses the chat composer path.
- Create: `internal/cli/cockpit/missioncontrol_types.go`
  - Define `MissionKind`, `MissionStatus`, `DecisionCategory`, and view structs including `HeaderView`.
- Create: `internal/cli/cockpit/missioncontrol_projector.go`
  - Build the deterministic projector from background tasks, shared pending approvals, learning suggestions, timeline buffer, and optional readers.
- Create: `internal/cli/cockpit/missioncontrol_projector_test.go`
  - Cover derivation, header context, ordering, overflow, and degraded states.
- Create: `internal/cli/cockpit/pages/missioncontrol.go`
  - Implement the page view, header context, focus model, responsive rendering, composer handoff, and loading/degraded states.
- Create: `internal/cli/cockpit/pages/missioncontrol_test.go`
  - Cover layout, typing intent, composer submit/echo behavior, timeline retention, responsive behavior, and discoverability text.
- Modify: `cmd/lango/main.go`
  - Pass the new deps and register Mission Control as the default page.

## Task Breakdown

### Task 1: Create the OpenSpec Change for Slice 1

**Owner:** Worker A

**Files:**
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/proposal.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/design.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/tasks.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/mission-control-tui/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/cockpit-shell/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/cockpit-pages/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/interactive-tui-chat/spec.md`
- Create: `openspec/changes/2026-05-02-mission-control-slice-one/specs/tui-approval-tiers/spec.md`

- [ ] **Step 1: Create the new OpenSpec change skeleton**

Create the Slice 1 change directory and add proposal/design/tasks docs that mirror the approved internal design:

- Mission Control is the default `lango` surface
- `lango chat` remains the direct chat fallback
- Slice 1 is projection plus producer hardening, not a mission domain engine
- approvals use the shared pending response path
- learning suggestions become proposed missions only as UI proposals

- [ ] **Step 2: Add the new `mission-control-tui` delta spec**

Define the new capability in:

- `openspec/changes/2026-05-02-mission-control-slice-one/specs/mission-control-tui/spec.md`

Minimum scenarios:

- default `lango` enters Mission Control
- missions derive from background tasks
- pending approvals render as live decisions
- learning suggestions render as proposed missions
- loading, empty, degraded, and narrow-terminal states

- [ ] **Step 3: Add the supporting delta specs inside the change**

Add change-local delta specs so they describe the new default shape rather than mutating main specs before code lands:

- `openspec/changes/2026-05-02-mission-control-slice-one/specs/cockpit-shell/spec.md`
- `openspec/changes/2026-05-02-mission-control-slice-one/specs/cockpit-pages/spec.md`
- `openspec/changes/2026-05-02-mission-control-slice-one/specs/interactive-tui-chat/spec.md`
- `openspec/changes/2026-05-02-mission-control-slice-one/specs/tui-approval-tiers/spec.md`

Focus on:

- default page migration
- shared pending approval owner
- minimal shortcut blast radius
- Mission Control page routing

- [ ] **Step 4: Sanity-check the OpenSpec slice before code**

Confirm the change does not claim any Slice 2 behavior:

- no durable mission persistence
- no LLM-based activity humanization
- no message-level transcript stream
- no page-lifetime unsubscribe semantics

### Task 2: Expose the Optional Producer Readers from `app`

**Owner:** Worker B

**Files:**
- Modify: `internal/app/types.go`
- Modify: `internal/app/modules.go`
- Modify: `internal/app/app.go`
- Modify: `internal/cli/cockpit/deps.go`
- Test: `internal/app/modules_test.go`
- Test: `internal/app/app_test.go`

- [ ] **Step 1: Add `AgentRunStore` to `App`**

Expose an optional `agentrt.AgentRunStore` on `App` next to `BackgroundManager` and `RunLedgerStore`. Keep it read-only from the cockpit consumer point of view.

- [ ] **Step 2: Carry the automation module's agent run store through initialization**

Update `automationValues` in `internal/app/modules.go` so the agent run store created in the automation module is retained and available during `App` assembly.

- [ ] **Step 3: Copy the optional agent run store into `App`**

In `internal/app/app.go`, wire `AgentRunStore` from automation values into the `App` struct the same way `BackgroundManager` and `WorkflowEngine` are copied today.

- [ ] **Step 4: Extend cockpit deps**

Add optional fields to `internal/cli/cockpit/deps.go`:

- `RunLedgerStore runledger.RunLedgerStore`
- `AgentRunStore agentrt.AgentRunStore`

Do not add approval provider plumbing here.

- [ ] **Step 5: Add or extend app wiring tests**

Cover:

- `RunLedgerStore` already still reaches `App`
- `AgentRunStore` is present when automation is enabled
- nil automation keeps the field nil

Run:

```bash
go test ./internal/app -run 'Test.*Automation.*AgentRun|Test.*RunLedger' -count=1
go test ./internal/app -run 'Test.*PopulateAppFields.*(AgentRun|Automation)' -count=1
```

Expected:

```text
ok
```

### Task 3: Add Cockpit-Lifetime Shared State for Pending Approvals and Learning Suggestions

**Owner:** Worker B

**Files:**
- Create: `internal/cli/cockpit/pending_approvals.go`
- Create: `internal/cli/cockpit/pending_approvals_test.go`
- Modify: `internal/cli/chat/chat.go`
- Modify: `internal/cli/chat/approval.go`
- Modify: `internal/cli/chat/chat_test.go`
- Create: `internal/cli/cockpit/learning_buffer.go`
- Create: `internal/cli/cockpit/learning_buffer_test.go`
- Create: `internal/cli/cockpit/activity_buffer.go`
- Create: `internal/cli/cockpit/activity_buffer_test.go`
- Create: `internal/cli/cockpit/mission_control_subscriptions.go`
- Modify: `cmd/lango/main.go`

- [ ] **Step 1: Implement `PendingApprovalRegistry`**

Add a cockpit-owned registry that:

- stores the latest pending `chat.ApprovalRequestMsg`
- exposes `Latest()`, `HasPending()`, and `Resolve(id, approval.ApprovalResponse) bool`
- writes exactly once to the original response channel on successful resolve
- clears the pending request after resolution

- [ ] **Step 2: Make cockpit the pending approval owner**

Move ownership explicitly:

- `cockpit.handleApprovalRequest(...)` registers the pending request instead of treating `ChatModel` as the owner
- `ChatModel` gains an optional shared-pending path for cockpit mode
- the old unconditional `switchPage(PageChat)` behavior is removed or narrowed so Mission Control can remain visible while approval is pending

This step must leave standalone `lango chat` approval behavior unchanged.

- [ ] **Step 3: Implement `LearningSuggestionBuffer`**

Add a ring buffer with:

- capacity 20
- 30 minute TTL
- explicit dismiss by suggestion ID
- mutex protection

The buffer stores `eventbus.LearningSuggestionEvent` items for Mission Control projection.

- [ ] **Step 4: Implement `MissionActivityBuffer`**

Add a cockpit-owned activity buffer that:

- stores the most recent 200 activity items
- supports append, snapshot, and reset
- keeps item ordering stable
- is safe for concurrent cockpit-lifetime event handlers

- [ ] **Step 5: Add cockpit-lifetime subscriptions**

Create a helper that subscribes at cockpit startup time and updates shared state for:

- learning suggestion events
- compaction completed / slow events
- other timeline-relevant EventBus events that already exist in Slice 1

Do not invent `Unsubscribe()`. The owner is the cockpit session.

- [ ] **Step 6: Append shell-level runtime events into `MissionActivityBuffer`**

From the cockpit shell path, append deterministic activity items for runtime messages that are not sourced from EventBus:

- `chat.ChannelMessageMsg`
- `chat.DelegationMsg`
- `chat.BudgetWarningMsg`
- `chat.RecoveryMsg`
- user composer submissions
- turn completion summaries when they are surfaced in Mission Control

- [ ] **Step 7: Wire the new shared state at startup**

In `cmd/lango/main.go`, create the registry and buffer once, inject them into cockpit deps, and make sure they live for the same duration as the TUI program.

- [ ] **Step 8: Add focused registry/buffer tests**

Cover:

- latest approval replacement
- double resolve guard
- cockpit owns the pending approval even when Chat is mounted
- learning suggestion TTL
- learning suggestion dismiss
- activity buffer retention
- continuity event append
- shell-level runtime event append

Run:

```bash
go test ./internal/cli/cockpit -run 'Test(PendingApproval|LearningSuggestion|MissionActivity)' -count=1
go test ./internal/cli/chat -run 'Test.*Approval.*Shared|Test.*Cockpit.*Approval' -count=1
```

Expected:

```text
ok
```

### Task 4: Build the Mission Control Projector

**Owner:** Worker C

**Files:**
- Create: `internal/cli/cockpit/missioncontrol_types.go`
- Create: `internal/cli/cockpit/missioncontrol_projector.go`
- Create: `internal/cli/cockpit/missioncontrol_projector_test.go`

- [ ] **Step 1: Define the enums and view structs**

Add:

- `MissionKind`
- `MissionStatus`
- `DecisionCategory`
- `MissionControlSnapshot`
- `MissionView`
- `DecisionView`
- `ActivityView`
- `HeaderView`

Keep the status space limited to:

- `Unknown`
- `Pending`
- `Running`
- `Blocked`
- `Done`
- `Failed`
- `Cancelled`

- [ ] **Step 2: Implement background-task derivation**

Project background tasks into active missions using the approved deterministic mapping:

- `bg:` ID prefix
- first non-empty prompt line as title
- status mapping from `TaskSnapshot.StatusText`
- deterministic `NextAction`
- `UpdatedAt` from newest relevant timestamp

- [ ] **Step 3: Implement pending approval derivation**

Project the shared pending approval into a single live decision using:

- request ID
- tool name plus short summary
- `ApprovalViewModel.RuleExplanation`
- `ApprovalViewModel.Risk.Level` and `Risk.Label`
- `ApprovalRequest.Summary` as effect text
- fixed `Approve`, `Deny`, `Allow for session` labels

- [ ] **Step 4: Implement learning-suggestion derivation**

Project buffered learning suggestions into proposed missions with:

- `learn:` prefix
- result-oriented title beginning with `Apply learning rule`
- prepared status
- deterministic ordering by recency

- [ ] **Step 5: Implement optional RunLedger and AgentRun readers**

Support nil readers first, then add derivation when readers are present:

- `RunLedgerStore` contributes optional mission fields and blocked state
- `AgentRunStore` contributes owner agent and runtime condition hints

If either reader is nil, the projector must emit a degraded note rather than fake data.

- [ ] **Step 6: Implement ordering and overflow logic**

Mission ordering:

- active before proposed
- status priority `running > blocked > pending > done > failed > cancelled`
- `UpdatedAt` descending

- [ ] **Step 7: Implement `HeaderView` derivation**

Derive:

- active agent summary from optional `AgentRunStore`
- model/provider summary from cockpit config
- pending decision count from `PendingApprovalRegistry`
- degraded note when optional readers are absent
- metrics/context summary from existing observability inputs when available

- [ ] **Step 8: Add projector tests**

Cover:

- background task title derivation
- approval-to-decision derivation
- learning-suggestion-to-proposed-mission derivation
- nil reader degraded state
- header context derivation
- mission ordering
- overflow summaries

Run:

```bash
go test ./internal/cli/cockpit -run 'TestMissionControl(Projector|Ordering|Degraded|Overflow|Header)' -count=1
```

Expected:

```text
ok
```

### Task 5: Implement the Mission Control Page

**Owner:** Worker C

**Files:**
- Create: `internal/cli/cockpit/pages/missioncontrol.go`
- Create: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Modify: `internal/cli/chat/input.go`
- Modify: `internal/cli/chat/commands.go`
- Modify: `internal/cli/chat/chat.go`

- [ ] **Step 1: Create the page skeleton**

Implement a new cockpit page with:

- `Title() string`
- `ShortHelp() []key.Binding`
- `Activate() tea.Cmd`
- `Deactivate()`
- `Update()`
- `View()`

Keep the page thin. It renders projector output and local focus state only.

- [ ] **Step 2: Reuse the existing chat composer path instead of cloning it**

Choose one concrete route and implement it consistently:

- either extend the chat child surface with a reusable composer/submit API that Mission Control calls
- or extract the composer + submit/interrupt/slash path from `ChatModel` into a shared controller used by both Chat and Mission Control

Do not create a second independent submit path.

- [ ] **Step 3: Implement the three-lane focus model**

Support:

- Missions lane
- Decisions lane
- Composer lane

`Tab` cycles focus between these lanes. Printable character input forces focus to Composer before applying the keystroke.

- [ ] **Step 4: Implement the split activity/composer surface**

Render:

- a scrollable activity timeline region
- a fixed one-line composer prompt at the bottom

Do not merge this into the chat transcript renderer. The page must:

- render `HeaderView`
- render timeline items from `MissionActivityBuffer`
- submit through the same turn path as chat mode
- echo the user's submitted text into the activity area in place

- [ ] **Step 5: Implement responsive behavior**

Behavior by size:

- width >= 120: side-by-side missions and decisions
- width 80-119: stacked missions then compact decisions
- width < 80: one focused lane at a time
- height < 24: compact header + focused lane + footer, composer opened inline

- [ ] **Step 6: Implement loading / empty / degraded views**

Add distinct render paths for:

- first tick before snapshot
- no missions and no pending decisions
- optional reader degradation

- [ ] **Step 7: Add page tests**

Cover:

- first-screen chat fallback copy
- header context rendering
- footer discoverability hint
- focus cycling
- printable-key handoff to composer
- `/` behavior with empty and non-empty composer
- composer submit reuses the existing turn path
- composer submit echoes user text into the activity area
- loading / empty / degraded rendering
- narrow and short terminal layouts

Run:

```bash
go test ./internal/cli/cockpit/pages -run 'TestMissionControl' -count=1
```

Expected:

```text
ok
```

### Task 6: Integrate Mission Control into the Cockpit Shell

**Owner:** Worker C

**Files:**
- Modify: `internal/cli/cockpit/router.go`
- Modify: `internal/cli/cockpit/cockpit.go`
- Modify: `internal/cli/cockpit/cockpit_test.go`
- Modify: `internal/cli/cockpit/router_test.go`
- Modify: `cmd/lango/main.go`

- [ ] **Step 1: Add `PageMissionControl` to routing**

Update:

- `PageID` enum
- `String()`
- `PageIDFromString()`
- `AllPageMetas()`

Keep shortcut blast radius minimal:

- preserve the existing global shortcut meanings unless a test-backed migration is clearly necessary
- make Mission Control reachable by default startup and sidebar placement
- keep `lango chat` as the direct fallback command

- [ ] **Step 2: Make Mission Control the default page**

Set cockpit startup `activePage` to `PageMissionControl`, not `PageChat`.

Keep:

- `lango chat` as the direct chat entry point
- approval routing visible even when Mission Control is active

- [ ] **Step 3: Register the page and pass the new deps**

In `cmd/lango/main.go`:

- construct the page
- inject shared pending approvals, learning buffer, metrics, and optional readers
- keep existing pages reachable

- [ ] **Step 4: Update cockpit tests**

Add or adjust coverage for:

- default page is Mission Control
- unregistered pages remain safe
- approval routing still reaches the shared pending approval owner
- Chat, Tasks, Approvals, Status, Settings, Tools, and Sessions remain reachable after the new page is inserted

Run:

```bash
go test ./internal/cli/cockpit -count=1
```

Expected:

```text
ok
```

### Task 7: Update Public Docs Only After the Surface Is Real

**Owner:** Worker A

**Files:**
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/cli/core.md` only if command wording changes

- [ ] **Step 1: Audit the landed command surface**

Verify before writing docs:

- `lango` opens Mission Control
- `lango chat` still opens direct chat
- page shortcuts and sidebar order
- approval and proposed mission wording
- first-screen migration hint and chat fallback copy

- [ ] **Step 2: Update public docs to match actual behavior**

Document:

- Mission Control as the default cockpit surface
- chat fallback command
- current Slice 1 limits:
  - no durable mission persistence
  - no LLM event humanization
  - optional RunLedger / AgentRun enrichment

- [ ] **Step 3: Build docs**

Run:

```bash
.venv/bin/zensical build
```

Expected:

```text
Build finished
```

### Task 8: Verification, Review Gates, and OpenSpec Completion

**Owner:** Main agent

- [ ] **Step 1: Run focused verification before full suite**

Run:

```bash
go test ./internal/app -run 'Test.*Automation.*AgentRun|Test.*RunLedger' -count=1
go test ./internal/cli/cockpit -run 'Test(PendingApproval|LearningSuggestion|MissionControl(Projector|Ordering|Degraded|Overflow))' -count=1
go test ./internal/cli/cockpit/pages -run 'TestMissionControl' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 2: Run repository-required full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
```

Expected:

```text
ok
Build finished
```

- [ ] **Step 3: Run subagent review before moving to the next slice**

Review targets:

- layer boundaries
- pending approval single-path correctness
- timeline source wiring and retention
- header context truthfulness
- Mission Control page integration
- docs/spec truth alignment

Do not start Slice 2 until the review is addressed and accepted.

- [ ] **Step 4: Complete the OpenSpec workflow**

After code is green and reviewed:

- mark the change tasks complete
- run verify against the landed code
- sync any accepted delta specs to main specs
- archive the change

- [ ] **Step 5: Commit in small slices during implementation**

Recommended commit sequence:

- `spec: add mission control slice one OpenSpec change`
- `feat: expose cockpit mission control producer readers`
- `feat: add cockpit pending approval registry and learning buffer`
- `feat: add mission control projector`
- `feat: add mission control page and default cockpit routing`
- `docs: update mission control public docs`

## Review Gates

### Gate 1: Before Code

- OpenSpec change exists
- internal design and implementation plan agree on scope
- user confirms the plan

### Gate 2: After Shared State and Reader Wiring

- pending approval path is still single-owner
- app wiring remains backward-compatible
- no page-lifetime unsubscribe assumptions remain

### Gate 3: After Mission Control UI Lands

- Mission Control is default
- `lango chat` still behaves as direct chat
- narrow-terminal rendering is usable
- loading / empty / degraded states are distinct

### Gate 4: Before Next Slice

- subagent review is complete
- full verification is green
- OpenSpec verify/sync/archive is done
