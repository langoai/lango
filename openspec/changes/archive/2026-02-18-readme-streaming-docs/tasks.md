## 1. Feature List Update

- [x] 1.1 Change Gateway feature bullet from "WebSocket/HTTP server for control plane" to "WebSocket/HTTP server with real-time streaming"

## 2. Configuration Reference Table

- [x] 2.1 Add `observationalMemory.maxReflectionsInContext` row (int, default `5`, "Max reflections injected into LLM context (0 = unlimited)")
- [x] 2.2 Add `observationalMemory.maxObservationsInContext` row (int, default `20`, "Max observations injected into LLM context (0 = unlimited)")

## 3. Embedding & RAG Section

- [x] 3.1 Add "Embedding Cache" subsection describing in-memory cache with 5-minute TTL and 100-entry limit

## 4. Observational Memory Section

- [x] 4.1 Add "Context Limits" bullet to the component list describing default limits (5 reflections, 20 observations)

## 5. WebSocket Events Section

- [x] 5.1 Add "WebSocket Events" subsection with table documenting `agent.thinking`, `agent.chunk`, `agent.done` events
- [x] 5.2 Add backward compatibility note for clients not handling `agent.chunk`
