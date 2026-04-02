## ADDED Requirements

### Requirement: External alert delivery channels
The alerting system MUST support external delivery channels via a `DeliveryChannel` interface with `Send(ctx, AlertEvent) error` and `Type() string` methods. A `WebhookDelivery` implementation MUST POST alert events as JSON to a configured URL.

#### Scenario: Webhook delivery on alert
- **WHEN** an `AlertEvent` is published with severity >= channel's `minSeverity`
- **THEN** the `DeliveryRouter` SHALL asynchronously POST the alert as JSON to the webhook URL

#### Scenario: Severity filtering
- **WHEN** an alert with severity "warning" is published
- **AND** the channel's `minSeverity` is "critical"
- **THEN** the channel SHALL NOT receive the alert

#### Scenario: Async dispatch
- **WHEN** a webhook endpoint is slow or unreachable
- **THEN** the EventBus publisher SHALL NOT be blocked

### Requirement: Alert delivery configuration
`AlertingConfig` MUST include a `Delivery []AlertDeliveryConfig` field with `type`, `webhookURL`, and `minSeverity` fields.

#### Scenario: Webhook channel configured
- **WHEN** `alerting.delivery` contains `{type: "webhook", webhookURL: "https://...", minSeverity: "warning"}`
- **THEN** the `DeliveryRouter` SHALL register a `WebhookDelivery` channel

#### Scenario: Unknown channel type skipped
- **WHEN** `alerting.delivery` contains `{type: "unknown"}`
- **THEN** the router SHALL log a warning and skip the entry
