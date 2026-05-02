# Mission Control Wave 1

## Why

The current default `lango` surface is still chat-first even though the product direction has shifted toward a mission-native operator experience. Users need to see ongoing work, the latest pending approval, and prepared suggestions immediately instead of discovering them only after navigating through cockpit pages or waiting for chat interruptions.

Wave 1 exists to change the first-screen shape without inventing a new mission domain engine. The runtime already has enough honest signals to project a useful Mission Control view, but some UI-facing producers and ownership boundaries need to be tightened first.

## What Changes

This change introduces Mission Control as the default `lango` TUI surface and defines the supporting contracts needed to ship it safely:

- `lango` opens Mission Control by default
- `lango chat` remains the direct chat fallback
- active missions are projected from existing background and optional runtime sources
- the latest pending approval becomes a live decision through a shared pending-response owner
- learning suggestions become proposed missions as UI proposals only
- cockpit keeps shared subscriptions and buffers for mission projections during the TUI session

## Scope Guardrails

Wave 1 is explicitly limited to projection and producer hardening:

- no durable mission persistence
- no new mission lifecycle engine
- no LLM-based activity humanization
- no message-level transcript event stream
- no page-lifetime unsubscribe semantics

## User Impact

The first screen becomes a mission-first control surface that answers what Lango is doing, what needs a decision, and what the user can steer next. Users who want the focused conversation surface still have a direct path through `lango chat`.
