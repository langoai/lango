## MODIFIED Requirements

### Requirement: Discord message handler context propagation
The Discord `Channel` SHALL propagate the `Start(ctx)` context to message handler callbacks. The `Channel` struct SHALL store the context passed to `Start(ctx)` and use it when invoking the message handler in `onMessageCreate`, instead of using `context.Background()`.

#### Scenario: Start context propagated to handler
- **WHEN** `Channel.Start(ctx)` is called with a context containing cancellation or deadline
- **AND** a message is received via `onMessageCreate`
- **THEN** the handler SHALL be invoked with the stored `Start` context (not `context.Background()`)

#### Scenario: Context carries session key downstream
- **WHEN** a Discord message triggers `onMessageCreate`
- **AND** the handler injects a session key into the propagated context
- **THEN** downstream approval providers SHALL be able to extract the session key via `session.SessionKeyFromContext`
