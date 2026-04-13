## Purpose

Capability spec for cockpit-channel-status. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: ChannelTracker aggregates channel status
A `ChannelTracker` SHALL subscribe to EventBus channel events and maintain per-channel connection status, message counts, and last activity timestamps.

#### Scenario: Message count incremented on received event
- **WHEN** `ChannelMessageReceivedEvent` for "telegram" is published
- **THEN** the tracker's telegram entry messageCount increments by 1

#### Scenario: SeedChannel registers connection status
- **WHEN** `SeedChannel("telegram", true)` is called after successful channel start
- **THEN** the tracker's telegram entry has Connected=true

#### Scenario: SeedChannel distinguishes start failure
- **WHEN** `SeedChannel("discord", false)` is called after failed channel start
- **THEN** the tracker's discord entry has Connected=false and MessageCount=0

#### Scenario: Snapshot sorted by name
- **WHEN** channels telegram, discord, slack are seeded
- **THEN** `Snapshot()` returns entries sorted alphabetically: discord, slack, telegram

### Requirement: EventBus channel event bridge
`SubscribeChannelEvents` SHALL subscribe to `ChannelMessageReceivedEvent` on the EventBus and forward each as a `ChannelMessageMsg` tea.Msg to the TUI program.

#### Scenario: Event forwarded to TUI
- **WHEN** `ChannelMessageReceivedEvent` is published with Channel="telegram", SenderName="bob"
- **THEN** the TUI program receives a `ChannelMessageMsg` with matching Channel and SenderName

#### Scenario: Nil bus is safe
- **WHEN** `SubscribeChannelEvents` is called with bus=nil
- **THEN** no panic occurs and no subscriptions are created
