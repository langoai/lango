## Context

The current TUI already uses a typed transcript model, turn-state strip, and inline approval interrupts, but two problems remain. First, assistant markdown rendering still uses glamour auto-style selection, which internally queries terminal background color through OSC 11 and leaks that response into the composer on some terminals. Second, the transcript blocks are technically separated but still visually too flat for a coding-agent product.

This change should stay narrow: stabilize terminal interaction and raise transcript readability without reopening the larger cockpit layout decisions.

## Goals / Non-Goals

**Goals:**
- Stop OSC 11 terminal response leakage at the source.
- Protect idle composer input from stray terminal response sequences.
- Improve visual separation between user, assistant, status, and approval transcript blocks.

**Non-Goals:**
- No multi-panel or diagnostics UI.
- No dynamic dark/light theme auto-detection in this phase.
- No changes to slash command surface or runtime behavior outside TUI rendering/input handling.

## Decisions

### 1. Remove markdown auto-style detection

Switch the TUI markdown renderer from `glamour.WithAutoStyle()` to an explicit standard style. Use Glamour dark style as the fixed default so the renderer never triggers terminal background color probing.

Alternative considered:
- Keep auto-style and filter OSC 11 responses only
Why not:
- It keeps the root cause alive and still depends on terminal query behavior.

### 2. Expand composer input guard to OSC

Keep the current CPR-oriented state machine but extend it to recognize OSC responses beginning with `Esc ]` and terminating with BEL or ST (`Esc \`). Apply this guard only while the idle or failed composer path accepts text input.

Alternative considered:
- Global terminal response filtering in all TUI states
Why not:
- Higher risk of interfering with approval and non-composer interactions.

### 3. Use restrained transcript block styling

Keep single-column block rendering, but add clearer distinction through accent bars, subtle background tinting, and stronger assistant labeling. Avoid heavy card borders so the transcript stays dense.

Alternative considered:
- Full card UI for user and assistant messages
Why not:
- Too visually heavy for the current single-column cockpit.

## Risks / Trade-offs

- **[Risk] Fixed markdown style may not match every terminal theme perfectly** → Accept in this phase; correctness and leak prevention are higher priority than adaptive theming.
- **[Risk] OSC filtering could mis-handle unusual `Esc ]` sequences** → Scope filtering to idle composer input and replay unmatched sequences in order.
- **[Risk] Stronger visual separation could make transcript too noisy** → Use subtle tint/accent rather than full bordered cards.

## Migration Plan

1. Update OpenSpec artifacts for the OSC leak and transcript styling pass.
2. Replace markdown auto-style with an explicit Glamour style.
3. Extend the composer guard to handle OSC responses safely.
4. Refine transcript render helpers for clearer user/assistant separation.
5. Update tests, docs, and verify with full build/test.

Rollback is limited to the TUI package and related docs.

## Open Questions

None for this phase. Theme configurability and richer visual customization are deferred.
