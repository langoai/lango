## 1. Fix explicitKeys semantic mismatch

- [x] 1.1 In `editor.go` handleMenuSelection("save") OnSave branch, change from `maps.Clone(e.state.Dirty)` to explicit keys based on `config.ContextRelatedKeys()`
- [x] 1.2 Remove unused `maps` import
- [x] 1.3 `editor_embed_test.go`: Verify that keys passed to OnSave are dotted paths
- [x] 1.4 `editor_embed_test.go`: Roundtrip verification that ResolveContextAutoEnable respects explicit keys after save

## 2. Fix live config contamination

- [x] 2.1 `config/types.go`: Add `func (c *Config) Clone() *Config` (json roundtrip, nil guard)
- [x] 2.2 `config/types_test.go`: Clone deep copy test (value copy, pointer independence)
- [x] 2.3 `config/types_test.go`: nil clone test
- [x] 2.4 `editor.go` `NewEditorForEmbedding`: Call `cfg.Clone()`
- [x] 2.5 `editor_embed_test.go`: Test that form modifications do not contaminate the original config

## 3. Fix context panel first toggle width=0

- [x] 3.1 `cockpit.go` `toggleContext()`: Propagate `WindowSizeMsg{Width: cpw}` to contextPanel on visibility toggle
- [x] 3.2 `cockpit_test.go`: Reproduce test for initial hidden(width=0) → first Ctrl+P → panel receives ContextPanelWidth

## 4. Fix sidebar cursor desynchronization

- [x] 4.1 `sidebar.go` `SetActive(id)`: Synchronize cursor with matching item index
- [x] 4.2 `sidebar_test.go`: Verify that cursor moves to the corresponding index after SetActive

## 5. Build Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/cli/settings/... ./internal/cli/cockpit/... ./internal/config/... ./cmd/lango/...` passes
- [x] 5.3 `go vet` passes
