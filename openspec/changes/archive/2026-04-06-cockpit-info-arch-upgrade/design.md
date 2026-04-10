## Context

After Phase 1-4 completion, state became concentrated in cockpit TUI hub files. ChatModel (764 lines, 23 fields), cockpit.go (453 lines, 8+ inline intercepts), 7 duplicate utilities across 6 page files. Hub file modification is unavoidable when adding new features.

## Goals / Non-Goals

**Goals:**
- Extract CPR/pending/approval state from ChatModel into independent types for separation of concerns
- Consolidate duplicate utility functions into a shared package
- Move sidebar item definitions to a centralized meta table
- Improve cockpit.go Update() readability (inline → named methods)
- Remove package-level globals

**Non-Goals:**
- Adding new features (pure refactoring)
- Moving transcript ownership (chatview.go is currently appropriate)
- Introducing renderer registry / surface registry (excessive for current scale)
- Adding shared state between pages

## Decisions

### D1: Sub-models remain as internal composite types within ChatModel
- cprFilter, pendingIndicator, approvalState are all unexported structs
- Composed as ChatModel fields (no interface needed)
- **Why**: Used only within the same package, interface extraction is premature abstraction

### D2: cprFilter.Flush() returns []tea.KeyMsg (not []tea.Cmd)
- Key replay is ChatModel's responsibility (handleKey, input.Update calls needed)
- **Why**: Separates cprFilter from depending on ChatModel

### D3: approvalState integrates dialog scroll/split state
- Move existing package globals (`dialogScrollOffset`, `dialogSplitMode`) to struct fields
- Add scrollOffset, splitMode parameters to renderApprovalDialog signature
- **Why**: Remove package globals, clarify state ownership

### D4: Two RelativeTime variants maintained (behavior preservation)
- `tui.RelativeTime(now, t)` — precise ("5s ago"), used in approvals
- `tui.RelativeTimeHuman(now, t)` — friendly ("just now"), used in sessions
- **Why**: Preserve existing UX, adhere to pure refactoring principle

### D5: Sidebar uses AllPageMetas() central table + full display
- Change to `sidebar.New(items)` parameter-based approach
- AllPageMetas() returns 7 items in the same order as current hardcoded order
- RegisterPage() does not interfere with sidebar (only stores page map)
- **Why**: Prevent duplicate registration, prevent conditional omission, 100% visible behavior preservation

### D6: cockpit.go Update() method extraction within same file
- Extract 8 handler methods within cockpit.go (no separate file needed)
- **Why**: Method extraction, not file splitting — maintains cohesion

## Risks / Trade-offs

- [Risk] Merge conflicts from cascading chat.go modifications across waves → Mitigation: Units within a wave designed to avoid file overlap, waves executed sequentially
- [Risk] approvalState extraction requires simultaneous test file modifications → Mitigation: approval_dialog_test.go and chat_test.go both modified in the same unit
- [Risk] sidebar.New() signature change affects external call sites → Mitigation: sidebar_test.go and cockpit_test.go both modified in the same unit
- [Risk] RelativeTimeHuman separation causes future confusion → Mitigation: Function name clearly conveys intent, difference documented in format_test.go
