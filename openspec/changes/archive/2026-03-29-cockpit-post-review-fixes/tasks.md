## 1. Fix explicitKeys 의미 불일치

- [x] 1.1 `editor.go` handleMenuSelection("save") OnSave 분기에서 `maps.Clone(e.state.Dirty)` → `config.ContextRelatedKeys()` 기반 explicit keys로 변경
- [x] 1.2 unused `maps` import 제거
- [x] 1.3 `editor_embed_test.go`: OnSave에 전달되는 키가 dotted path인지 검증
- [x] 1.4 `editor_embed_test.go`: save 후 ResolveContextAutoEnable이 명시적 키를 존중하는지 roundtrip 검증

## 2. Fix live config 오염

- [x] 2.1 `config/types.go`: `func (c *Config) Clone() *Config` 추가 (json roundtrip, nil guard)
- [x] 2.2 `config/types_test.go`: Clone deep copy 테스트 (값 복사, 포인터 독립)
- [x] 2.3 `config/types_test.go`: nil clone 테스트
- [x] 2.4 `editor.go` `NewEditorForEmbedding`: `cfg.Clone()` 호출
- [x] 2.5 `editor_embed_test.go`: form 수정이 원본 config를 오염시키지 않는 테스트

## 3. Fix context panel 첫 토글 width=0

- [x] 3.1 `cockpit.go` `toggleContext()`: visible 전환 시 contextPanel에 `WindowSizeMsg{Width: cpw}` 전파
- [x] 3.2 `cockpit_test.go`: 초기 hidden(width=0) → 첫 Ctrl+P → panel이 ContextPanelWidth 받는 경로 재현 테스트

## 4. Fix sidebar cursor 비동기화

- [x] 4.1 `sidebar.go` `SetActive(id)`: matching item index로 cursor 동기화
- [x] 4.2 `sidebar_test.go`: SetActive 후 cursor가 해당 index로 이동하는지 검증

## 5. Build Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/cli/settings/... ./internal/cli/cockpit/... ./internal/config/... ./cmd/lango/...` passes
- [x] 5.3 `go vet` passes
