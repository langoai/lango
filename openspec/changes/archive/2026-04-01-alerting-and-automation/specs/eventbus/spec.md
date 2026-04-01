## ADDED Requirements

### Requirement: AlertEvent type
The eventbus package SHALL define an AlertEvent struct with fields: Type (string), Severity (string), Message (string), Details (map[string]interface{}), SessionKey (string), and Timestamp (time.Time). The EventName() method SHALL return "alert.triggered".

#### Scenario: AlertEvent implements Event interface
- **WHEN** an AlertEvent is created
- **THEN** calling EventName() returns "alert.triggered"

### Requirement: Alert event name constant
The eventbus package SHALL define an EventAlertTriggered constant with value "alert.triggered".

#### Scenario: Constant matches EventName
- **WHEN** the EventAlertTriggered constant is used
- **THEN** its value equals the return value of AlertEvent.EventName()
