## Context

The automation subsystems (Cron, Background, Workflow) have five bugs discovered through production usage:

1. **Cron: Identical outputs** — Isolated sessions (`cron:{name}:{ms}`) give the LLM no memory of prior runs, leading to deterministic identical responses.
2. **Cron: Remove/Pause doesn't stop running jobs** — `robfig/cron.Remove()` only prevents future triggers; already-dispatched goroutines complete.
3. **Cron: Name-based removal fails** — `cron_remove` tool only accepts UUID; AI agents passing job names hit `uuid.Parse()` errors.
4. **Background: Cancel status overwritten** — `Cancel()` sets `Cancelled` → `runner.Run()` returns error → `Fail()` overwrites to `Failed`.
5. **Workflow: Session contamination + incomplete cancellation** — Session key `workflow:{name}:{stepID}` is shared across re-runs; cancelled workflows continue executing pending steps.

## Goals / Non-Goals

**Goals:**
- Cron jobs produce diverse outputs across executions by enriching prompts with history context
- Removing or pausing a cron job immediately cancels any in-flight execution
- AI agents can refer to cron jobs by name or UUID interchangeably
- Cancelled background tasks remain in `Cancelled` state regardless of runner error
- Workflow re-runs get independent sessions; cancellation prevents new step starts

**Non-Goals:**
- Changing the session mode architecture (isolated vs. main)
- Adding persistent session storage for cron jobs
- Modifying the robfig/cron library itself
- Background task persistence (tasks remain in-memory)

## Decisions

### D1: History injection via prompt enrichment (not session replay)
**Choice**: Prepend recent execution results to the prompt as context.
**Alternative**: Replay previous conversations into the session — requires session store changes and much higher token cost.
**Rationale**: Prompt enrichment is simple, low-cost, and sufficient for diversity. The LLM sees "don't repeat these" with previous outputs. Graceful degradation if history query fails.

### D2: In-flight tracking via `map[string]context.CancelFunc`
**Choice**: Scheduler maintains `inFlight` map of jobID→cancel, registered in `executeWithSemaphore`, called on Remove/Pause/Stop.
**Alternative**: Use a global context per-scheduler — too coarse, would cancel all jobs.
**Rationale**: Per-job cancellation is precise. The map is protected by a dedicated `inFlightMu` mutex to avoid contention with the entries map.

### D3: Name-or-ID resolution at scheduler level
**Choice**: `ResolveJobID(ctx, nameOrID)` method on Scheduler that checks `uuid.Parse()` first, falls back to `store.GetByName()`.
**Alternative**: Resolution at tool handler level — duplicated logic across 3 handlers.
**Rationale**: Single resolution method, reusable. Tool handlers become thinner.

### D4: Terminal state guard pattern for background tasks
**Choice**: `Fail()` and `Complete()` check `if t.Status == Cancelled { return }` inside the lock.
**Alternative**: Check context cancellation in `execute()` only — doesn't cover all race windows.
**Rationale**: Defense in depth. Both the guard in task methods AND the context check in execute() work together to prevent status overwrite.

### D5: RunID in workflow session keys
**Choice**: Change session key from `workflow:{name}:{stepID}` to `workflow:{name}:{runID}:{stepID}`.
**Rationale**: runID is already available in executeStep. This ensures re-runs get independent sessions with zero additional queries.

## Risks / Trade-offs

- **History injection increases prompt size** → Mitigated by limiting to 10 entries, 200 chars each (~2K tokens max). Graceful fallback on query failure.
- **In-flight cancellation may interrupt mid-response generation** → Acceptable: the job was explicitly removed/paused. Better than silent completion after deletion.
- **Name-based lookup adds a DB query on non-UUID inputs** → Negligible cost for human-initiated operations (remove/pause/resume).
- **Workflow session key format change** → Non-breaking: session keys are ephemeral and not stored long-term. Existing running workflows complete with old format.
