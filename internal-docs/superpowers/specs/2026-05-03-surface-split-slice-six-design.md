# Surface Split Slice 6 Design

## Strategic Goal

By the end of Slice 5, Lango has one strong mission-native control surface, but it still exposes that surface through two CLI commands that behave the same:

- `lango`
- `lango cockpit`

That ambiguity now hurts product clarity.

Slice 6 should finish the surface story:

- `lango` becomes the default mission workbench
- `lango chat` remains the focused conversation surface
- `lango cockpit` becomes the explicit operator dashboard

This is not a feature-add slice first. It is a surface-contract slice.

## Product Intent

The user should be able to infer the right entry point from the command name alone.

The three commands should answer three different intents:

- `lango`: "show me what Lango is moving forward and let me steer it"
- `lango chat`: "I want direct focused conversation"
- `lango cockpit`: "show me the broader operator dashboard and diagnostics"

Today the first and third answers collapse into the same TUI. Slice 6 removes that collapse.

## Fixed Decisions

The following decisions are fixed for Slice 6:

- bare `lango` will no longer be equivalent to `lango cockpit`
- `lango` will launch a standalone mission workbench surface rather than the full multi-page cockpit shell
- `lango chat` will remain the direct focused-chat fallback and will not become mission-aware by default
- `lango cockpit` will remain the explicit advanced dashboard surface
- Slice 6 will reuse existing Mission Control projection, lifecycle, loop, proposal, and collaboration assets rather than introducing a new mission domain
- the first Slice 6 slice may keep Mission Control available inside `lango cockpit`, but it must no longer be the bare-`lango` contract

## Why This Needs Its Own Slice

Slice 1 intentionally postponed surface split because the first problem was to build Mission Control itself. That was the right sequence.

After Slices 2 through 5, the product is now meaningfully different:

- durable mission lifecycle exists
- proposal preparation exists
- operating loops exist
- collaboration context exists

At this point, keeping `lango` and `lango cockpit` behaviorally identical hides the product's actual center of gravity.

## Current Constraint Audit

Current code and docs still define:

- root `lango` -> `runCockpit(...)`
- `lango cockpit` -> `runCockpit(...)`
- `lango chat` -> `runChat(...)`

Current docs and specs also explicitly describe bare `lango` as equivalent to `lango cockpit`.

That means Slice 6 must change:

- CLI routing
- TUI root-shell composition
- public command documentation
- OpenSpec truth for `mission-workbench-tui`, `mission-control-tui`, `tui-cockpit-layout`, `cockpit-shell`, `cockpit-pages`, and `interactive-tui-chat`

Slice 6 must also close the existing main-spec contradiction around bare `lango`. The current main OpenSpec set does not fully agree on whether bare `lango` means cockpit or direct chat. This slice should repair that truth as part of the split.

## Recommended Shape

Slice 6 should introduce a new standalone root shell:

- `internal/cli/workbench`

This shell should host the existing Mission Control experience without the full sidebar/context-panel cockpit frame.

Recommended command behavior:

- `lango` -> `runWorkbench(...)`
- `lango chat` -> `runChat(...)`
- `lango cockpit` -> `runCockpit(...)`

This is the cleanest split that preserves existing mission work while making the explicit dashboard remain available for advanced operator use.

## Surface Contracts

### 1. `lango`: Mission Workbench

Bare `lango` should become the default mission-native surface.

It should present:

- Mission Control header
- mission list
- live decision lane
- agenda / operating loops
- activity feed
- inline shared composer

It should not expose the full cockpit sidebar or context-panel chrome by default.

Primary user promise:

> "Open Lango and immediately see the work, decisions, loops, and collaboration that matter now."

The workbench should still show:

- pending approvals through the shared approval path
- proposal acceptance / dismissal
- direct mission-start path from composer
- collaboration summaries on mission rows
- agenda/loops

### 2. `lango chat`: Focused Chat

`lango chat` should remain intentionally simpler:

- transcript-first
- no mission board
- no agenda panel
- no operator dashboard framing

It should keep the current direct turn path and approval handling model.

Slice 6 should not make plain chat silently create mission rows.

### 3. `lango cockpit`: Operator Dashboard

