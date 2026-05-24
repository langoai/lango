# Proposal: Chat Setup Readiness Guard

## Summary

Make focused `lango chat` share the same first-run readiness contract as the mission workbench. When the active profile cannot start a normal agent turn, plain chat should show setup-required guidance and refuse normal turn submission instead of advertising "Ready" and trying a doomed turn.

## Motivation

`config.EvaluateAgentSetup` already centralizes whether the current agent/provider/model/API-key path is usable. The mission workbench uses that contract to show setup guidance and block starter prompts for `config.DefaultConfig()`, but plain chat still renders `default · auto`, shows `Ready`, and attempts to submit user input.

That mismatch makes the focused chat fallback less production-ready than the default workbench surface. A first-run user can enter `lango chat`, see a ready state, submit text, and only then discover the profile is incomplete through a lower-level provider failure.

## Scope

- Add a chat-level readiness snapshot derived from `config.EvaluateAgentSetup`.
- Render setup-required copy in the focused chat header/turn strip/help affordances when setup is incomplete.
- Block normal non-slash turn submission while setup is incomplete and append actionable guidance mentioning `lango onboard`, `lango settings`, and `lango doctor`.
- Preserve slash commands so `/help`, `/status`, `/model`, `/mode`, and related inspection commands remain available before setup is complete.
- Preserve normal turn submission for ready remote providers and local Ollama-style providers.
- Update public CLI docs and main OpenSpec specs after implementation.

## Non-Goals

- Redesign the full chat shell layout.
- Change mission workbench readiness behavior.
- Change provider validation rules in `config.EvaluateAgentSetup`.
- Implement automatic provider setup from chat input.
