# Tasks: Add Event Bus Package

## Implementation

- [x] Create `internal/eventbus/bus.go` with Bus struct, New, Subscribe, Publish, SubscribeTyped
- [x] Create `internal/eventbus/events.go` with ContentSavedEvent, TriplesExtractedEvent, TurnCompletedEvent, ReputationChangedEvent, MemoryGraphEvent, Triple
- [x] Create `internal/eventbus/bus_test.go` with comprehensive test suite

## Verification

- [x] `go build ./internal/eventbus/...` passes
- [x] `go test -v -race ./internal/eventbus/...` passes (10/10 tests)
- [x] `go build ./...` passes (full project build)
- [x] No modifications to existing files
