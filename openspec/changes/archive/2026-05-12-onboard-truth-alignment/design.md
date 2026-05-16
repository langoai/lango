## Context

The default interactive entry point is bare `lango`, and the repository already documents that `lango doctor` is the validation path after setup. However, the current onboard completion output only pushes `lango serve`, and several public docs overstate what the 5-step wizard can configure. This creates avoidable confusion exactly at the first-run UX boundary.

## Goals / Non-Goals

**Goals:**
- Make onboard completion output truthfully advertise the real primary next steps.
- Ensure public docs do not describe nonexistent onboard submenu paths for advanced features.
- Preserve the existing five-step wizard scope and avoid feature creep in this slice.

**Non-Goals:**
- Expanding the wizard to support advanced embedding, graph, multi-agent, A2A, or auth configuration.
- Redesigning the workbench, cockpit, or settings editor.
- Changing runtime configuration semantics for any advanced subsystem.

## Decisions

### D1: Fix guidance instead of expanding wizard scope
The fastest truth-alignment win is to make the current five-step onboarding explicit and route advanced setup to `lango settings` or `lango config import/export`. This avoids inflating the wizard while still improving first-run clarity.

### D2: Treat bare `lango` as the primary post-save entry point
Since the root command already launches the workbench in interactive mode, onboard completion should advertise that as the default place to begin using Lango after configuration.

### D3: Add a focused regression for post-save messaging
The onboarding package now gains a stdout-capture regression that asserts the next-step output mentions all four required commands. This keeps future copy edits from drifting away from product reality again.

## Risks / Trade-offs

- **[Docs become terser]** → Some previous wording implied a more guided advanced setup path. Mitigation: point users directly at `lango settings` and config import/export, which are real and already documented.
- **[More commands in post-save output]** → The next-step block is slightly longer. Mitigation: it remains a short flat list and removes ambiguity about where to go next.