`lango cockpit` should become the explicit multi-page operator/dashboard surface.

The first slice should keep the existing cockpit implementation but reposition its contract:

- advanced navigation
- detailed status/metrics pages
- sessions/tasks/dead-letter/approvals surfaces
- explicit operator use

Mission Control may remain accessible as one cockpit page, but the product contract should no longer say cockpit and bare `lango` are the same thing.

`--with-channels` should remain cockpit-only in the first slice. Bare `lango` should stay the lighter local interactive surface rather than becoming the implicit place where live channel adapters may start.

## Cockpit Positioning In The First Slice

Slice 6 does not need a whole new diagnostics UI before the split becomes valuable.

Therefore the recommended first slice is:

- keep current cockpit page set
- keep Mission Control available inside cockpit
- keep the current cockpit startup page in the first slice to avoid unnecessary churn

However, the **minimum required** distinction is:

- bare `lango` is a standalone workbench
- `lango cockpit` is the multi-page shell

Diagnostics-first positioning in the first slice comes from explicit command/shell separation, page availability, and docs/help copy rather than from a forced page-order reset.

## Architecture Shape

Slice 6 should add a separate root shell instead of forking the Mission Control page itself.

Recommended structure:

- `internal/cli/workbench/model.go`
- `internal/cli/workbench/model_test.go`
- optional `internal/cli/workbench/theme.go` only if needed

The workbench root should:

- own the shared `chat.ChatModel`
- own the same shared pending approval registry used by Mission Control
- own learning/activity buffers and mission-control event subscriptions
- mount one `pages.MissionControlPage` directly as the primary body
- forward runtime/chat/runtime-event messages to that page and child chat model
- expose its own compact footer/help rather than the sidebar/context-panel shell

This preserves a thin UI layer and reuses the existing projection stack.

First-slice honesty:

- the reusable Mission Control page and projector may remain under the existing `internal/cli/cockpit/...` namespace in Slice 6
- product surface split is the priority; UI package-namespace cleanup is secondary unless it becomes necessary for implementation safety

## Subscription Lifetime

The existing `EventBus` utility does not provide unsubscribe support. Slice 6 should therefore keep Mission Control-related subscriptions aligned to program lifetime rather than pretending page-level unsubscribe exists.

That means:

- workbench-owned Mission Control subscriptions live for the lifetime of the workbench program
- cockpit-owned Mission Control subscriptions continue to live for the lifetime of the cockpit program
- Slice 6 should not claim page-level subscribe/unsubscribe semantics that the current bus cannot support

## Session Identity

Workbench should become its own first-class interactive shell, so it should have an explicit session identity rather than silently reusing cockpit naming.

Recommended session prefixes:

- `workbench-*` for bare `lango`
- `tui-*` for `lango chat`
- `cockpit-*` for `lango cockpit`

Slice 6 must verify that no mission, approval, attribution, or continuity behavior incorrectly depends on the old bare-`lango` cockpit prefix.

## Why A New Root Shell Is Better Than More Cockpit Modes

An alternative would be to keep one cockpit root and add many feature flags:

- hide sidebar
- hide context panel
- lock page set
- swap command help

That would technically work, but it would keep the product model muddied:

- the code would still say "everything is cockpit"
- the CLI would still describe distinct surfaces that are only skin-deep variants
- tests and docs would have to reason about implicit cockpit modes instead of explicit surfaces

The new-root-shell approach is cleaner:

- workbench is the mission-native surface
- cockpit is the dashboard shell
- chat is the focused conversation shell

## Reuse Boundaries

Slice 6 should explicitly reuse:

- `chat.ChatModel`
- `PendingApprovalRegistry`
- `LearningSuggestionBuffer`
- `MissionActivityBuffer`
- `SubscribeMissionControlEvents(...)`
- `MissionControlProjector`
- `pages.MissionControlPage`
- existing mission/proposal/loop/collaboration readers from `app.App`

Slice 6 should avoid:

- duplicating mission projection logic
- inventing a second approval path
- inventing a second proposal/loop/collaboration projector

## Minimal CLI Refactor

The CLI layer should move toward three explicit interactive entry points:

