# Surface Split Wave 6 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` for each task. Every task must go through implementation, spec-compliance review, and code-quality review before the next task starts.

**Goal:** Land the first practical slice of Wave 6 by separating the interactive surfaces so bare `lango` launches a standalone mission workbench, `lango chat` stays focused chat, and `lango cockpit` stays the explicit multi-page operator dashboard.

**Architecture:** Wave 6 is a surface-contract change, not a new mission-domain wave. It adds a new `internal/cli/workbench` root shell that reuses the existing Mission Control projection/page stack, shared approval path, proposal path, loop path, and collaboration path without duplicating business logic. The full cockpit shell remains intact and explicit, while root CLI routing and public truth are updated so bare `lango` no longer aliases `lango cockpit`.

First-slice surface contract:

- bare `lango`: standalone mission workbench
- `lango chat`: focused transcript-first chat
- `lango cockpit`: explicit multi-page dashboard

First-slice honesty contract:

- workbench reuses current Mission Control assets even if they still live under `internal/cli/cockpit/...`
- no EventBus unsubscribe support is invented; subscriptions remain program-lifetime
- cockpit page set and current landing page remain unchanged in the first slice
- `--with-channels` remains cockpit-only

**Tech Stack:** Go, Bubble Tea, `internal/cli/workbench`, `internal/cli/cockpit`, `internal/cli/chat`, `cmd/lango`, OpenSpec, Zensical docs

## Scope Guardrails

- do not invent a second mission projection system
- do not redesign plain chat into a mission UI
- do not redesign cockpit pages or diagnostics layout in this slice
- do not claim page-level EventBus unsubscribe behavior
- do not start live channels from bare `lango`
- do not let docs or main specs continue to disagree about bare `lango`

## File Map

### Worker A: OpenSpec / Docs / Public Truth

- Create: `openspec/changes/surface-split-wave-six/proposal.md`
- Create: `openspec/changes/surface-split-wave-six/design.md`
- Create: `openspec/changes/surface-split-wave-six/tasks.md`
- Create: `openspec/changes/surface-split-wave-six/specs/mission-workbench-tui/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/mission-control-tui/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/tui-cockpit-layout/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/cockpit-shell/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/cockpit-pages/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/interactive-tui-chat/spec.md`
- Modify: `README.md`
- Modify: `docs/cli/core.md`
- Modify: `docs/cli/index.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

### Worker B: Workbench Root Shell

- Create: `internal/cli/workbench/model.go`
- Create: `internal/cli/workbench/model_test.go`

### Worker C: Mission Control Reuse / CLI Wiring

- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

## Task Breakdown

### Task 1: Create the Wave 6 OpenSpec Change

**Owner:** Worker A

**Files:**
- Create: `openspec/changes/surface-split-wave-six/proposal.md`
- Create: `openspec/changes/surface-split-wave-six/design.md`
- Create: `openspec/changes/surface-split-wave-six/tasks.md`
- Create: `openspec/changes/surface-split-wave-six/specs/mission-workbench-tui/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/mission-control-tui/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/tui-cockpit-layout/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/cockpit-shell/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/cockpit-pages/spec.md`
- Create: `openspec/changes/surface-split-wave-six/specs/interactive-tui-chat/spec.md`

- [ ] **Step 1: Create the change skeleton**

The change must capture:

- bare `lango` is no longer equivalent to `lango cockpit`
- a new mission workbench root surface exists
- chat remains direct and focused
- cockpit remains explicit and multi-page

- [ ] **Step 2: Add the new workbench spec and delta specs**

The new `mission-workbench-tui` main spec should define:

- bare `lango` launches the standalone mission workbench
- the workbench hosts Mission Control content without the full cockpit sidebar shell
- workbench hints mention both `lango chat` and `lango cockpit`

The delta specs should define:

- `mission-control-tui`: Mission Control behavior remains reusable across workbench and cockpit
- `tui-cockpit-layout`: bare `lango` no longer aliases cockpit
- `cockpit-shell`: cockpit remains explicit rather than root-default
- `cockpit-pages`: page routing remains valid but no longer owns the bare-`lango` contract
- `interactive-tui-chat`: `lango chat` remains direct focused chat and bare `lango` no longer means chat

- [ ] **Step 3: Validate the change**

Run:

```bash
openspec validate surface-split-wave-six --strict
```

Expected:

```text
Change 'surface-split-wave-six' is valid
```

### Task 2: Add the Standalone Workbench Root Shell

**Owner:** Worker B

**Files:**
- Create: `internal/cli/workbench/model.go`
- Create: `internal/cli/workbench/model_test.go`

- [ ] **Step 1: Define the workbench root model**

The root should be a thin Bubble Tea shell that:

- owns one shared `chat.ChatModel`
- mounts one `pages.MissionControlPage`
- delegates `SetProgram(...)` to the shared chat model
- activates Mission Control immediately at startup

- [ ] **Step 2: Reuse existing shared Mission Control support**

The workbench root must own and wire:

- `PendingApprovalRegistry`
- `LearningSuggestionBuffer`
- `MissionActivityBuffer`
- `SubscribeMissionControlEvents(...)`

These subscriptions should be workbench-program lifetime, not page lifetime.

- [ ] **Step 3: Forward runtime and input messages through the existing Mission Control path**

The workbench root should forward:

- `tea.WindowSizeMsg`
- `tea.KeyMsg`
- runtime/chat messages already consumed by `MissionControlPage`
- approval/runtime messages that must still reach the shared chat model path

The workbench root should not grow sidebar/router/context-panel behavior.

- [ ] **Step 4: Add focused workbench tests**

Cover:

- startup activates Mission Control immediately
- program delegation reaches the shared chat model
- loading/empty/root rendering works without cockpit sidebar state
- pending approval/runtime refresh still flows through Mission Control

Run:

```bash
go test ./internal/cli/workbench -count=1
```

Expected:

```text
ok
```

### Task 3: Make Mission Control Copy And Hints Surface-Aware

**Owner:** Worker C

**Files:**
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`

