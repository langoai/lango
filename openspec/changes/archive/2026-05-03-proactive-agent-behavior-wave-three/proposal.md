# Proactive Agent Behavior Wave 3

## Why

Wave 2 made missions durable and made Mission Control read durable mission rows first, but the product is still mostly reactive. Users can see accepted work, yet the system does not clearly surface useful next work that it has already noticed and prepared in a safe, explainable way.

Wave 3 introduces the first proactive slice without weakening user control. The system should notice one concrete class of useful work, prepare a low-risk brief automatically, show that prepared proposal in Mission Control, and let the user accept it into the durable mission lifecycle without losing the prepared context.

## What Changes

This change adds the first practical proactive-behavior slice:

- introduce a transient proposal model separate from durable `Mission` rows
- use `LearningSuggestionEvent` as the only active proposal producer in this slice
- generate a deterministic prepared brief from source-native evidence rather than broad heuristic agent work
- render `suggested`, `preparing`, and `prepared` proposal states in Mission Control
- accept a prepared proposal into a durable mission while preserving the prepared brief context
- keep `proposed` transient; durable mission rows still begin at `prepared` or `active`

## Scope Guardrails

Wave 3 is intentionally narrow:

- no durable `proposed` mission rows before acceptance
- no librarian-gap proposal producer in this slice
- no runtime-failure proposal producer in this slice
- no generic proposal-owned background execution or `RunLedger` execution in this slice
- no LLM-generated preparation work
- no risky preparation behavior such as filesystem mutation, messaging, payments, calendar confirmation, or dangerous command execution

## User Impact

Mission Control becomes more agentic without becoming opaque. Users can see one concrete proactive proposal source, inspect a prepared brief that explains why the proposal exists, and accept that proposal into durable mission truth while keeping the prepared context intact.
