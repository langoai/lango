## MODIFIED Requirements

### Requirement: Gateway chat message session key injection
The `handleChatMessage` RPC handler SHALL inject the computed session key into the context via `session.WithSessionKey` before passing it to `agent.RunAndCollect`. This ensures downstream components (approval providers, learning engine) can route by session key.

#### Scenario: Session key available in agent context
- **WHEN** a `chat.message` RPC request is processed with session key "default"
- **THEN** `session.WithSessionKey(ctx, "default")` SHALL be called before `agent.RunAndCollect`
- **AND** `session.SessionKeyFromContext` within the agent pipeline SHALL return "default"

#### Scenario: Authenticated session key propagated
- **WHEN** an authenticated client sends a `chat.message` RPC request
- **THEN** the client's authenticated session key SHALL be injected into the context
- **AND** downstream approval routing SHALL use that session key