- [ ] **Step 1: Add a small surface-aware option to Mission Control page construction**

The page needs a narrow way to distinguish:

- workbench rendering
- cockpit rendering

This should stay a UI concern only. Do not move mission logic into the page options.

- [ ] **Step 2: Update first-slice copy without changing mission semantics**

Workbench copy should mention:

- `lango chat` for focused chat
- `lango cockpit` for the advanced dashboard

Cockpit copy should stop implying it is the same as bare `lango`.

This task must not change:

- mission-start on plain text
- slash-command passthrough
- approval resolution path
- proposal accept/dismiss behavior

- [ ] **Step 3: Add rendering tests for workbench-vs-cockpit copy**

Cover:

- workbench footer/header mentions `lango cockpit`
- workbench still mentions `lango chat`
- cockpit rendering no longer claims bare `lango` equivalence

Run:

```bash
go test ./internal/cli/cockpit/pages -run 'TestMissionControl.*' -count=1
```

Expected:

```text
ok
```

### Task 4: Route Bare `lango` To The Workbench And Keep Other Surfaces Explicit

**Owner:** Worker C

**Files:**
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

- [ ] **Step 1: Add `runWorkbench(initialMode string)`**

The new path should:

- use local interactive app boot like cockpit/chat
- initialize logging to its own TUI log file
- pre-create the session on valid `--mode`
- use a `workbench-*` session key prefix
- install the TUI approval fallback provider just like chat/cockpit

- [ ] **Step 2: Reuse existing Mission Control dependency assembly**

Bare `lango` should build the same narrow Mission Control read/write dependencies already used by cockpit.

If the existing helper name now becomes misleading, rename it to something surface-neutral rather than keeping cockpit-only naming around shared behavior.

- [ ] **Step 3: Change root CLI routing**

Update the root command so:

- interactive bare `lango` -> `runWorkbench(...)`
- `lango cockpit` -> `runCockpit(...)`
- `lango chat` -> `runChat(...)`

Keep:

- non-interactive bare `lango` help behavior unless implementation proves a stricter change is needed
- `--with-channels` only on `lango cockpit`

- [ ] **Step 4: Add command-routing tests**

Cover:

- root `RunE` selects workbench
- `cockpitCmd()` still selects cockpit
- `chatCmd()` still selects chat
- help text no longer describes bare `lango` as equivalent to cockpit

Run:

```bash
go test ./cmd/lango -count=1
```

Expected:

```text
ok
```

### Task 5: Update Public Docs To Match The New Surface Truth

**Owner:** Worker A

**Files:**
- Modify: `README.md`
- Modify: `docs/cli/core.md`
- Modify: `docs/cli/index.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

- [ ] **Step 1: Update root command descriptions**

Docs must describe:

- bare `lango` as the mission workbench
- `lango chat` as focused chat
- `lango cockpit` as the explicit multi-page dashboard

- [ ] **Step 2: Remove stale equivalence language**

Delete or rewrite all public wording that still says:

- bare `lango` is the same as `lango cockpit`
- cockpit is the default root shell for bare `lango`

- [ ] **Step 3: Keep docs aligned with actual first-slice behavior**

Do not claim:

- a new diagnostics-first cockpit layout if code did not change it
- live channels start from bare `lango`
- chat now creates missions automatically

- [ ] **Step 4: Verify docs build**

Run:

```bash
.venv/bin/zensical build
```

Expected:

```text
Build completed successfully
```

### Task 6: Final Verification, OpenSpec Sync, And Archive

**Owner:** Worker A

**Files:**
- Modify: `openspec/specs/mission-workbench-tui/spec.md` (via archive result)
- Modify: `openspec/specs/mission-control-tui/spec.md` (via archive result)
- Modify: `openspec/specs/tui-cockpit-layout/spec.md` (via archive result)
- Modify: `openspec/specs/cockpit-shell/spec.md` (via archive result)
- Modify: `openspec/specs/cockpit-pages/spec.md` (via archive result)
- Modify: `openspec/specs/interactive-tui-chat/spec.md` (via archive result)
- Modify: `openspec/changes/archive/...` (archive result)

- [ ] **Step 1: Run focused and repo-wide verification**

Run:

```bash
go test ./internal/cli/workbench ./internal/cli/cockpit/pages ./cmd/lango ./internal/archtest -count=1
go build ./...
go test ./...
.venv/bin/zensical build
openspec validate surface-split-wave-six --strict
openspec validate --changes --strict
```

Expected:

```text
ok
ok
Build completed successfully
Change 'surface-split-wave-six' is valid
```

- [ ] **Step 2: Archive the completed change**

Run:

```bash
openspec archive surface-split-wave-six --yes
```

Expected:

```text
Task status: complete
Specs updated and change archived
```

- [ ] **Step 3: Re-run final verification on archived state**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
openspec validate --changes --strict
```

Expected:

```text
ok
ok
Build completed successfully
```

## Success Criteria

The plan is complete when all of the following are true:

- bare `lango` launches a standalone workbench shell instead of the full cockpit shell
- `lango chat` still launches direct focused chat
- `lango cockpit` still launches the explicit multi-page dashboard
- Mission Control behavior is reused rather than duplicated
- public docs and main OpenSpec no longer claim bare `lango` equals `lango cockpit`
- full repo verification and OpenSpec archive succeed on final HEAD
