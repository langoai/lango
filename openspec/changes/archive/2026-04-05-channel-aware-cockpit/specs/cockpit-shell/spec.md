## MODIFIED Requirements

### Requirement: Message routing to chat child
The cockpit shell SHALL forward `ChannelMessageMsg` to the chat child model regardless of which page is currently active, ensuring channel traffic is never lost when the operator browses non-chat pages.

#### Scenario: Channel message arrives on Settings page
- **WHEN** the active page is PageSettings and a `ChannelMessageMsg` arrives
- **THEN** the message is forwarded to the chat child (not the active page)

#### Scenario: Channel message arrives on Chat page
- **WHEN** the active page is PageChat and a `ChannelMessageMsg` arrives
- **THEN** the message is forwarded to the chat child via normal `forwardToActive`
