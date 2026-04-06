## Context

Phase 1-4 완료 후 cockpit TUI hub 파일에 상태가 집중되었다. ChatModel(764줄, 23 fields), cockpit.go(453줄, 8+ inline intercept), 6개 페이지 파일에 7개 중복 유틸리티. 새 기능 추가 시 hub 파일 수정이 불가피한 구조.

## Goals / Non-Goals

**Goals:**
- ChatModel에서 CPR/pending/approval 상태를 독립 타입으로 추출하여 관심사 분리
- 중복 유틸리티 함수를 공유 패키지로 통합
- Sidebar 아이템 정의를 중앙 메타 테이블로 이동
- cockpit.go Update() 가독성 개선 (인라인 → 명명된 메서드)
- Package-level globals 제거

**Non-Goals:**
- 새 기능 추가 (순수 리팩토링)
- transcript ownership 이동 (현재 chatview.go가 적절)
- renderer registry / surface registry 도입 (현재 규모에 과도)
- 페이지 간 공유 state 추가

## Decisions

### D1: Sub-model은 ChatModel 내부 합성 타입으로 유지
- cprFilter, pendingIndicator, approvalState 모두 unexported struct
- ChatModel의 field로 합성 (interface 불필요)
- **Why**: 같은 package 내에서만 사용, interface 추출은 premature abstraction

### D2: cprFilter.Flush()는 []tea.KeyMsg 반환 (not []tea.Cmd)
- 키 replay는 ChatModel 책임 (handleKey, input.Update 호출 필요)
- **Why**: cprFilter가 ChatModel에 의존하지 않도록 분리

### D3: approvalState에 dialog scroll/split 상태 통합
- 기존 package globals (`dialogScrollOffset`, `dialogSplitMode`)를 struct 필드로 이동
- renderApprovalDialog 시그니처에 scrollOffset, splitMode 파라미터 추가
- **Why**: package globals 제거, 상태의 소유권 명확화

### D4: RelativeTime 두 변형 유지 (동작 보존)
- `tui.RelativeTime(now, t)` — 정확 ("5s ago"), approvals에서 사용
- `tui.RelativeTimeHuman(now, t)` — 친근 ("just now"), sessions에서 사용
- **Why**: 기존 UX 보존, 순수 리팩토링 원칙 준수

### D5: Sidebar는 AllPageMetas() 중앙 테이블 + 전체 표시
- sidebar.New(items) 파라미터 방식으로 변경
- AllPageMetas()가 현재 하드코딩 순서와 동일한 7개 아이템 반환
- RegisterPage()는 sidebar 불간섭 (page map 저장만)
- **Why**: 중복 등록 방지, 조건부 누락 방지, visible 동작 100% 보존

### D6: cockpit.go Update() 메서드 추출은 같은 파일 내
- 8개 handler method를 cockpit.go 내부에 추출 (별도 파일 불필요)
- **Why**: 메서드 추출이지 파일 분할이 아님, 응집도 유지

## Risks / Trade-offs

- [Risk] Wave 간 chat.go 연쇄 수정 시 merge 충돌 → Mitigation: Wave 내 unit은 파일 겹침 없도록 설계, Wave 간 순차 실행
- [Risk] approvalState 추출 시 테스트 파일도 함께 수정 필요 → Mitigation: approval_dialog_test.go, chat_test.go 모두 같은 unit에서 수정
- [Risk] sidebar.New() 시그니처 변경이 외부 호출부에 영향 → Mitigation: sidebar_test.go, cockpit_test.go 모두 같은 unit에서 수정
- [Risk] RelativeTimeHuman 분리가 향후 혼란 유발 → Mitigation: 함수명이 의도 명확히 전달, format_test.go에서 차이점 문서화