- `runWorkbench(initialMode string) error`
- `runChat(initialMode string) error`
- `runCockpit(initialMode string) error`

Shared bootstrapping should be refactored only as far as needed to avoid obvious duplication:

- local interactive app start
- session creation / mode pre-seeding
- log initialization
- shared approval provider TUI fallback

Slice 6 does not require a generalized interactive-surface framework if that abstraction is not already clearly justified by the code.

## Approval And Composer Semantics

Workbench should preserve the existing Mission Control semantics:

- mission-start on plain user text
- slash commands go straight through the shared chat path
- pending approvals resolve through the shared approval response path

Chat should preserve current semantics:

- plain chat stays plain chat
- no implicit mission row creation on every chat turn

Cockpit should preserve current semantics unless the surface split itself requires local adjustments.

## Session And Mode Semantics

Slice 6 should keep `--mode` behavior aligned across all interactive surfaces:

- `lango`
- `lango chat`
- `lango cockpit`

If a valid `--mode` is supplied, the interactive session should be pre-created with that mode just as current chat/cockpit already do.

The split should not create surface-specific mode drift.

## Default Migration

Current users are used to:

- `lango` -> full cockpit

Slice 6 changes that default, so the first slice should include direct in-product discoverability:

- workbench footer or header should mention `lango cockpit` for the advanced dashboard
- workbench copy should still mention `lango chat` for focused chat
- cockpit copy/help should no longer claim it is the same as bare `lango`

No environment-flag rollback is required for the first slice.

The explicit fallback commands are sufficient:

- `lango chat`
- `lango cockpit`

## OpenSpec Impact

Slice 6 should update at least these contracts:

- `mission-workbench-tui` (new main spec)
- `mission-control-tui`
- `tui-cockpit-layout`
- `cockpit-shell`
- `cockpit-pages`
- `interactive-tui-chat`

Expected delta themes:

- bare `lango` no longer implies cockpit shell
- the new `mission-workbench-tui` spec becomes the bare-`lango` contract
- `mission-control-tui` narrows toward reusable Mission Control behavior shared by workbench and cockpit
- cockpit remains explicit and advanced
- chat remains direct and focused

The new main spec is recommended rather than optional. Slice 6 is introducing a first-class named root surface, so leaving that contract only as a Mission Control delta would keep product truth muddy.

## Documentation Impact

Slice 6 must update all public truth that currently says bare `lango` equals `lango cockpit`, including:

- `README.md`
- `docs/cli/core.md`
- `docs/cli/index.md`
- `docs/features/cockpit.md`
- `docs/architecture/project-structure.md`

The docs must reflect actual wiring, not aspirational wording.

## Testing Strategy

Slice 6 should add focused tests for:

- root command routing (`lango` -> workbench, `lango cockpit` -> cockpit, `lango chat` -> chat)
- workbench approval handling through the shared pending path
- workbench mission-start path through composer
- workbench loading / empty / degraded / compact-layout behavior
- cockpit still reachable and still mounts its page set
- docs/help text drift tests where they already exist

Repo-wide verification must still include:

- `go build ./...`
- `go test ./...`
- `.venv/bin/zensical build`
- `openspec validate --changes --strict`

## Non-Goals

Slice 6 first slice should not:

- redesign the plain chat surface into a mission UI
- build a new fully custom diagnostics cockpit from scratch
- add a fourth interactive surface
- introduce a second mission projection system
- introduce mission persistence changes beyond the routing needed for the workbench shell
- fold external P2P team UX into this split

## Recommended Implementation Order

1. Create the Slice 6 OpenSpec change
2. Add standalone `workbench` root shell using existing Mission Control assets
3. Route bare `lango` to the workbench shell
4. Keep `lango cockpit` explicit and update help/contracts/docs
5. Verify all three interactive surfaces and archive the change

## Success Criteria

Slice 6 is successful when:

- bare `lango` no longer launches the same shell as `lango cockpit`
- `lango` feels mission-native on first paint
- `lango cockpit` remains available for explicit dashboard use
- `lango chat` remains a real direct fallback
- public docs and OpenSpec truth no longer describe `lango` and `lango cockpit` as equivalent
