## Why

Lango's current context management relies on a hard 100-message cap and a message-count-based sliding window. When messages are trimmed, all context (user intent, decisions, progress) is permanently lost. Long conversations degrade in quality as irrelevant old messages consume token budget while critical early context disappears. Observational Memory (OM) solves this by compressing older conversation history into structured observation notes, preserving essential context while dramatically reducing token usage.

## What Changes

- Add a new `observational-memory` module that observes conversation history and produces compressed observation notes via background LLM calls
- Add a `reflector` that further condenses observations when they accumulate beyond a token threshold
- Add token counting infrastructure to enable token-budget-based context management instead of message-count-based
- Integrate observations and reflections into the context assembly pipeline (system prompt augmentation alongside existing Knowledge RAG)
- Replace the hard 100-message cap with dynamic token-budget-based history truncation
- Add Ent schemas for `Observation` and `Reflection` entities tied to sessions
- Add configuration options for OM thresholds, observer model selection, and async buffering

## Capabilities

### New Capabilities
- `observational-memory`: Core OM system — Observer agent that monitors conversation history and generates compressed observation notes, Reflector agent that condenses accumulated observations, token-budget-based context assembly, async buffering via goroutines
- `token-counter`: Token counting infrastructure — approximate token counting for messages, observations, and reflections across multiple LLM providers

### Modified Capabilities
- `context-retriever`: Add observation/reflection injection into the context assembly pipeline (new context layers for observations and reflections)
- `ent-session-store`: Add Observation and Reflection entity schemas, extend session queries to include OM data
- `adk-architecture`: Replace hard 100-message cap in EventsAdapter with token-budget-based dynamic truncation

## Impact

- **New packages**: `internal/memory/` (observer, reflector, token counter, buffer)
- **Schema changes**: New `Observation` and `Reflection` Ent schemas with session edges
- **Modified packages**:
  - `internal/adk/` — context_model.go (OM injection), state.go (token-based truncation)
  - `internal/knowledge/` — retriever.go (new layers), types.go (layer constants)
  - `internal/session/` — ent_store.go (OM data access)
  - `internal/app/` — app.go (OM goroutine lifecycle)
  - `internal/config/` — types.go (OM config section)
- **New dependencies**: Token counting library (e.g., tiktoken-go or character-based approximation)
- **Config**: New `observationalMemory` section in lango.json
- **LLM cost**: Additional LLM calls for observation/reflection generation (mitigated by using low-cost models like Gemini Flash or local Ollama)
