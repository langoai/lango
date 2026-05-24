## Context

The default interactive entry point is now capable of separating incomplete profiles from ready ones. The next UX gap is the ready-profile empty state: it avoids setup noise, but it still leaves the operator to invent the first useful request alone.

## Goals / Non-Goals

**Goals:**
- Give ready-profile operators concrete first prompts they can use immediately.
- Preserve the incomplete-profile setup guidance path unchanged.
- Keep the change lightweight and localized to the workbench-flavored Mission Control empty state.

**Non-Goals:**
- Adding clickable UI widgets or interactive menus.
- Changing the composer behavior or command system.
- Expanding starter prompts into dynamic recommendations.

## Decisions

### D1: Use static high-signal starter prompts
Three prompts cover common first-run intents without requiring repository introspection or additional services: summarize the repository, explain the current project structure, and review recent changes.

### D2: Gate starter prompts behind the ready-profile heuristic
Starter prompts only help when the operator can actually use the system. The existing ready-profile heuristic is reused so the workbench continues to prioritize setup guidance when configuration is incomplete.

## Risks / Trade-offs

- **[Static prompts are generic]** → They are not personalized per repository. Mitigation: they are still concrete and immediately actionable, which is enough for the first-success path.
- **[More empty-state copy]** → The empty state grows slightly. Mitigation: starter prompts are only shown for ready profiles and replace ambiguity with directed action.
