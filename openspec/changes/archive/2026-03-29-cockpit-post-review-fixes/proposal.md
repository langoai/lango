## Why

Change-1~3 코드 리뷰에서 P1 3개 + P2 1개가 발견됨. 테스트는 통과하지만 런타임에서 사용자 경험을 깨뜨리는 실제 버그: embedded settings save가 auto-enable을 무력화, live config 오염으로 취소해도 런타임 상태 변경, context panel 첫 토글 시 width=0 렌더링, sidebar cursor와 active page 비동기화.

## What Changes

- embedded save 시 `state.Dirty` (카테고리 키) 대신 `config.ContextRelatedKeys()` (dotted path) 전달하여 explicitKeys 의미 복원
- `Config.Clone()` 메서드 추가, `NewEditorForEmbedding`에서 deep copy하여 live config 격리
- `toggleContext()`에서 panel visible 전환 시 contextPanel에 올바른 width 전파
- `SetActive(id)`에서 cursor 동기화하여 visual active와 keyboard cursor 일치

## Capabilities

### New Capabilities
- `config-clone`: Config.Clone() deep copy 메서드 (json roundtrip)

### Modified Capabilities
- `cockpit-settings-page`: embedded save의 explicitKeys가 dotted context-related path로 변경, editor가 config deep copy로 작업
- `cockpit-shell`: toggleContext()에서 context panel에 width 전파 추가
- `cockpit-sidebar`: SetActive()에서 cursor 동기화 추가

## Impact

- **Modified**: `internal/config/types.go` — Clone() 추가
- **Modified**: `internal/cli/settings/editor.go` — explicitKeys 수정 + cfg.Clone() 호출
- **Modified**: `internal/cli/cockpit/cockpit.go` — toggleContext() width 전파
- **Modified**: `internal/cli/cockpit/sidebar/sidebar.go` — SetActive() cursor 동기화
- **Tests**: config/types_test.go, editor_embed_test.go, cockpit_test.go, sidebar_test.go
