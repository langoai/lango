## Context

The embedded settings editor renders its own context-sensitive help footer in the view layer: the menu advertises navigation/search/tier toggles, forms advertise field navigation and editing controls, and provider/auth/MCP lists advertise list-management keys. The cockpit help bar intentionally stays empty for this page.

## Goals / Non-Goals

**Goals:**

- Document the existing Settings page key surface accurately.
- Make it explicit that discoverability lives inside the embedded editor view rather than in the cockpit help bar.

**Non-Goals:**

- Change Settings runtime behavior.
- Document every single editor sub-mode exhaustively.

## Decisions

- Describe the key surface at a high level by mode: menu, forms, and list screens.
  - Rationale: this mirrors the actual editor structure without turning the feature doc into a full settings-manual duplicate.

- Mention the empty cockpit help bar only in contrast to the embedded inline footer.
  - Rationale: the important operator takeaway is where to look for controls.

## Risks / Trade-offs

- [Docs become slightly more detailed] → Keep the description short and anchored to the actual inline footer behavior already present in code.
