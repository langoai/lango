## Context

Change-1~3 코드 리뷰에서 4개 런타임 버그 발견. 모두 테스트가 커버하지 못하는 영역 — shared pointer mutation, 의미 불일치, lifecycle timing.

## Goals / Non-Goals

**Goals:**
- explicitKeys 의미 복원 (embedded save와 standalone save 동일 동작)
- live config 격리 (embedded editor가 런타임 config를 오염시키지 않음)
- context panel 첫 토글에서 올바른 width 렌더링
- sidebar cursor와 active page 동기화

**Non-Goals:**
- narrow terminal min-width clamp (별도 change)
- editor width-aware rendering (Change-3에서 이미 deferred)

## Decisions

### D1: explicitKeys → context-related dotted paths
embedded save에서 `state.Dirty` (카테고리 키)를 넘기는 대신, standalone과 동일하게 `config.ContextRelatedKeys()`를 explicit keys로 전달. auto-enable이 사용자의 명시적 설정을 존중함.

### D2: Config.Clone() via JSON roundtrip
`encoding/json` Marshal/Unmarshal로 deep copy. 성능보다 정확성 우선 (설정 저장은 hot path가 아님). nil guard 포함.

### D3: Clone 위치는 NewEditorForEmbedding
SettingsPage가 아닌 Editor 생성자에서 clone. 이유: Editor가 config 소유권 경계. 모든 embedded 사용처가 자동으로 보호됨.

### D4: toggleContext()에서 panel에 width 전파
`propagateResize()`는 child+pages만 갱신. contextPanel은 page가 아니므로 별도 처리 필요. toggle ON 시 `contextPanel.Update(WindowSizeMsg{Width: cpw})` 호출.

### D5: SetActive()에서 cursor 동기화
SetActive(id)에서 items를 순회하여 matching index를 cursor에 대입. 최소 수정으로 가장 안전.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| JSON roundtrip이 일부 zero-value 필드를 변경할 수 있음 | Config의 json tag가 omitempty를 쓰는 곳은 slice/map뿐. 기본형은 safe. Clone 테스트로 검증. |
| explicitKeys 변경이 standalone save와 미묘하게 다를 수 있음 | 동일한 `config.ContextRelatedKeys()` 소스를 사용하므로 보장됨 |
