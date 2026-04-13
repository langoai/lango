## 1. OpenSpec + Docs

- [x] 1.1 Add a settings delta spec for provenance configuration editing
- [x] 1.2 Update proposal/design to keep `session_isolation` out of settings scope

## 2. Settings UI

- [x] 2.1 Add `provenance` category to the Automation section menu
- [x] 2.2 Add `NewProvenanceForm(cfg)` with the five config-backed provenance fields
- [x] 2.3 Wire `provenance` category to the form in setup flow/editor mapping

## 3. State Mapping + Tests

- [x] 3.1 Add provenance field mapping in `tuicore` state update logic
- [x] 3.2 Add form/default/update tests for provenance settings
- [x] 3.3 Add menu test ensuring Automation includes Provenance

## 4. Verification

- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run `go test ./...`
- [x] 4.3 Validate and archive the change
