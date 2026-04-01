## 1. EventBus Extension

- [x] 1.1 Add EventAlertTriggered constant and AlertEvent struct to internal/eventbus/events.go

## 2. Ent Schema

- [x] 2.1 Add "alert" to action enum Values list in internal/ent/schema/audit_log.go (do NOT run go generate)

## 3. Config

- [x] 3.1 Add AlertingConfig struct to internal/config/types.go and Alerting field to Config struct
- [x] 3.2 Add Alerting defaults in DefaultConfig() in internal/config/loader.go
- [x] 3.3 Add AlertingConfig defaults test in internal/config/types_defaults_test.go

## 4. Alerting Dispatcher

- [x] 4.1 Create internal/alerting/dispatcher.go with threshold rules, sliding window, deduplication, and AlertEvent publishing
- [x] 4.2 Create internal/alerting/dispatcher_test.go with tests for threshold detection, deduplication, and disabled state

## 5. Audit Recorder

- [x] 5.1 Add handleAlert method and SubscribeTyped[eventbus.AlertEvent] in internal/observability/audit/recorder.go

## 6. App Wiring

- [x] 6.1 Wire alerting dispatcher in internal/app/wiring_observability.go (create dispatcher, subscribe to policy events)
- [x] 6.2 Add /alerts endpoint in internal/app/routes_observability.go (query audit DB for action="alert")

## 7. CLI

- [x] 7.1 Create internal/cli/alerts/alerts.go with NewAlertsCmd(), list and summary subcommands
- [x] 7.2 Add alertsCmd to cmd/lango/main.go with GroupID="sys"

## 8. Documentation

- [x] 8.1 Create docs/cli/alerts.md with CLI reference
- [x] 8.2 Create docs/features/alerting.md with feature documentation

## 9. Verification

- [x] 9.1 Run go build ./... and go test ./internal/alerting/... ./internal/config/... -v
